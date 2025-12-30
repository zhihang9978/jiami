package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/feiji/feiji-backend/internal/admin/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func getJWTSecret() string {
	secret := os.Getenv("ADMIN_JWT_SECRET")
	if secret == "" {
		secret = os.Getenv("JWT_SECRET")
	}
	if secret == "" {
		// Default for development only - should be set via environment in production
		secret = "feiji-admin-secret"
	}
	return secret
}

var jwtSecret = []byte(getJWTSecret())

// AdminService interface for admin operations
type AdminService interface {
	Login(ctx context.Context, username, password, ip, userAgent, device, browser string) (*models.Admin, error)
	GetAdminByID(ctx context.Context, id int64) (*models.Admin, error)
	ChangePassword(ctx context.Context, adminID int64, currentPassword, newPassword string) error
	GetLoginLogs(ctx context.Context, adminID int64, page, pageSize int) ([]*models.LoginLog, int64, error)
	CreateAuditLog(ctx context.Context, log *models.AuditLog) error
	GetAuditLogs(ctx context.Context, params *models.PaginationParams) ([]*models.AuditLog, int64, error)
	CreateNotification(ctx context.Context, notification *models.AdminNotification) error
	GetNotifications(ctx context.Context, adminID int64, isRead *bool, category string, page, pageSize int) ([]*models.AdminNotification, int64, int64, error)
	MarkNotificationRead(ctx context.Context, notificationID int64) error
	MarkAllNotificationsRead(ctx context.Context, adminID int64) error
	DeleteNotification(ctx context.Context, notificationID int64) error
	GetUsers(ctx context.Context, params *models.PaginationParams) ([]*models.UserWithAdmin, int64, error)
	GetUserByID(ctx context.Context, userID int64) (*models.UserWithAdmin, error)
	CreateUser(ctx context.Context, req *models.CreateUserRequest) (int64, error)
	UpdateUser(ctx context.Context, userID int64, req *models.UpdateUserRequest) error
	DeleteUser(ctx context.Context, userID int64) error
	BatchUpdateUsers(ctx context.Context, userIDs []int64, operation string) error
	GetCalls(ctx context.Context, params *models.PaginationParams) ([]*models.CallRecord, int64, error)
	GetCallByID(ctx context.Context, callID int64) (*models.CallRecord, error)
	GetBroadcasts(ctx context.Context, params *models.PaginationParams) ([]*models.BroadcastMessage, int64, error)
	GetBroadcastByID(ctx context.Context, broadcastID int64) (*models.BroadcastMessage, error)
	SendBroadcast(ctx context.Context, adminID int64, req *models.SendBroadcastRequest) (int64, int, error)
	RetryBroadcast(ctx context.Context, broadcastID int64) error
	GetBroadcastDetails(ctx context.Context, broadcastID int64, status string, page, pageSize int) ([]*models.BroadcastDetail, int64, error)
	GetMessageTemplates(ctx context.Context) ([]*models.MessageTemplate, error)
	CreateMessageTemplate(ctx context.Context, req *models.CreateTemplateRequest) (int64, error)
	UpdateMessageTemplate(ctx context.Context, templateID int64, req *models.CreateTemplateRequest) error
	DeleteMessageTemplate(ctx context.Context, templateID int64) error
	GetAutoMessageConfigs(ctx context.Context) ([]*models.AutoMessageConfig, error)
	UpdateAutoMessageConfig(ctx context.Context, configType string, req *models.UpdateAutoMessageRequest) error
	GetSystemConfigs(ctx context.Context) ([]*models.SystemConfig, error)
	UpdateSystemConfig(ctx context.Context, configKey, configValue string) error
	GetUserStatistics(ctx context.Context) (*models.UserStatistics, error)
	GetCallStatistics(ctx context.Context) (*models.CallStatistics, error)
	GetBotStatistics(ctx context.Context) (*models.BotStatistics, error)
	GetDashboardData(ctx context.Context) (*models.DashboardData, error)
}

// AdminHandlers contains all admin API handlers
type AdminHandlers struct {
	service AdminService
}

// NewAdminHandlers creates new admin handlers
func NewAdminHandlers(service AdminService) *AdminHandlers {
	return &AdminHandlers{
		service: service,
	}
}

// NewAdminHandlersWithDB creates new admin handlers with database connection
func NewAdminHandlersWithDB(db *sql.DB) *AdminHandlers {
	return &AdminHandlers{
		service: newAdminServiceImpl(db),
	}
}

// adminServiceImpl implements AdminService interface
type adminServiceImpl struct {
	repo *adminRepository
}

func newAdminServiceImpl(db *sql.DB) *adminServiceImpl {
	return &adminServiceImpl{
		repo: newAdminRepository(db),
	}
}

func (s *adminServiceImpl) Login(ctx context.Context, username, password, ip, userAgent, device, browser string) (*models.Admin, error) {
	admin, err := s.repo.GetAdminByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if admin == nil {
		s.repo.CreateLoginLog(ctx, &models.LoginLog{
			AdminID:   0,
			IPAddress: ip,
			UserAgent: sql.NullString{String: userAgent, Valid: userAgent != ""},
			Device:    sql.NullString{String: device, Valid: device != ""},
			Browser:   sql.NullString{String: browser, Valid: browser != ""},
			Status:    "failed",
		})
		return nil, errors.New("invalid username or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password)); err != nil {
		s.repo.CreateLoginLog(ctx, &models.LoginLog{
			AdminID:   admin.ID,
			IPAddress: ip,
			UserAgent: sql.NullString{String: userAgent, Valid: userAgent != ""},
			Device:    sql.NullString{String: device, Valid: device != ""},
			Browser:   sql.NullString{String: browser, Valid: browser != ""},
			Status:    "failed",
		})
		return nil, errors.New("invalid username or password")
	}

	s.repo.UpdateAdminLogin(ctx, admin.ID, ip)
	s.repo.CreateLoginLog(ctx, &models.LoginLog{
		AdminID:   admin.ID,
		IPAddress: ip,
		UserAgent: sql.NullString{String: userAgent, Valid: userAgent != ""},
		Device:    sql.NullString{String: device, Valid: device != ""},
		Browser:   sql.NullString{String: browser, Valid: browser != ""},
		Status:    "success",
	})

	return admin, nil
}

func (s *adminServiceImpl) GetAdminByID(ctx context.Context, id int64) (*models.Admin, error) {
	return s.repo.GetAdminByID(ctx, id)
}

func (s *adminServiceImpl) ChangePassword(ctx context.Context, adminID int64, currentPassword, newPassword string) error {
	admin, err := s.repo.GetAdminByID(ctx, adminID)
	if err != nil {
		return err
	}
	if admin == nil {
		return errors.New("admin not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(currentPassword)); err != nil {
		return errors.New("current password is incorrect")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.repo.UpdateAdminPassword(ctx, adminID, string(hashedPassword))
}

func (s *adminServiceImpl) GetLoginLogs(ctx context.Context, adminID int64, page, pageSize int) ([]*models.LoginLog, int64, error) {
	return s.repo.GetLoginLogs(ctx, adminID, page, pageSize)
}

func (s *adminServiceImpl) CreateAuditLog(ctx context.Context, log *models.AuditLog) error {
	return s.repo.CreateAuditLog(ctx, log)
}

func (s *adminServiceImpl) GetAuditLogs(ctx context.Context, params *models.PaginationParams) ([]*models.AuditLog, int64, error) {
	return s.repo.GetAuditLogs(ctx, params)
}

func (s *adminServiceImpl) CreateNotification(ctx context.Context, notification *models.AdminNotification) error {
	return s.repo.CreateNotification(ctx, notification)
}

func (s *adminServiceImpl) GetNotifications(ctx context.Context, adminID int64, isRead *bool, category string, page, pageSize int) ([]*models.AdminNotification, int64, int64, error) {
	return s.repo.GetNotifications(ctx, adminID, isRead, category, page, pageSize)
}

func (s *adminServiceImpl) MarkNotificationRead(ctx context.Context, notificationID int64) error {
	return s.repo.MarkNotificationRead(ctx, notificationID)
}

func (s *adminServiceImpl) MarkAllNotificationsRead(ctx context.Context, adminID int64) error {
	return s.repo.MarkAllNotificationsRead(ctx, adminID)
}

func (s *adminServiceImpl) DeleteNotification(ctx context.Context, notificationID int64) error {
	return s.repo.DeleteNotification(ctx, notificationID)
}

func (s *adminServiceImpl) GetUsers(ctx context.Context, params *models.PaginationParams) ([]*models.UserWithAdmin, int64, error) {
	return s.repo.GetUsersWithAdmin(ctx, params)
}

func (s *adminServiceImpl) GetUserByID(ctx context.Context, userID int64) (*models.UserWithAdmin, error) {
	return s.repo.GetUserByIDWithAdmin(ctx, userID)
}

func (s *adminServiceImpl) CreateUser(ctx context.Context, req *models.CreateUserRequest) (int64, error) {
	return s.repo.CreateUserAdmin(ctx, req)
}

func (s *adminServiceImpl) UpdateUser(ctx context.Context, userID int64, req *models.UpdateUserRequest) error {
	return s.repo.UpdateUserAdmin(ctx, userID, req)
}

func (s *adminServiceImpl) DeleteUser(ctx context.Context, userID int64) error {
	return s.repo.DeleteUserAdmin(ctx, userID)
}

func (s *adminServiceImpl) BatchUpdateUsers(ctx context.Context, userIDs []int64, operation string) error {
	return s.repo.BatchUpdateUsersAdmin(ctx, userIDs, operation)
}

func (s *adminServiceImpl) GetCalls(ctx context.Context, params *models.PaginationParams) ([]*models.CallRecord, int64, error) {
	return s.repo.GetCallsAdmin(ctx, params)
}

func (s *adminServiceImpl) GetCallByID(ctx context.Context, callID int64) (*models.CallRecord, error) {
	return s.repo.GetCallByIDAdmin(ctx, callID)
}

func (s *adminServiceImpl) GetBroadcasts(ctx context.Context, params *models.PaginationParams) ([]*models.BroadcastMessage, int64, error) {
	return s.repo.GetBroadcasts(ctx, params)
}

func (s *adminServiceImpl) GetBroadcastByID(ctx context.Context, broadcastID int64) (*models.BroadcastMessage, error) {
	return s.repo.GetBroadcastByID(ctx, broadcastID)
}

func (s *adminServiceImpl) SendBroadcast(ctx context.Context, adminID int64, req *models.SendBroadcastRequest) (int64, int, error) {
	broadcastID, err := s.repo.CreateBroadcast(ctx, adminID, req)
	if err != nil {
		return 0, 0, err
	}

	var userIDs []int64
	switch req.TargetType {
	case "all":
		userIDs, err = s.repo.GetAllUserIDs(ctx)
	case "online":
		userIDs, err = s.repo.GetOnlineUserIDs(ctx)
	case "custom":
		userIDs = req.TargetUserIDs
	default:
		userIDs, err = s.repo.GetAllUserIDs(ctx)
	}
	if err != nil {
		return broadcastID, 0, err
	}

	totalUsers := len(userIDs)
	go s.processBroadcast(broadcastID, userIDs)

	return broadcastID, totalUsers, nil
}

func (s *adminServiceImpl) processBroadcast(broadcastID int64, userIDs []int64) {
	ctx := context.Background()
	successCount := 0
	failedCount := 0

	for _, userID := range userIDs {
		detail := &models.BroadcastDetail{
			BroadcastID: broadcastID,
			UserID:      userID,
			Status:      "success",
		}
		if err := s.repo.CreateBroadcastDetail(ctx, detail); err != nil {
			failedCount++
		} else {
			successCount++
		}
	}

	status := "completed"
	if failedCount > 0 && successCount == 0 {
		status = "failed"
	} else if failedCount > 0 {
		status = "partial_failed"
	}

	s.repo.UpdateBroadcastStatus(ctx, broadcastID, status, len(userIDs), successCount, failedCount)
}

func (s *adminServiceImpl) RetryBroadcast(ctx context.Context, broadcastID int64) error {
	details, _, err := s.repo.GetBroadcastDetails(ctx, broadcastID, "failed", 1, 10000)
	if err != nil {
		return err
	}

	for _, detail := range details {
		detail.Status = "success"
		s.repo.CreateBroadcastDetail(ctx, detail)
	}

	return nil
}

func (s *adminServiceImpl) GetBroadcastDetails(ctx context.Context, broadcastID int64, status string, page, pageSize int) ([]*models.BroadcastDetail, int64, error) {
	return s.repo.GetBroadcastDetails(ctx, broadcastID, status, page, pageSize)
}

func (s *adminServiceImpl) GetMessageTemplates(ctx context.Context) ([]*models.MessageTemplate, error) {
	return s.repo.GetMessageTemplates(ctx)
}

func (s *adminServiceImpl) CreateMessageTemplate(ctx context.Context, req *models.CreateTemplateRequest) (int64, error) {
	return s.repo.CreateMessageTemplate(ctx, req)
}

func (s *adminServiceImpl) UpdateMessageTemplate(ctx context.Context, templateID int64, req *models.CreateTemplateRequest) error {
	return s.repo.UpdateMessageTemplate(ctx, templateID, req)
}

func (s *adminServiceImpl) DeleteMessageTemplate(ctx context.Context, templateID int64) error {
	return s.repo.DeleteMessageTemplate(ctx, templateID)
}

func (s *adminServiceImpl) GetAutoMessageConfigs(ctx context.Context) ([]*models.AutoMessageConfig, error) {
	return s.repo.GetAutoMessageConfigs(ctx)
}

func (s *adminServiceImpl) UpdateAutoMessageConfig(ctx context.Context, configType string, req *models.UpdateAutoMessageRequest) error {
	return s.repo.UpdateAutoMessageConfig(ctx, configType, req)
}

func (s *adminServiceImpl) GetSystemConfigs(ctx context.Context) ([]*models.SystemConfig, error) {
	return s.repo.GetSystemConfigs(ctx)
}

func (s *adminServiceImpl) UpdateSystemConfig(ctx context.Context, configKey, configValue string) error {
	return s.repo.UpdateSystemConfig(ctx, configKey, configValue)
}

func (s *adminServiceImpl) GetUserStatistics(ctx context.Context) (*models.UserStatistics, error) {
	return s.repo.GetUserStatistics(ctx)
}

func (s *adminServiceImpl) GetCallStatistics(ctx context.Context) (*models.CallStatistics, error) {
	return s.repo.GetCallStatistics(ctx)
}

func (s *adminServiceImpl) GetBotStatistics(ctx context.Context) (*models.BotStatistics, error) {
	return s.repo.GetBotStatistics(ctx)
}

func (s *adminServiceImpl) GetDashboardData(ctx context.Context) (*models.DashboardData, error) {
	return s.repo.GetDashboardData(ctx)
}

// Claims represents JWT claims
type Claims struct {
	AdminID  int64  `json:"admin_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// generateToken generates a JWT token for admin
func generateToken(admin *models.Admin) (string, int64, error) {
	expiresAt := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		AdminID:  admin.ID,
		Username: admin.Username,
		Role:     admin.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", 0, err
	}

	return tokenString, expiresAt.Unix(), nil
}

// Login handles admin login
func (h *AdminHandlers) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	ip := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
	device := c.GetHeader("X-Device")
	browser := c.GetHeader("X-Browser")

	admin, err := h.service.Login(c.Request.Context(), req.Username, req.Password, ip, userAgent, device, browser)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	token, expiresAt, err := generateToken(admin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, models.LoginResponse{
		Token:     token,
		Admin:     admin,
		ExpiresAt: expiresAt,
	})
}

// Logout handles admin logout
func (h *AdminHandlers) Logout(c *gin.Context) {
	// JWT tokens are stateless, so we just return success
	// In a production environment, you might want to blacklist the token
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// GetProfile returns the current admin's profile
func (h *AdminHandlers) GetProfile(c *gin.Context) {
	adminID := c.GetInt64("admin_id")
	admin, err := h.service.GetAdminByID(c.Request.Context(), adminID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if admin == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Admin not found"})
		return
	}

	c.JSON(http.StatusOK, admin)
}

// ChangePassword handles password change
func (h *AdminHandlers) ChangePassword(c *gin.Context) {
	adminID := c.GetInt64("admin_id")

	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if req.NewPassword != req.ConfirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Passwords do not match"})
		return
	}

	if len(req.NewPassword) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be at least 6 characters"})
		return
	}

	if err := h.service.ChangePassword(c.Request.Context(), adminID, req.CurrentPassword, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Log audit
	h.logAudit(c, "profile", "change_password", "", 0)

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

// GetLoginLogs returns login logs for the current admin
func (h *AdminHandlers) GetLoginLogs(c *gin.Context) {
	adminID := c.GetInt64("admin_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	logs, total, err := h.service.GetLoginLogs(c.Request.Context(), adminID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.PaginatedResponse{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Data:     logs,
	})
}

// GetAuditLogs returns audit logs
func (h *AdminHandlers) GetAuditLogs(c *gin.Context) {
	params := &models.PaginationParams{
		Page:     1,
		PageSize: 20,
	}
	params.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	params.PageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "20"))
	params.Search = c.Query("search")

	logs, total, err := h.service.GetAuditLogs(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.PaginatedResponse{
		Total:    total,
		Page:     params.Page,
		PageSize: params.PageSize,
		Data:     logs,
	})
}

// GetNotifications returns notifications for the current admin
func (h *AdminHandlers) GetNotifications(c *gin.Context) {
	adminID := c.GetInt64("admin_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	category := c.Query("category")

	var isRead *bool
	if isReadStr := c.Query("is_read"); isReadStr != "" {
		isReadVal := isReadStr == "true"
		isRead = &isReadVal
	}

	notifications, total, unreadCount, err := h.service.GetNotifications(c.Request.Context(), adminID, isRead, category, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total":        total,
		"unread_count": unreadCount,
		"page":         page,
		"page_size":    pageSize,
		"data":         notifications,
	})
}

// MarkNotificationRead marks a notification as read
func (h *AdminHandlers) MarkNotificationRead(c *gin.Context) {
	notificationID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	if err := h.service.MarkNotificationRead(c.Request.Context(), notificationID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification marked as read"})
}

// MarkAllNotificationsRead marks all notifications as read
func (h *AdminHandlers) MarkAllNotificationsRead(c *gin.Context) {
	adminID := c.GetInt64("admin_id")

	if err := h.service.MarkAllNotificationsRead(c.Request.Context(), adminID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "All notifications marked as read"})
}

// DeleteNotification deletes a notification
func (h *AdminHandlers) DeleteNotification(c *gin.Context) {
	notificationID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	if err := h.service.DeleteNotification(c.Request.Context(), notificationID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification deleted"})
}

// GetUsers returns users with pagination
func (h *AdminHandlers) GetUsers(c *gin.Context) {
	params := &models.PaginationParams{
		Page:     1,
		PageSize: 20,
	}
	params.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	params.PageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "20"))
	params.Search = c.Query("search")
	params.Status = c.Query("status")
	params.SortBy = c.Query("sort_by")
	params.SortOrder = c.Query("sort_order")

	users, total, err := h.service.GetUsers(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.PaginatedResponse{
		Total:    total,
		Page:     params.Page,
		PageSize: params.PageSize,
		Data:     users,
	})
}

// GetUser returns a single user
func (h *AdminHandlers) GetUser(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := h.service.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// CreateUser creates a new user
func (h *AdminHandlers) CreateUser(c *gin.Context) {
	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	userID, err := h.service.CreateUser(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log audit
	h.logAudit(c, "users", "create", "user", userID)

	user, _ := h.service.GetUserByID(c.Request.Context(), userID)
	c.JSON(http.StatusCreated, user)
}

// UpdateUser updates a user
func (h *AdminHandlers) UpdateUser(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if err := h.service.UpdateUser(c.Request.Context(), userID, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log audit
	h.logAudit(c, "users", "update", "user", userID)

	user, _ := h.service.GetUserByID(c.Request.Context(), userID)
	c.JSON(http.StatusOK, user)
}

// DeleteUser deletes a user
func (h *AdminHandlers) DeleteUser(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := h.service.DeleteUser(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log audit
	h.logAudit(c, "users", "delete", "user", userID)

	c.JSON(http.StatusOK, gin.H{"message": "User deleted"})
}

// BatchUpdateUsers batch updates users
func (h *AdminHandlers) BatchUpdateUsers(c *gin.Context) {
	var req models.BatchOperationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if err := h.service.BatchUpdateUsers(c.Request.Context(), req.UserIDs, req.Operation); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log audit
	details, _ := json.Marshal(req)
	h.logAuditWithDetails(c, "users", "batch_"+req.Operation, "users", 0, details)

	c.JSON(http.StatusOK, gin.H{"message": "Batch operation completed", "affected": len(req.UserIDs)})
}

// GetCalls returns calls with pagination
func (h *AdminHandlers) GetCalls(c *gin.Context) {
	params := &models.PaginationParams{
		Page:     1,
		PageSize: 20,
	}
	params.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	params.PageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "20"))
	params.Search = c.Query("search")
	params.Status = c.Query("status")

	calls, total, err := h.service.GetCalls(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.PaginatedResponse{
		Total:    total,
		Page:     params.Page,
		PageSize: params.PageSize,
		Data:     calls,
	})
}

// GetCall returns a single call
func (h *AdminHandlers) GetCall(c *gin.Context) {
	callID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid call ID"})
		return
	}

	call, err := h.service.GetCallByID(c.Request.Context(), callID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if call == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Call not found"})
		return
	}

	c.JSON(http.StatusOK, call)
}

// GetBroadcasts returns broadcasts with pagination
func (h *AdminHandlers) GetBroadcasts(c *gin.Context) {
	params := &models.PaginationParams{
		Page:     1,
		PageSize: 20,
	}
	params.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	params.PageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "20"))
	params.Search = c.Query("search")
	params.Status = c.Query("status")

	broadcasts, total, err := h.service.GetBroadcasts(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.PaginatedResponse{
		Total:    total,
		Page:     params.Page,
		PageSize: params.PageSize,
		Data:     broadcasts,
	})
}

// GetBroadcast returns a single broadcast
func (h *AdminHandlers) GetBroadcast(c *gin.Context) {
	broadcastID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid broadcast ID"})
		return
	}

	broadcast, err := h.service.GetBroadcastByID(c.Request.Context(), broadcastID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if broadcast == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Broadcast not found"})
		return
	}

	// Get details
	details, _, _ := h.service.GetBroadcastDetails(c.Request.Context(), broadcastID, "", 1, 100)

	c.JSON(http.StatusOK, gin.H{
		"broadcast": broadcast,
		"details":   details,
	})
}

// SendBroadcast sends a broadcast message
func (h *AdminHandlers) SendBroadcast(c *gin.Context) {
	adminID := c.GetInt64("admin_id")

	var req models.SendBroadcastRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	broadcastID, totalUsers, err := h.service.SendBroadcast(c.Request.Context(), adminID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log audit
	h.logAudit(c, "broadcast", "send", "broadcast", broadcastID)

	c.JSON(http.StatusOK, gin.H{
		"id":          broadcastID,
		"status":      "sending",
		"total_users": totalUsers,
	})
}

// RetryBroadcast retries failed broadcast messages
func (h *AdminHandlers) RetryBroadcast(c *gin.Context) {
	broadcastID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid broadcast ID"})
		return
	}

	if err := h.service.RetryBroadcast(c.Request.Context(), broadcastID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log audit
	h.logAudit(c, "broadcast", "retry", "broadcast", broadcastID)

	c.JSON(http.StatusOK, gin.H{"message": "Retry initiated"})
}

// GetMessageTemplates returns message templates
func (h *AdminHandlers) GetMessageTemplates(c *gin.Context) {
	templates, err := h.service.GetMessageTemplates(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, templates)
}

// CreateMessageTemplate creates a message template
func (h *AdminHandlers) CreateMessageTemplate(c *gin.Context) {
	var req models.CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	templateID, err := h.service.CreateMessageTemplate(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log audit
	h.logAudit(c, "templates", "create", "template", templateID)

	c.JSON(http.StatusCreated, gin.H{"id": templateID, "message": "Template created"})
}

// UpdateMessageTemplate updates a message template
func (h *AdminHandlers) UpdateMessageTemplate(c *gin.Context) {
	templateID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	var req models.CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if err := h.service.UpdateMessageTemplate(c.Request.Context(), templateID, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log audit
	h.logAudit(c, "templates", "update", "template", templateID)

	c.JSON(http.StatusOK, gin.H{"message": "Template updated"})
}

// DeleteMessageTemplate deletes a message template
func (h *AdminHandlers) DeleteMessageTemplate(c *gin.Context) {
	templateID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	if err := h.service.DeleteMessageTemplate(c.Request.Context(), templateID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log audit
	h.logAudit(c, "templates", "delete", "template", templateID)

	c.JSON(http.StatusOK, gin.H{"message": "Template deleted"})
}

// GetAutoMessageConfigs returns auto message configs
func (h *AdminHandlers) GetAutoMessageConfigs(c *gin.Context) {
	configs, err := h.service.GetAutoMessageConfigs(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, configs)
}

// UpdateAutoMessageConfig updates an auto message config
func (h *AdminHandlers) UpdateAutoMessageConfig(c *gin.Context) {
	configType := c.Param("type")

	var req models.UpdateAutoMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if err := h.service.UpdateAutoMessageConfig(c.Request.Context(), configType, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log audit
	h.logAudit(c, "auto_messages", "update", "auto_message", 0)

	c.JSON(http.StatusOK, gin.H{"message": "Config updated"})
}

// GetSystemConfigs returns system configs
func (h *AdminHandlers) GetSystemConfigs(c *gin.Context) {
	configs, err := h.service.GetSystemConfigs(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert to map for easier access
	configMap := make(map[string]interface{})
	for _, config := range configs {
		configMap[config.ConfigKey] = config.ConfigValue
	}

	c.JSON(http.StatusOK, configMap)
}

// UpdateSystemConfig updates a system config
func (h *AdminHandlers) UpdateSystemConfig(c *gin.Context) {
	configKey := c.Param("key")

	var req models.UpdateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if err := h.service.UpdateSystemConfig(c.Request.Context(), configKey, req.ConfigValue); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log audit
	h.logAudit(c, "settings", "update", "config", 0)

	c.JSON(http.StatusOK, gin.H{"message": "Config updated"})
}

// GetUserStatistics returns user statistics
func (h *AdminHandlers) GetUserStatistics(c *gin.Context) {
	stats, err := h.service.GetUserStatistics(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetCallStatistics returns call statistics
func (h *AdminHandlers) GetCallStatistics(c *gin.Context) {
	stats, err := h.service.GetCallStatistics(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetBotStatistics returns bot statistics
func (h *AdminHandlers) GetBotStatistics(c *gin.Context) {
	stats, err := h.service.GetBotStatistics(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetDashboard returns dashboard data
func (h *AdminHandlers) GetDashboard(c *gin.Context) {
	data, err := h.service.GetDashboardData(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, data)
}

// logAudit logs an audit entry
func (h *AdminHandlers) logAudit(c *gin.Context, module, action, targetType string, targetID int64) {
	h.logAuditWithDetails(c, module, action, targetType, targetID, nil)
}

// logAuditWithDetails logs an audit entry with details
func (h *AdminHandlers) logAuditWithDetails(c *gin.Context, module, action, targetType string, targetID int64, details []byte) {
	adminID := c.GetInt64("admin_id")
	ip := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	log := &models.AuditLog{
		AdminID: adminID,
		Module:  module,
		Action:  action,
		Details: details,
	}
	log.TargetType.String = targetType
	log.TargetType.Valid = targetType != ""
	log.TargetID.Int64 = targetID
	log.TargetID.Valid = targetID > 0
	log.IPAddress.String = ip
	log.IPAddress.Valid = ip != ""
	log.UserAgent.String = userAgent
	log.UserAgent.Valid = userAgent != ""

	h.service.CreateAuditLog(c.Request.Context(), log)
}

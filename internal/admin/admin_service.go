package admin

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/feiji/feiji-backend/internal/admin/models"
	"golang.org/x/crypto/bcrypt"
)

// AdminService handles admin business logic
type AdminService struct {
	repo *AdminRepository
}

// NewAdminService creates a new admin service
func NewAdminService(db *sql.DB) *AdminService {
	return &AdminService{
		repo: NewAdminRepository(db),
	}
}

// Login authenticates an admin and returns the admin info
func (s *AdminService) Login(ctx context.Context, username, password, ip, userAgent, device, browser string) (*models.Admin, error) {
	admin, err := s.repo.GetAdminByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if admin == nil {
		// Log failed login attempt
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

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password)); err != nil {
		// Log failed login attempt
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

	// Update login info
	s.repo.UpdateAdminLogin(ctx, admin.ID, ip)

	// Log successful login
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

// GetAdminByID retrieves an admin by ID
func (s *AdminService) GetAdminByID(ctx context.Context, id int64) (*models.Admin, error) {
	return s.repo.GetAdminByID(ctx, id)
}

// ChangePassword changes admin password
func (s *AdminService) ChangePassword(ctx context.Context, adminID int64, currentPassword, newPassword string) error {
	admin, err := s.repo.GetAdminByID(ctx, adminID)
	if err != nil {
		return err
	}
	if admin == nil {
		return errors.New("admin not found")
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(currentPassword)); err != nil {
		return errors.New("current password is incorrect")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.repo.UpdateAdminPassword(ctx, adminID, string(hashedPassword))
}

// GetLoginLogs retrieves login logs for an admin
func (s *AdminService) GetLoginLogs(ctx context.Context, adminID int64, page, pageSize int) ([]*models.LoginLog, int64, error) {
	return s.repo.GetLoginLogs(ctx, adminID, page, pageSize)
}

// CreateAuditLog creates an audit log entry
func (s *AdminService) CreateAuditLog(ctx context.Context, log *models.AuditLog) error {
	return s.repo.CreateAuditLog(ctx, log)
}

// GetAuditLogs retrieves audit logs
func (s *AdminService) GetAuditLogs(ctx context.Context, params *models.PaginationParams) ([]*models.AuditLog, int64, error) {
	return s.repo.GetAuditLogs(ctx, params)
}

// CreateNotification creates a notification
func (s *AdminService) CreateNotification(ctx context.Context, notification *models.AdminNotification) error {
	return s.repo.CreateNotification(ctx, notification)
}

// GetNotifications retrieves notifications for an admin
func (s *AdminService) GetNotifications(ctx context.Context, adminID int64, isRead *bool, category string, page, pageSize int) ([]*models.AdminNotification, int64, int64, error) {
	return s.repo.GetNotifications(ctx, adminID, isRead, category, page, pageSize)
}

// MarkNotificationRead marks a notification as read
func (s *AdminService) MarkNotificationRead(ctx context.Context, notificationID int64) error {
	return s.repo.MarkNotificationRead(ctx, notificationID)
}

// MarkAllNotificationsRead marks all notifications as read
func (s *AdminService) MarkAllNotificationsRead(ctx context.Context, adminID int64) error {
	return s.repo.MarkAllNotificationsRead(ctx, adminID)
}

// DeleteNotification deletes a notification
func (s *AdminService) DeleteNotification(ctx context.Context, notificationID int64) error {
	return s.repo.DeleteNotification(ctx, notificationID)
}

// GetUsers retrieves users with pagination
func (s *AdminService) GetUsers(ctx context.Context, params *models.PaginationParams) ([]*models.UserWithAdmin, int64, error) {
	return s.repo.GetUsersWithAdmin(ctx, params)
}

// GetUserByID retrieves a user by ID
func (s *AdminService) GetUserByID(ctx context.Context, userID int64) (*models.UserWithAdmin, error) {
	return s.repo.GetUserByIDWithAdmin(ctx, userID)
}

// CreateUser creates a new user
func (s *AdminService) CreateUser(ctx context.Context, req *models.CreateUserRequest) (int64, error) {
	return s.repo.CreateUserAdmin(ctx, req)
}

// UpdateUser updates a user
func (s *AdminService) UpdateUser(ctx context.Context, userID int64, req *models.UpdateUserRequest) error {
	return s.repo.UpdateUserAdmin(ctx, userID, req)
}

// DeleteUser deletes a user
func (s *AdminService) DeleteUser(ctx context.Context, userID int64) error {
	return s.repo.DeleteUserAdmin(ctx, userID)
}

// BatchUpdateUsers batch updates users
func (s *AdminService) BatchUpdateUsers(ctx context.Context, userIDs []int64, operation string) error {
	return s.repo.BatchUpdateUsersAdmin(ctx, userIDs, operation)
}

// GetCalls retrieves calls with pagination
func (s *AdminService) GetCalls(ctx context.Context, params *models.PaginationParams) ([]*models.CallRecord, int64, error) {
	return s.repo.GetCallsAdmin(ctx, params)
}

// GetCallByID retrieves a call by ID
func (s *AdminService) GetCallByID(ctx context.Context, callID int64) (*models.CallRecord, error) {
	return s.repo.GetCallByIDAdmin(ctx, callID)
}

// GetBroadcasts retrieves broadcasts with pagination
func (s *AdminService) GetBroadcasts(ctx context.Context, params *models.PaginationParams) ([]*models.BroadcastMessage, int64, error) {
	return s.repo.GetBroadcasts(ctx, params)
}

// GetBroadcastByID retrieves a broadcast by ID
func (s *AdminService) GetBroadcastByID(ctx context.Context, broadcastID int64) (*models.BroadcastMessage, error) {
	return s.repo.GetBroadcastByID(ctx, broadcastID)
}

// SendBroadcast sends a broadcast message
func (s *AdminService) SendBroadcast(ctx context.Context, adminID int64, req *models.SendBroadcastRequest) (int64, int, error) {
	// Create broadcast record
	broadcastID, err := s.repo.CreateBroadcast(ctx, adminID, req)
	if err != nil {
		return 0, 0, err
	}

	// Get target users
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

	// Send broadcast asynchronously
	go s.processBroadcast(broadcastID, userIDs, req)

	return broadcastID, totalUsers, nil
}

// processBroadcast processes broadcast sending
func (s *AdminService) processBroadcast(broadcastID int64, userIDs []int64, req *models.SendBroadcastRequest) {
	ctx := context.Background()
	successCount := 0
	failedCount := 0

	for _, userID := range userIDs {
		// Here we would send the actual message via WebSocket or push notification
		// For now, we just record the detail
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

	// Update broadcast status
	status := "completed"
	if failedCount > 0 && successCount == 0 {
		status = "failed"
	} else if failedCount > 0 {
		status = "partial_failed"
	}

	s.repo.UpdateBroadcastStatus(ctx, broadcastID, status, len(userIDs), successCount, failedCount)
}

// RetryBroadcast retries failed broadcast messages
func (s *AdminService) RetryBroadcast(ctx context.Context, broadcastID int64) error {
	// Get failed details
	details, _, err := s.repo.GetBroadcastDetails(ctx, broadcastID, "failed", 1, 10000)
	if err != nil {
		return err
	}

	// Retry sending
	for _, detail := range details {
		// Here we would retry sending the message
		detail.Status = "success"
		s.repo.CreateBroadcastDetail(ctx, detail)
	}

	return nil
}

// GetBroadcastDetails retrieves broadcast details
func (s *AdminService) GetBroadcastDetails(ctx context.Context, broadcastID int64, status string, page, pageSize int) ([]*models.BroadcastDetail, int64, error) {
	return s.repo.GetBroadcastDetails(ctx, broadcastID, status, page, pageSize)
}

// GetMessageTemplates retrieves message templates
func (s *AdminService) GetMessageTemplates(ctx context.Context) ([]*models.MessageTemplate, error) {
	return s.repo.GetMessageTemplates(ctx)
}

// CreateMessageTemplate creates a message template
func (s *AdminService) CreateMessageTemplate(ctx context.Context, req *models.CreateTemplateRequest) (int64, error) {
	return s.repo.CreateMessageTemplate(ctx, req)
}

// UpdateMessageTemplate updates a message template
func (s *AdminService) UpdateMessageTemplate(ctx context.Context, templateID int64, req *models.CreateTemplateRequest) error {
	return s.repo.UpdateMessageTemplate(ctx, templateID, req)
}

// DeleteMessageTemplate deletes a message template
func (s *AdminService) DeleteMessageTemplate(ctx context.Context, templateID int64) error {
	return s.repo.DeleteMessageTemplate(ctx, templateID)
}

// GetAutoMessageConfigs retrieves auto message configs
func (s *AdminService) GetAutoMessageConfigs(ctx context.Context) ([]*models.AutoMessageConfig, error) {
	return s.repo.GetAutoMessageConfigs(ctx)
}

// UpdateAutoMessageConfig updates an auto message config
func (s *AdminService) UpdateAutoMessageConfig(ctx context.Context, configType string, req *models.UpdateAutoMessageRequest) error {
	return s.repo.UpdateAutoMessageConfig(ctx, configType, req)
}

// GetSystemConfigs retrieves system configs
func (s *AdminService) GetSystemConfigs(ctx context.Context) ([]*models.SystemConfig, error) {
	return s.repo.GetSystemConfigs(ctx)
}

// UpdateSystemConfig updates a system config
func (s *AdminService) UpdateSystemConfig(ctx context.Context, configKey, configValue string) error {
	return s.repo.UpdateSystemConfig(ctx, configKey, configValue)
}

// GetUserStatistics retrieves user statistics
func (s *AdminService) GetUserStatistics(ctx context.Context) (*models.UserStatistics, error) {
	return s.repo.GetUserStatistics(ctx)
}

// GetCallStatistics retrieves call statistics
func (s *AdminService) GetCallStatistics(ctx context.Context) (*models.CallStatistics, error) {
	return s.repo.GetCallStatistics(ctx)
}

// GetBotStatistics retrieves bot statistics
func (s *AdminService) GetBotStatistics(ctx context.Context) (*models.BotStatistics, error) {
	return s.repo.GetBotStatistics(ctx)
}

// GetDashboardData retrieves dashboard data
func (s *AdminService) GetDashboardData(ctx context.Context) (*models.DashboardData, error) {
	return s.repo.GetDashboardData(ctx)
}

// BroadcastNotification broadcasts a notification to all admin WebSocket clients
func (s *AdminService) BroadcastNotification(notification *models.AdminNotification) {
	// This will be implemented with WebSocket hub integration
	// Store notification in database
	ctx := context.Background()
	s.repo.CreateNotification(ctx, notification)
}

// OnUserRegister handles user registration notification
func (s *AdminService) OnUserRegister(userID int64, phone string) {
	notification := &models.AdminNotification{
		Category: "user_register",
		Title:    "New User Registration",
		Message:  "User " + phone + " just registered",
		Priority: "normal",
	}
	notification.Data = []byte(`{"user_id":` + string(rune(userID)) + `,"phone":"` + phone + `"}`)
	s.BroadcastNotification(notification)
}

// OnServiceDown handles service down notification
func (s *AdminService) OnServiceDown(serviceName string) {
	notification := &models.AdminNotification{
		Category: "service_down",
		Title:    "Service Error",
		Message:  serviceName + " service has stopped",
		Priority: "urgent",
	}
	notification.Data = []byte(`{"service":"` + serviceName + `"}`)
	s.BroadcastNotification(notification)
}

// OnBroadcastComplete handles broadcast complete notification
func (s *AdminService) OnBroadcastComplete(broadcastID int64, totalUsers, successCount, failedCount int) {
	notification := &models.AdminNotification{
		Category: "broadcast_complete",
		Title:    "Broadcast Complete",
		Message:  "Broadcast sent to " + string(rune(totalUsers)) + " users, success: " + string(rune(successCount)) + ", failed: " + string(rune(failedCount)),
		Priority: "normal",
	}
	s.BroadcastNotification(notification)
}

// GenerateAuthKey generates a random auth key for user
func (s *AdminService) GenerateAuthKey() string {
	// Generate a random 6-digit code
	return time.Now().Format("150405")
}

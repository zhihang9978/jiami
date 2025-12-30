package admin

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/feiji/feiji-backend/internal/admin/models"
)

// AdminRepository handles admin panel database operations
type AdminRepository struct {
	db *sql.DB
}

// NewAdminRepository creates a new admin repository
func NewAdminRepository(db *sql.DB) *AdminRepository {
	return &AdminRepository{db: db}
}

// GetAdminByUsername retrieves an admin by username
func (r *AdminRepository) GetAdminByUsername(ctx context.Context, username string) (*models.Admin, error) {
	query := `SELECT id, username, password_hash, role, created_at, updated_at, last_login_at, last_login_ip 
			  FROM admins WHERE username = ?`
	
	var admin models.Admin
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&admin.ID, &admin.Username, &admin.PasswordHash, &admin.Role,
		&admin.CreatedAt, &admin.UpdatedAt, &admin.LastLoginAt, &admin.LastLoginIP,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

// GetAdminByID retrieves an admin by ID
func (r *AdminRepository) GetAdminByID(ctx context.Context, id int64) (*models.Admin, error) {
	query := `SELECT id, username, password_hash, role, created_at, updated_at, last_login_at, last_login_ip 
			  FROM admins WHERE id = ?`
	
	var admin models.Admin
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&admin.ID, &admin.Username, &admin.PasswordHash, &admin.Role,
		&admin.CreatedAt, &admin.UpdatedAt, &admin.LastLoginAt, &admin.LastLoginIP,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

// UpdateAdminLogin updates admin login info
func (r *AdminRepository) UpdateAdminLogin(ctx context.Context, adminID int64, ip string) error {
	query := `UPDATE admins SET last_login_at = NOW(), last_login_ip = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, ip, adminID)
	return err
}

// UpdateAdminPassword updates admin password
func (r *AdminRepository) UpdateAdminPassword(ctx context.Context, adminID int64, passwordHash string) error {
	query := `UPDATE admins SET password_hash = ?, updated_at = NOW() WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, passwordHash, adminID)
	return err
}

// CreateLoginLog creates a login log entry
func (r *AdminRepository) CreateLoginLog(ctx context.Context, log *models.LoginLog) error {
	query := `INSERT INTO login_logs (admin_id, ip_address, user_agent, device, browser, status) 
			  VALUES (?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, log.AdminID, log.IPAddress, log.UserAgent, log.Device, log.Browser, log.Status)
	return err
}

// GetLoginLogs retrieves login logs for an admin
func (r *AdminRepository) GetLoginLogs(ctx context.Context, adminID int64, page, pageSize int) ([]*models.LoginLog, int64, error) {
	offset := (page - 1) * pageSize
	
	var total int64
	countQuery := `SELECT COUNT(*) FROM login_logs WHERE admin_id = ?`
	if err := r.db.QueryRowContext(ctx, countQuery, adminID).Scan(&total); err != nil {
		return nil, 0, err
	}
	
	query := `SELECT id, admin_id, ip_address, user_agent, device, browser, status, created_at 
			  FROM login_logs WHERE admin_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`
	
	rows, err := r.db.QueryContext(ctx, query, adminID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	var logs []*models.LoginLog
	for rows.Next() {
		var log models.LoginLog
		if err := rows.Scan(&log.ID, &log.AdminID, &log.IPAddress, &log.UserAgent, &log.Device, &log.Browser, &log.Status, &log.CreatedAt); err != nil {
			return nil, 0, err
		}
		logs = append(logs, &log)
	}
	return logs, total, nil
}

// CreateAuditLog creates an audit log entry
func (r *AdminRepository) CreateAuditLog(ctx context.Context, log *models.AuditLog) error {
	query := `INSERT INTO audit_logs (admin_id, module, action, target_type, target_id, ip_address, user_agent, details) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, log.AdminID, log.Module, log.Action, log.TargetType, log.TargetID, log.IPAddress, log.UserAgent, log.Details)
	return err
}

// GetAuditLogs retrieves audit logs with pagination
func (r *AdminRepository) GetAuditLogs(ctx context.Context, params *models.PaginationParams) ([]*models.AuditLog, int64, error) {
	offset := (params.Page - 1) * params.PageSize
	
	whereClause := "1=1"
	args := []interface{}{}
	
	if params.Search != "" {
		whereClause += " AND (module LIKE ? OR action LIKE ?)"
		args = append(args, "%"+params.Search+"%", "%"+params.Search+"%")
	}
	
	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM audit_logs WHERE %s", whereClause)
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	
	query := fmt.Sprintf(`SELECT id, admin_id, module, action, target_type, target_id, ip_address, user_agent, details, created_at 
			  FROM audit_logs WHERE %s ORDER BY created_at DESC LIMIT ? OFFSET ?`, whereClause)
	args = append(args, params.PageSize, offset)
	
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	var logs []*models.AuditLog
	for rows.Next() {
		var log models.AuditLog
		if err := rows.Scan(&log.ID, &log.AdminID, &log.Module, &log.Action, &log.TargetType, &log.TargetID, &log.IPAddress, &log.UserAgent, &log.Details, &log.CreatedAt); err != nil {
			return nil, 0, err
		}
		logs = append(logs, &log)
	}
	return logs, total, nil
}

// CreateNotification creates a notification
func (r *AdminRepository) CreateNotification(ctx context.Context, notification *models.AdminNotification) error {
	query := `INSERT INTO admin_notifications (admin_id, category, title, message, data, priority, link) 
			  VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, notification.AdminID, notification.Category, notification.Title, notification.Message, notification.Data, notification.Priority, notification.Link)
	return err
}

// GetNotifications retrieves notifications for an admin
func (r *AdminRepository) GetNotifications(ctx context.Context, adminID int64, isRead *bool, category string, page, pageSize int) ([]*models.AdminNotification, int64, int64, error) {
	offset := (page - 1) * pageSize
	
	whereClause := "(admin_id IS NULL OR admin_id = ?)"
	args := []interface{}{adminID}
	
	if isRead != nil {
		whereClause += " AND is_read = ?"
		args = append(args, *isRead)
	}
	if category != "" {
		whereClause += " AND category = ?"
		args = append(args, category)
	}
	
	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM admin_notifications WHERE %s", whereClause)
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, 0, err
	}
	
	var unreadCount int64
	unreadQuery := `SELECT COUNT(*) FROM admin_notifications WHERE (admin_id IS NULL OR admin_id = ?) AND is_read = FALSE`
	r.db.QueryRowContext(ctx, unreadQuery, adminID).Scan(&unreadCount)
	
	query := fmt.Sprintf(`SELECT id, admin_id, category, title, message, data, priority, is_read, link, created_at 
			  FROM admin_notifications WHERE %s ORDER BY created_at DESC LIMIT ? OFFSET ?`, whereClause)
	args = append(args, pageSize, offset)
	
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, 0, err
	}
	defer rows.Close()
	
	var notifications []*models.AdminNotification
	for rows.Next() {
		var n models.AdminNotification
		if err := rows.Scan(&n.ID, &n.AdminID, &n.Category, &n.Title, &n.Message, &n.Data, &n.Priority, &n.IsRead, &n.Link, &n.CreatedAt); err != nil {
			return nil, 0, 0, err
		}
		notifications = append(notifications, &n)
	}
	return notifications, total, unreadCount, nil
}

// MarkNotificationRead marks a notification as read
func (r *AdminRepository) MarkNotificationRead(ctx context.Context, notificationID int64) error {
	query := `UPDATE admin_notifications SET is_read = TRUE WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, notificationID)
	return err
}

// MarkAllNotificationsRead marks all notifications as read for an admin
func (r *AdminRepository) MarkAllNotificationsRead(ctx context.Context, adminID int64) error {
	query := `UPDATE admin_notifications SET is_read = TRUE WHERE (admin_id IS NULL OR admin_id = ?) AND is_read = FALSE`
	_, err := r.db.ExecContext(ctx, query, adminID)
	return err
}

// DeleteNotification deletes a notification
func (r *AdminRepository) DeleteNotification(ctx context.Context, notificationID int64) error {
	query := `DELETE FROM admin_notifications WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, notificationID)
	return err
}

// GetUsersWithAdmin retrieves users with admin-specific fields
func (r *AdminRepository) GetUsersWithAdmin(ctx context.Context, params *models.PaginationParams) ([]*models.UserWithAdmin, int64, error) {
	offset := (params.Page - 1) * params.PageSize
	
	whereClause := "(deleted_at IS NULL OR deleted_at = '0000-00-00 00:00:00')"
	args := []interface{}{}
	
	if params.Search != "" {
		whereClause += " AND (phone LIKE ? OR username LIKE ? OR first_name LIKE ? OR CAST(id AS CHAR) = ?)"
		args = append(args, "%"+params.Search+"%", "%"+params.Search+"%", "%"+params.Search+"%", params.Search)
	}
	if params.Status != "" {
		whereClause += " AND status = ?"
		args = append(args, params.Status)
	}
	
	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM users WHERE %s", whereClause)
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	
	orderBy := "created_at DESC"
	if params.SortBy != "" {
		order := "ASC"
		if params.SortOrder == "desc" {
			order = "DESC"
		}
		orderBy = fmt.Sprintf("%s %s", params.SortBy, order)
	}
	
	query := fmt.Sprintf(`SELECT id, phone, COALESCE(username, ''), COALESCE(first_name, ''), COALESCE(last_name, ''), COALESCE(bio, ''), 
			  COALESCE(status, 'active'), COALESCE(custom_code, ''), code_expires_at, COALESCE(allow_call, TRUE), COALESCE(allow_video_call, TRUE), 
			  COALESCE(remark, ''), COALESCE(is_bot, FALSE), created_at, updated_at, deleted_at
			  FROM users WHERE %s ORDER BY %s LIMIT ? OFFSET ?`, whereClause, orderBy)
	args = append(args, params.PageSize, offset)
	
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	var users []*models.UserWithAdmin
	for rows.Next() {
		var u models.UserWithAdmin
		if err := rows.Scan(&u.ID, &u.Phone, &u.Username, &u.FirstName, &u.LastName, &u.Bio, &u.Status,
			&u.CustomCode, &u.CodeExpiresAt, &u.AllowCall, &u.AllowVideoCall, &u.Remark, &u.IsBot,
			&u.CreatedAt, &u.UpdatedAt, &u.DeletedAt); err != nil {
			return nil, 0, err
		}
		users = append(users, &u)
	}
	return users, total, nil
}

// GetUserByIDWithAdmin retrieves a user by ID with admin fields
func (r *AdminRepository) GetUserByIDWithAdmin(ctx context.Context, userID int64) (*models.UserWithAdmin, error) {
	query := `SELECT id, phone, COALESCE(username, ''), COALESCE(first_name, ''), COALESCE(last_name, ''), COALESCE(bio, ''), 
			  COALESCE(status, 'active'), COALESCE(custom_code, ''), code_expires_at, COALESCE(allow_call, TRUE), COALESCE(allow_video_call, TRUE), 
			  COALESCE(remark, ''), COALESCE(is_bot, FALSE), created_at, updated_at, deleted_at
			  FROM users WHERE id = ?`
	
	var u models.UserWithAdmin
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&u.ID, &u.Phone, &u.Username, &u.FirstName, &u.LastName, &u.Bio, &u.Status,
		&u.CustomCode, &u.CodeExpiresAt, &u.AllowCall, &u.AllowVideoCall, &u.Remark, &u.IsBot,
		&u.CreatedAt, &u.UpdatedAt, &u.DeletedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// CreateUserAdmin creates a new user from admin panel
func (r *AdminRepository) CreateUserAdmin(ctx context.Context, req *models.CreateUserRequest) (int64, error) {
	query := `INSERT INTO users (phone, username, first_name, custom_code, code_expires_at, status, allow_call, allow_video_call, remark, created_at, updated_at) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())`
	
	var codeExpiresAt interface{}
	if req.CodeExpiresAt != "" {
		t, err := time.Parse("2006-01-02 15:04:05", req.CodeExpiresAt)
		if err == nil {
			codeExpiresAt = t
		}
	}
	
	status := req.Status
	if status == "" {
		status = "active"
	}
	
	result, err := r.db.ExecContext(ctx, query, req.Phone, req.Username, req.FirstName, req.CustomCode, codeExpiresAt, status, req.AllowCall, req.AllowVideoCall, req.Remark)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// UpdateUserAdmin updates a user from admin panel
func (r *AdminRepository) UpdateUserAdmin(ctx context.Context, userID int64, req *models.UpdateUserRequest) error {
	setClauses := []string{"updated_at = NOW()"}
	args := []interface{}{}
	
	if req.Username != "" {
		setClauses = append(setClauses, "username = ?")
		args = append(args, req.Username)
	}
	if req.FirstName != "" {
		setClauses = append(setClauses, "first_name = ?")
		args = append(args, req.FirstName)
	}
	if req.Phone != "" {
		setClauses = append(setClauses, "phone = ?")
		args = append(args, req.Phone)
	}
	if req.CustomCode != "" {
		setClauses = append(setClauses, "custom_code = ?")
		args = append(args, req.CustomCode)
	}
	if req.CodeExpiresAt != "" {
		t, err := time.Parse("2006-01-02 15:04:05", req.CodeExpiresAt)
		if err == nil {
			setClauses = append(setClauses, "code_expires_at = ?")
			args = append(args, t)
		}
	}
	if req.Status != "" {
		setClauses = append(setClauses, "status = ?")
		args = append(args, req.Status)
	}
	if req.AllowCall != nil {
		setClauses = append(setClauses, "allow_call = ?")
		args = append(args, *req.AllowCall)
	}
	if req.AllowVideoCall != nil {
		setClauses = append(setClauses, "allow_video_call = ?")
		args = append(args, *req.AllowVideoCall)
	}
	if req.Remark != "" {
		setClauses = append(setClauses, "remark = ?")
		args = append(args, req.Remark)
	}
	
	args = append(args, userID)
	query := fmt.Sprintf("UPDATE users SET %s WHERE id = ?", strings.Join(setClauses, ", "))
	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

// DeleteUserAdmin soft deletes a user from admin panel
func (r *AdminRepository) DeleteUserAdmin(ctx context.Context, userID int64) error {
	query := `UPDATE users SET deleted_at = NOW(), status = 'deleted', is_deleted = 1 WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

// BatchUpdateUsersAdmin batch updates users from admin panel
func (r *AdminRepository) BatchUpdateUsersAdmin(ctx context.Context, userIDs []int64, operation string) error {
	if len(userIDs) == 0 {
		return nil
	}
	
	placeholders := make([]string, len(userIDs))
	args := make([]interface{}, len(userIDs))
	for i, id := range userIDs {
		placeholders[i] = "?"
		args[i] = id
	}
	
	var query string
	switch operation {
	case "disable":
		query = fmt.Sprintf("UPDATE users SET status = 'disabled', is_restricted = 1, updated_at = NOW() WHERE id IN (%s)", strings.Join(placeholders, ","))
	case "enable":
		query = fmt.Sprintf("UPDATE users SET status = 'active', is_restricted = 0, updated_at = NOW() WHERE id IN (%s)", strings.Join(placeholders, ","))
	case "delete":
		query = fmt.Sprintf("UPDATE users SET deleted_at = NOW(), status = 'deleted', is_deleted = 1 WHERE id IN (%s)", strings.Join(placeholders, ","))
	default:
		return fmt.Errorf("unknown operation: %s", operation)
	}
	
	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

// GetCallsAdmin retrieves calls with pagination for admin
func (r *AdminRepository) GetCallsAdmin(ctx context.Context, params *models.PaginationParams) ([]*models.CallRecord, int64, error) {
	offset := (params.Page - 1) * params.PageSize
	
	whereClause := "1=1"
	args := []interface{}{}
	
	if params.Search != "" {
		whereClause += " AND (CAST(c.id AS CHAR) = ? OR CAST(c.caller_id AS CHAR) = ? OR CAST(c.callee_id AS CHAR) = ?)"
		args = append(args, params.Search, params.Search, params.Search)
	}
	if params.Status != "" {
		whereClause += " AND c.status = ?"
		args = append(args, params.Status)
	}
	
	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM calls c WHERE %s", whereClause)
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	
	query := fmt.Sprintf(`SELECT c.id, c.caller_id, COALESCE(u1.first_name, u1.phone, ''), c.callee_id, COALESCE(u2.first_name, u2.phone, ''),
			  c.is_video, c.status, COALESCE(c.duration, 0), c.created_at, c.updated_at, c.access_hash
			  FROM calls c
			  LEFT JOIN users u1 ON c.caller_id = u1.id
			  LEFT JOIN users u2 ON c.callee_id = u2.id
			  WHERE %s ORDER BY c.created_at DESC LIMIT ? OFFSET ?`, whereClause)
	args = append(args, params.PageSize, offset)
	
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	var calls []*models.CallRecord
	for rows.Next() {
		var c models.CallRecord
		if err := rows.Scan(&c.ID, &c.CallerID, &c.CallerName, &c.CalleeID, &c.CalleeName,
			&c.IsVideo, &c.Status, &c.Duration, &c.StartTime, &c.EndTime, &c.AccessHash); err != nil {
			return nil, 0, err
		}
		c.E2EEEnabled = true
		c.E2EEVersion = "1.0"
		calls = append(calls, &c)
	}
	return calls, total, nil
}

// GetCallByIDAdmin retrieves a call by ID for admin
func (r *AdminRepository) GetCallByIDAdmin(ctx context.Context, callID int64) (*models.CallRecord, error) {
	query := `SELECT c.id, c.caller_id, COALESCE(u1.first_name, u1.phone, ''), c.callee_id, COALESCE(u2.first_name, u2.phone, ''),
			  c.is_video, c.status, COALESCE(c.duration, 0), c.created_at, c.updated_at, c.access_hash
			  FROM calls c
			  LEFT JOIN users u1 ON c.caller_id = u1.id
			  LEFT JOIN users u2 ON c.callee_id = u2.id
			  WHERE c.id = ?`
	
	var c models.CallRecord
	err := r.db.QueryRowContext(ctx, query, callID).Scan(&c.ID, &c.CallerID, &c.CallerName, &c.CalleeID, &c.CalleeName,
		&c.IsVideo, &c.Status, &c.Duration, &c.StartTime, &c.EndTime, &c.AccessHash)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	c.E2EEEnabled = true
	c.E2EEVersion = "1.0"
	return &c, nil
}

// GetBroadcasts retrieves broadcasts with pagination
func (r *AdminRepository) GetBroadcasts(ctx context.Context, params *models.PaginationParams) ([]*models.BroadcastMessage, int64, error) {
	offset := (params.Page - 1) * params.PageSize
	
	whereClause := "1=1"
	args := []interface{}{}
	
	if params.Search != "" {
		whereClause += " AND message_content LIKE ?"
		args = append(args, "%"+params.Search+"%")
	}
	if params.Status != "" {
		whereClause += " AND status = ?"
		args = append(args, params.Status)
	}
	
	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM broadcast_messages WHERE %s", whereClause)
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	
	query := fmt.Sprintf(`SELECT id, admin_id, message_type, message_content, media_url, target_type, 
			  target_user_ids, target_filters, total_users, success_count, failed_count, status, 
			  scheduled_at, sent_at, created_at, updated_at
			  FROM broadcast_messages WHERE %s ORDER BY created_at DESC LIMIT ? OFFSET ?`, whereClause)
	args = append(args, params.PageSize, offset)
	
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	var broadcasts []*models.BroadcastMessage
	for rows.Next() {
		var b models.BroadcastMessage
		if err := rows.Scan(&b.ID, &b.AdminID, &b.MessageType, &b.MessageContent, &b.MediaURL, &b.TargetType,
			&b.TargetUserIDs, &b.TargetFilters, &b.TotalUsers, &b.SuccessCount, &b.FailedCount, &b.Status,
			&b.ScheduledAt, &b.SentAt, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, 0, err
		}
		broadcasts = append(broadcasts, &b)
	}
	return broadcasts, total, nil
}

// GetBroadcastByID retrieves a broadcast by ID
func (r *AdminRepository) GetBroadcastByID(ctx context.Context, broadcastID int64) (*models.BroadcastMessage, error) {
	query := `SELECT id, admin_id, message_type, message_content, media_url, target_type, 
			  target_user_ids, target_filters, total_users, success_count, failed_count, status, 
			  scheduled_at, sent_at, created_at, updated_at
			  FROM broadcast_messages WHERE id = ?`
	
	var b models.BroadcastMessage
	err := r.db.QueryRowContext(ctx, query, broadcastID).Scan(&b.ID, &b.AdminID, &b.MessageType, &b.MessageContent, &b.MediaURL, &b.TargetType,
		&b.TargetUserIDs, &b.TargetFilters, &b.TotalUsers, &b.SuccessCount, &b.FailedCount, &b.Status,
		&b.ScheduledAt, &b.SentAt, &b.CreatedAt, &b.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &b, nil
}

// CreateBroadcast creates a new broadcast
func (r *AdminRepository) CreateBroadcast(ctx context.Context, adminID int64, req *models.SendBroadcastRequest) (int64, error) {
	query := `INSERT INTO broadcast_messages (admin_id, message_type, message_content, media_url, target_type, target_user_ids, target_filters, scheduled_at, status) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	
	var targetUserIDs interface{}
	if len(req.TargetUserIDs) > 0 {
		data, _ := json.Marshal(req.TargetUserIDs)
		targetUserIDs = string(data)
	}
	
	var scheduledAt interface{}
	status := "sending"
	if req.ScheduledAt != "" {
		t, err := time.Parse("2006-01-02 15:04:05", req.ScheduledAt)
		if err == nil {
			scheduledAt = t
			status = "pending"
		}
	}
	
	result, err := r.db.ExecContext(ctx, query, adminID, req.MessageType, req.MessageContent, req.MediaURL, req.TargetType, targetUserIDs, req.TargetFilters, scheduledAt, status)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// UpdateBroadcastStatus updates broadcast status and counts
func (r *AdminRepository) UpdateBroadcastStatus(ctx context.Context, broadcastID int64, status string, totalUsers, successCount, failedCount int) error {
	query := `UPDATE broadcast_messages SET status = ?, total_users = ?, success_count = ?, failed_count = ?, sent_at = NOW(), updated_at = NOW() WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, status, totalUsers, successCount, failedCount, broadcastID)
	return err
}

// CreateBroadcastDetail creates a broadcast detail entry
func (r *AdminRepository) CreateBroadcastDetail(ctx context.Context, detail *models.BroadcastDetail) error {
	query := `INSERT INTO broadcast_details (broadcast_id, user_id, status, error_message) VALUES (?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, detail.BroadcastID, detail.UserID, detail.Status, detail.ErrorMessage)
	return err
}

// GetBroadcastDetails retrieves broadcast details
func (r *AdminRepository) GetBroadcastDetails(ctx context.Context, broadcastID int64, status string, page, pageSize int) ([]*models.BroadcastDetail, int64, error) {
	offset := (page - 1) * pageSize
	
	whereClause := "broadcast_id = ?"
	args := []interface{}{broadcastID}
	
	if status != "" {
		whereClause += " AND status = ?"
		args = append(args, status)
	}
	
	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM broadcast_details WHERE %s", whereClause)
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	
	query := fmt.Sprintf(`SELECT id, broadcast_id, user_id, status, error_message, sent_at 
			  FROM broadcast_details WHERE %s ORDER BY sent_at DESC LIMIT ? OFFSET ?`, whereClause)
	args = append(args, pageSize, offset)
	
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	var details []*models.BroadcastDetail
	for rows.Next() {
		var d models.BroadcastDetail
		if err := rows.Scan(&d.ID, &d.BroadcastID, &d.UserID, &d.Status, &d.ErrorMessage, &d.SentAt); err != nil {
			return nil, 0, err
		}
		details = append(details, &d)
	}
	return details, total, nil
}

// GetMessageTemplates retrieves message templates
func (r *AdminRepository) GetMessageTemplates(ctx context.Context) ([]*models.MessageTemplate, error) {
	query := `SELECT id, name, message_type, message_content, media_url, usage_count, created_at, updated_at FROM message_templates ORDER BY created_at DESC`
	
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var templates []*models.MessageTemplate
	for rows.Next() {
		var t models.MessageTemplate
		if err := rows.Scan(&t.ID, &t.Name, &t.MessageType, &t.MessageContent, &t.MediaURL, &t.UsageCount, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		templates = append(templates, &t)
	}
	return templates, nil
}

// CreateMessageTemplate creates a message template
func (r *AdminRepository) CreateMessageTemplate(ctx context.Context, req *models.CreateTemplateRequest) (int64, error) {
	query := `INSERT INTO message_templates (name, message_type, message_content, media_url) VALUES (?, ?, ?, ?)`
	result, err := r.db.ExecContext(ctx, query, req.Name, req.MessageType, req.MessageContent, req.MediaURL)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// UpdateMessageTemplate updates a message template
func (r *AdminRepository) UpdateMessageTemplate(ctx context.Context, templateID int64, req *models.CreateTemplateRequest) error {
	query := `UPDATE message_templates SET name = ?, message_type = ?, message_content = ?, media_url = ?, updated_at = NOW() WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, req.Name, req.MessageType, req.MessageContent, req.MediaURL, templateID)
	return err
}

// DeleteMessageTemplate deletes a message template
func (r *AdminRepository) DeleteMessageTemplate(ctx context.Context, templateID int64) error {
	query := `DELETE FROM message_templates WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, templateID)
	return err
}

// IncrementTemplateUsage increments template usage count
func (r *AdminRepository) IncrementTemplateUsage(ctx context.Context, templateID int64) error {
	query := `UPDATE message_templates SET usage_count = usage_count + 1 WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, templateID)
	return err
}

// GetAutoMessageConfigs retrieves auto message configs
func (r *AdminRepository) GetAutoMessageConfigs(ctx context.Context) ([]*models.AutoMessageConfig, error) {
	query := `SELECT id, type, message_content, is_enabled, trigger_condition, created_at, updated_at FROM auto_message_configs`
	
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var configs []*models.AutoMessageConfig
	for rows.Next() {
		var c models.AutoMessageConfig
		if err := rows.Scan(&c.ID, &c.Type, &c.MessageContent, &c.IsEnabled, &c.TriggerCondition, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		configs = append(configs, &c)
	}
	return configs, nil
}

// UpdateAutoMessageConfig updates an auto message config
func (r *AdminRepository) UpdateAutoMessageConfig(ctx context.Context, configType string, req *models.UpdateAutoMessageRequest) error {
	setClauses := []string{"updated_at = NOW()"}
	args := []interface{}{}
	
	if req.MessageContent != "" {
		setClauses = append(setClauses, "message_content = ?")
		args = append(args, req.MessageContent)
	}
	if req.IsEnabled != nil {
		setClauses = append(setClauses, "is_enabled = ?")
		args = append(args, *req.IsEnabled)
	}
	
	args = append(args, configType)
	query := fmt.Sprintf("UPDATE auto_message_configs SET %s WHERE type = ?", strings.Join(setClauses, ", "))
	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

// GetSystemConfigs retrieves system configs
func (r *AdminRepository) GetSystemConfigs(ctx context.Context) ([]*models.SystemConfig, error) {
	query := `SELECT id, config_key, config_value, config_type, description, created_at, updated_at FROM system_configs`
	
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var configs []*models.SystemConfig
	for rows.Next() {
		var c models.SystemConfig
		if err := rows.Scan(&c.ID, &c.ConfigKey, &c.ConfigValue, &c.ConfigType, &c.Description, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		configs = append(configs, &c)
	}
	return configs, nil
}

// UpdateSystemConfig updates a system config
func (r *AdminRepository) UpdateSystemConfig(ctx context.Context, configKey, configValue string) error {
	query := `UPDATE system_configs SET config_value = ?, updated_at = NOW() WHERE config_key = ?`
	_, err := r.db.ExecContext(ctx, query, configValue, configKey)
	return err
}

// GetUserStatistics retrieves user statistics
func (r *AdminRepository) GetUserStatistics(ctx context.Context) (*models.UserStatistics, error) {
	stats := &models.UserStatistics{}
	
	// Total users
	r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE is_deleted = 0").Scan(&stats.TotalUsers)
	
	// Today new users
	r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE is_deleted = 0 AND DATE(created_at) = CURDATE()").Scan(&stats.TodayNewUsers)
	
	// Week new users
	r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE is_deleted = 0 AND created_at >= DATE_SUB(CURDATE(), INTERVAL 7 DAY)").Scan(&stats.WeekNewUsers)
	
	// Month new users
	r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE is_deleted = 0 AND created_at >= DATE_SUB(CURDATE(), INTERVAL 30 DAY)").Scan(&stats.MonthNewUsers)
	
	// Growth trend (last 30 days)
	rows, err := r.db.QueryContext(ctx, `
		SELECT DATE(created_at) as date, COUNT(*) as count 
		FROM users WHERE is_deleted = 0 AND created_at >= DATE_SUB(CURDATE(), INTERVAL 30 DAY)
		GROUP BY DATE(created_at) ORDER BY date`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var dc models.DailyCount
			rows.Scan(&dc.Date, &dc.Count)
			stats.GrowthTrend = append(stats.GrowthTrend, dc)
		}
	}
	
	return stats, nil
}

// GetCallStatistics retrieves call statistics
func (r *AdminRepository) GetCallStatistics(ctx context.Context) (*models.CallStatistics, error) {
	stats := &models.CallStatistics{
		TypeDistribution:   make(map[string]int),
		StatusDistribution: make(map[string]int),
		HourlyDistribution: make([]int, 24),
	}
	
	// Total calls
	r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM calls").Scan(&stats.TotalCalls)
	
	// Total duration
	r.db.QueryRowContext(ctx, "SELECT COALESCE(SUM(duration), 0) FROM calls").Scan(&stats.TotalDuration)
	
	// Today calls
	r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM calls WHERE DATE(created_at) = CURDATE()").Scan(&stats.TodayCalls)
	
	// Today duration
	r.db.QueryRowContext(ctx, "SELECT COALESCE(SUM(duration), 0) FROM calls WHERE DATE(created_at) = CURDATE()").Scan(&stats.TodayDuration)
	
	// Week calls
	r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM calls WHERE created_at >= DATE_SUB(CURDATE(), INTERVAL 7 DAY)").Scan(&stats.WeekCalls)
	
	// Month calls
	r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM calls WHERE created_at >= DATE_SUB(CURDATE(), INTERVAL 30 DAY)").Scan(&stats.MonthCalls)
	
	// Success rate
	var successCount int
	r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM calls WHERE status IN ('confirmed', 'ended')").Scan(&successCount)
	if stats.TotalCalls > 0 {
		stats.SuccessRate = float64(successCount) / float64(stats.TotalCalls) * 100
	}
	
	// Type distribution
	rows, err := r.db.QueryContext(ctx, "SELECT is_video, COUNT(*) FROM calls GROUP BY is_video")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var isVideo bool
			var count int
			rows.Scan(&isVideo, &count)
			if isVideo {
				stats.TypeDistribution["video"] = count
			} else {
				stats.TypeDistribution["voice"] = count
			}
		}
	}
	
	// Status distribution
	rows2, err := r.db.QueryContext(ctx, "SELECT status, COUNT(*) FROM calls GROUP BY status")
	if err == nil {
		defer rows2.Close()
		for rows2.Next() {
			var status string
			var count int
			rows2.Scan(&status, &count)
			stats.StatusDistribution[status] = count
		}
	}
	
	// Daily trend
	rows3, err := r.db.QueryContext(ctx, `
		SELECT DATE(created_at) as date, COUNT(*) as count 
		FROM calls WHERE created_at >= DATE_SUB(CURDATE(), INTERVAL 30 DAY)
		GROUP BY DATE(created_at) ORDER BY date`)
	if err == nil {
		defer rows3.Close()
		for rows3.Next() {
			var dc models.DailyCount
			rows3.Scan(&dc.Date, &dc.Count)
			stats.DailyTrend = append(stats.DailyTrend, dc)
		}
	}
	
	return stats, nil
}

// GetBotStatistics retrieves bot statistics
func (r *AdminRepository) GetBotStatistics(ctx context.Context) (*models.BotStatistics, error) {
	stats := &models.BotStatistics{}
	
	// Total broadcasts
	r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM broadcast_messages").Scan(&stats.BroadcastCount)
	
	// Total messages sent
	r.db.QueryRowContext(ctx, "SELECT COALESCE(SUM(success_count), 0) FROM broadcast_messages").Scan(&stats.TotalMessages)
	
	// Today messages
	r.db.QueryRowContext(ctx, "SELECT COALESCE(SUM(success_count), 0) FROM broadcast_messages WHERE DATE(sent_at) = CURDATE()").Scan(&stats.TodayMessages)
	
	// Week messages
	r.db.QueryRowContext(ctx, "SELECT COALESCE(SUM(success_count), 0) FROM broadcast_messages WHERE sent_at >= DATE_SUB(CURDATE(), INTERVAL 7 DAY)").Scan(&stats.WeekMessages)
	
	// Month messages
	r.db.QueryRowContext(ctx, "SELECT COALESCE(SUM(success_count), 0) FROM broadcast_messages WHERE sent_at >= DATE_SUB(CURDATE(), INTERVAL 30 DAY)").Scan(&stats.MonthMessages)
	
	// Average delivery rate
	var totalUsers, successUsers int
	r.db.QueryRowContext(ctx, "SELECT COALESCE(SUM(total_users), 0), COALESCE(SUM(success_count), 0) FROM broadcast_messages WHERE status = 'completed'").Scan(&totalUsers, &successUsers)
	if totalUsers > 0 {
		stats.AverageDeliveryRate = float64(successUsers) / float64(totalUsers) * 100
	}
	
	return stats, nil
}

// GetDashboardData retrieves dashboard data
func (r *AdminRepository) GetDashboardData(ctx context.Context) (*models.DashboardData, error) {
	data := &models.DashboardData{}
	
	// Total users
	r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE is_deleted = 0").Scan(&data.TotalUsers)
	
	// Today calls
	r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM calls WHERE DATE(created_at) = CURDATE()").Scan(&data.TodayCalls)
	
	// Total call duration
	r.db.QueryRowContext(ctx, "SELECT COALESCE(SUM(duration), 0) FROM calls").Scan(&data.TotalCallDuration)
	
	// User growth trend
	rows, err := r.db.QueryContext(ctx, `
		SELECT DATE(created_at) as date, COUNT(*) as count 
		FROM users WHERE is_deleted = 0 AND created_at >= DATE_SUB(CURDATE(), INTERVAL 7 DAY)
		GROUP BY DATE(created_at) ORDER BY date`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var dc models.DailyCount
			rows.Scan(&dc.Date, &dc.Count)
			data.UserGrowthTrend = append(data.UserGrowthTrend, dc)
		}
	}
	
	// Call trend
	rows2, err := r.db.QueryContext(ctx, `
		SELECT DATE(created_at) as date, COUNT(*) as count 
		FROM calls WHERE created_at >= DATE_SUB(CURDATE(), INTERVAL 7 DAY)
		GROUP BY DATE(created_at) ORDER BY date`)
	if err == nil {
		defer rows2.Close()
		for rows2.Next() {
			var dc models.DailyCount
			rows2.Scan(&dc.Date, &dc.Count)
			data.CallTrend = append(data.CallTrend, dc)
		}
	}
	
	// Recent calls
	rows3, err := r.db.QueryContext(ctx, `
		SELECT c.id, c.caller_id, COALESCE(u1.first_name, u1.phone, ''), c.callee_id, COALESCE(u2.first_name, u2.phone, ''),
		c.is_video, c.status, COALESCE(c.duration, 0), c.created_at, c.updated_at, c.access_hash
		FROM calls c
		LEFT JOIN users u1 ON c.caller_id = u1.id
		LEFT JOIN users u2 ON c.callee_id = u2.id
		ORDER BY c.created_at DESC LIMIT 5`)
	if err == nil {
		defer rows3.Close()
		for rows3.Next() {
			var c models.CallRecord
			rows3.Scan(&c.ID, &c.CallerID, &c.CallerName, &c.CalleeID, &c.CalleeName,
				&c.IsVideo, &c.Status, &c.Duration, &c.StartTime, &c.EndTime, &c.AccessHash)
			c.E2EEEnabled = true
			c.E2EEVersion = "1.0"
			data.RecentCalls = append(data.RecentCalls, c)
		}
	}
	
	// Recent audit logs
	rows4, err := r.db.QueryContext(ctx, `
		SELECT id, admin_id, module, action, target_type, target_id, ip_address, user_agent, details, created_at 
		FROM audit_logs ORDER BY created_at DESC LIMIT 5`)
	if err == nil {
		defer rows4.Close()
		for rows4.Next() {
			var log models.AuditLog
			rows4.Scan(&log.ID, &log.AdminID, &log.Module, &log.Action, &log.TargetType, &log.TargetID, &log.IPAddress, &log.UserAgent, &log.Details, &log.CreatedAt)
			data.RecentLogs = append(data.RecentLogs, log)
		}
	}
	
	return data, nil
}

// GetAllUserIDs retrieves all active user IDs
func (r *AdminRepository) GetAllUserIDs(ctx context.Context) ([]int64, error) {
	query := `SELECT id FROM users WHERE is_deleted = 0 AND COALESCE(is_bot, FALSE) = FALSE`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// GetOnlineUserIDs retrieves online user IDs
func (r *AdminRepository) GetOnlineUserIDs(ctx context.Context) ([]int64, error) {
	// Users online in last 5 minutes
	query := `SELECT id FROM users WHERE is_deleted = 0 AND COALESCE(is_bot, FALSE) = FALSE AND last_online > ?`
	rows, err := r.db.QueryContext(ctx, query, time.Now().Unix()-300)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

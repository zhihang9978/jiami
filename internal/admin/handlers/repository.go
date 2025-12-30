package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/feiji/feiji-backend/internal/admin/models"
)

// adminRepository handles admin panel database operations
type adminRepository struct {
	db *sql.DB
}

// newAdminRepository creates a new admin repository
func newAdminRepository(db *sql.DB) *adminRepository {
	return &adminRepository{db: db}
}

// GetAdminByUsername retrieves an admin by username
func (r *adminRepository) GetAdminByUsername(ctx context.Context, username string) (*models.Admin, error) {
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
func (r *adminRepository) GetAdminByID(ctx context.Context, id int64) (*models.Admin, error) {
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
func (r *adminRepository) UpdateAdminLogin(ctx context.Context, adminID int64, ip string) error {
	query := `UPDATE admins SET last_login_at = NOW(), last_login_ip = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, ip, adminID)
	return err
}

// UpdateAdminPassword updates admin password
func (r *adminRepository) UpdateAdminPassword(ctx context.Context, adminID int64, passwordHash string) error {
	query := `UPDATE admins SET password_hash = ?, updated_at = NOW() WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, passwordHash, adminID)
	return err
}

// CreateLoginLog creates a login log entry
func (r *adminRepository) CreateLoginLog(ctx context.Context, log *models.LoginLog) error {
	query := `INSERT INTO login_logs (admin_id, ip_address, user_agent, device, browser, status) 
			  VALUES (?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, log.AdminID, log.IPAddress, log.UserAgent, log.Device, log.Browser, log.Status)
	return err
}

// GetLoginLogs retrieves login logs for an admin
func (r *adminRepository) GetLoginLogs(ctx context.Context, adminID int64, page, pageSize int) ([]*models.LoginLog, int64, error) {
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
func (r *adminRepository) CreateAuditLog(ctx context.Context, log *models.AuditLog) error {
	query := `INSERT INTO audit_logs (admin_id, module, action, target_type, target_id, ip_address, user_agent, details) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, log.AdminID, log.Module, log.Action, log.TargetType, log.TargetID, log.IPAddress, log.UserAgent, log.Details)
	return err
}

// GetAuditLogs retrieves audit logs with pagination
func (r *adminRepository) GetAuditLogs(ctx context.Context, params *models.PaginationParams) ([]*models.AuditLog, int64, error) {
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
func (r *adminRepository) CreateNotification(ctx context.Context, notification *models.AdminNotification) error {
	query := `INSERT INTO admin_notifications (admin_id, category, title, message, data, priority, link) 
			  VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, notification.AdminID, notification.Category, notification.Title, notification.Message, notification.Data, notification.Priority, notification.Link)
	return err
}

// GetNotifications retrieves notifications for an admin
func (r *adminRepository) GetNotifications(ctx context.Context, adminID int64, isRead *bool, category string, page, pageSize int) ([]*models.AdminNotification, int64, int64, error) {
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
func (r *adminRepository) MarkNotificationRead(ctx context.Context, notificationID int64) error {
	query := `UPDATE admin_notifications SET is_read = TRUE WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, notificationID)
	return err
}

// MarkAllNotificationsRead marks all notifications as read for an admin
func (r *adminRepository) MarkAllNotificationsRead(ctx context.Context, adminID int64) error {
	query := `UPDATE admin_notifications SET is_read = TRUE WHERE (admin_id IS NULL OR admin_id = ?) AND is_read = FALSE`
	_, err := r.db.ExecContext(ctx, query, adminID)
	return err
}

// DeleteNotification deletes a notification
func (r *adminRepository) DeleteNotification(ctx context.Context, notificationID int64) error {
	query := `DELETE FROM admin_notifications WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, notificationID)
	return err
}

// GetUsersWithAdmin retrieves users with admin-specific fields
func (r *adminRepository) GetUsersWithAdmin(ctx context.Context, params *models.PaginationParams) ([]*models.UserWithAdmin, int64, error) {
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
func (r *adminRepository) GetUserByIDWithAdmin(ctx context.Context, userID int64) (*models.UserWithAdmin, error) {
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
func (r *adminRepository) CreateUserAdmin(ctx context.Context, req *models.CreateUserRequest) (int64, error) {
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
func (r *adminRepository) UpdateUserAdmin(ctx context.Context, userID int64, req *models.UpdateUserRequest) error {
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
func (r *adminRepository) DeleteUserAdmin(ctx context.Context, userID int64) error {
	query := `UPDATE users SET deleted_at = NOW(), status = 'deleted', is_deleted = 1 WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

// BatchUpdateUsersAdmin batch updates users from admin panel
func (r *adminRepository) BatchUpdateUsersAdmin(ctx context.Context, userIDs []int64, operation string) error {
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
		query = fmt.Sprintf("UPDATE users SET status = 'disabled', updated_at = NOW() WHERE id IN (%s)", strings.Join(placeholders, ","))
	case "enable":
		query = fmt.Sprintf("UPDATE users SET status = 'active', updated_at = NOW() WHERE id IN (%s)", strings.Join(placeholders, ","))
	case "delete":
		query = fmt.Sprintf("UPDATE users SET deleted_at = NOW(), status = 'deleted', is_deleted = 1 WHERE id IN (%s)", strings.Join(placeholders, ","))
	default:
		return fmt.Errorf("unknown operation: %s", operation)
	}
	
	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

// GetCallsAdmin retrieves calls with pagination
func (r *adminRepository) GetCallsAdmin(ctx context.Context, params *models.PaginationParams) ([]*models.CallRecord, int64, error) {
	offset := (params.Page - 1) * params.PageSize
	
	whereClause := "1=1"
	args := []interface{}{}
	
	if params.Search != "" {
		whereClause += " AND (caller_id = ? OR callee_id = ?)"
		args = append(args, params.Search, params.Search)
	}
	if params.Status != "" {
		whereClause += " AND status = ?"
		args = append(args, params.Status)
	}
	
	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM calls WHERE %s", whereClause)
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	
	query := fmt.Sprintf(`SELECT id, caller_id, callee_id, call_type, status, COALESCE(duration, 0), 
			  started_at, ended_at, created_at
			  FROM calls WHERE %s ORDER BY created_at DESC LIMIT ? OFFSET ?`, whereClause)
	args = append(args, params.PageSize, offset)
	
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	var calls []*models.CallRecord
	for rows.Next() {
		var c models.CallRecord
		if err := rows.Scan(&c.ID, &c.CallerID, &c.CalleeID, &c.CallType, &c.Status, &c.Duration,
			&c.StartedAt, &c.EndedAt, &c.CreatedAt); err != nil {
			return nil, 0, err
		}
		calls = append(calls, &c)
	}
	return calls, total, nil
}

// GetCallByIDAdmin retrieves a call by ID
func (r *adminRepository) GetCallByIDAdmin(ctx context.Context, callID int64) (*models.CallRecord, error) {
	query := `SELECT id, caller_id, callee_id, call_type, status, COALESCE(duration, 0), 
			  started_at, ended_at, created_at
			  FROM calls WHERE id = ?`
	
	var c models.CallRecord
	err := r.db.QueryRowContext(ctx, query, callID).Scan(&c.ID, &c.CallerID, &c.CalleeID, &c.CallType, &c.Status, &c.Duration,
		&c.StartedAt, &c.EndedAt, &c.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// GetBroadcasts retrieves broadcasts with pagination
func (r *adminRepository) GetBroadcasts(ctx context.Context, params *models.PaginationParams) ([]*models.BroadcastMessage, int64, error) {
	offset := (params.Page - 1) * params.PageSize
	
	whereClause := "1=1"
	args := []interface{}{}
	
	if params.Search != "" {
		whereClause += " AND (title LIKE ? OR content LIKE ?)"
		args = append(args, "%"+params.Search+"%", "%"+params.Search+"%")
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
	
	query := fmt.Sprintf(`SELECT id, admin_id, title, content, message_type, target_type, status, 
			  total_count, success_count, failed_count, created_at, completed_at
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
		if err := rows.Scan(&b.ID, &b.AdminID, &b.Title, &b.Content, &b.MessageType, &b.TargetType, &b.Status,
			&b.TotalCount, &b.SuccessCount, &b.FailedCount, &b.CreatedAt, &b.CompletedAt); err != nil {
			return nil, 0, err
		}
		broadcasts = append(broadcasts, &b)
	}
	return broadcasts, total, nil
}

// GetBroadcastByID retrieves a broadcast by ID
func (r *adminRepository) GetBroadcastByID(ctx context.Context, broadcastID int64) (*models.BroadcastMessage, error) {
	query := `SELECT id, admin_id, title, content, message_type, target_type, status, 
			  total_count, success_count, failed_count, created_at, completed_at
			  FROM broadcast_messages WHERE id = ?`
	
	var b models.BroadcastMessage
	err := r.db.QueryRowContext(ctx, query, broadcastID).Scan(&b.ID, &b.AdminID, &b.Title, &b.Content, &b.MessageType, &b.TargetType, &b.Status,
		&b.TotalCount, &b.SuccessCount, &b.FailedCount, &b.CreatedAt, &b.CompletedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &b, nil
}

// CreateBroadcast creates a new broadcast
func (r *adminRepository) CreateBroadcast(ctx context.Context, adminID int64, req *models.SendBroadcastRequest) (int64, error) {
	query := `INSERT INTO broadcast_messages (admin_id, title, content, message_type, target_type, status, created_at) 
			  VALUES (?, ?, ?, ?, ?, 'sending', NOW())`
	
	result, err := r.db.ExecContext(ctx, query, adminID, req.Title, req.Content, req.MessageType, req.TargetType)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// UpdateBroadcastStatus updates broadcast status
func (r *adminRepository) UpdateBroadcastStatus(ctx context.Context, broadcastID int64, status string, totalCount, successCount, failedCount int) error {
	query := `UPDATE broadcast_messages SET status = ?, total_count = ?, success_count = ?, failed_count = ?, completed_at = NOW() WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, status, totalCount, successCount, failedCount, broadcastID)
	return err
}

// CreateBroadcastDetail creates a broadcast detail
func (r *adminRepository) CreateBroadcastDetail(ctx context.Context, detail *models.BroadcastDetail) error {
	query := `INSERT INTO broadcast_details (broadcast_id, user_id, status, sent_at) VALUES (?, ?, ?, NOW())`
	_, err := r.db.ExecContext(ctx, query, detail.BroadcastID, detail.UserID, detail.Status)
	return err
}

// GetBroadcastDetails retrieves broadcast details
func (r *adminRepository) GetBroadcastDetails(ctx context.Context, broadcastID int64, status string, page, pageSize int) ([]*models.BroadcastDetail, int64, error) {
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
	
	query := fmt.Sprintf(`SELECT id, broadcast_id, user_id, status, sent_at, error_message
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
		if err := rows.Scan(&d.ID, &d.BroadcastID, &d.UserID, &d.Status, &d.SentAt, &d.ErrorMessage); err != nil {
			return nil, 0, err
		}
		details = append(details, &d)
	}
	return details, total, nil
}

// GetMessageTemplates retrieves message templates
func (r *adminRepository) GetMessageTemplates(ctx context.Context) ([]*models.MessageTemplate, error) {
	query := `SELECT id, name, content, template_type, variables, usage_count, created_at, updated_at
			  FROM message_templates ORDER BY created_at DESC`
	
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var templates []*models.MessageTemplate
	for rows.Next() {
		var t models.MessageTemplate
		var variables sql.NullString
		if err := rows.Scan(&t.ID, &t.Name, &t.Content, &t.TemplateType, &variables, &t.UsageCount, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		if variables.Valid {
			json.Unmarshal([]byte(variables.String), &t.Variables)
		}
		templates = append(templates, &t)
	}
	return templates, nil
}

// CreateMessageTemplate creates a message template
func (r *adminRepository) CreateMessageTemplate(ctx context.Context, req *models.CreateTemplateRequest) (int64, error) {
	variablesJSON, _ := json.Marshal(req.Variables)
	query := `INSERT INTO message_templates (name, content, template_type, variables, created_at, updated_at) 
			  VALUES (?, ?, ?, ?, NOW(), NOW())`
	result, err := r.db.ExecContext(ctx, query, req.Name, req.Content, req.TemplateType, string(variablesJSON))
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// UpdateMessageTemplate updates a message template
func (r *adminRepository) UpdateMessageTemplate(ctx context.Context, templateID int64, req *models.CreateTemplateRequest) error {
	variablesJSON, _ := json.Marshal(req.Variables)
	query := `UPDATE message_templates SET name = ?, content = ?, template_type = ?, variables = ?, updated_at = NOW() WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, req.Name, req.Content, req.TemplateType, string(variablesJSON), templateID)
	return err
}

// DeleteMessageTemplate deletes a message template
func (r *adminRepository) DeleteMessageTemplate(ctx context.Context, templateID int64) error {
	query := `DELETE FROM message_templates WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, templateID)
	return err
}

// GetAutoMessageConfigs retrieves auto message configs
func (r *adminRepository) GetAutoMessageConfigs(ctx context.Context) ([]*models.AutoMessageConfig, error) {
	query := `SELECT id, config_type, is_enabled, template_id, delay_seconds, created_at, updated_at
			  FROM auto_message_configs ORDER BY config_type`
	
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var configs []*models.AutoMessageConfig
	for rows.Next() {
		var c models.AutoMessageConfig
		if err := rows.Scan(&c.ID, &c.ConfigType, &c.IsEnabled, &c.TemplateID, &c.DelaySeconds, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		configs = append(configs, &c)
	}
	return configs, nil
}

// UpdateAutoMessageConfig updates an auto message config
func (r *adminRepository) UpdateAutoMessageConfig(ctx context.Context, configType string, req *models.UpdateAutoMessageRequest) error {
	query := `UPDATE auto_message_configs SET is_enabled = ?, template_id = ?, delay_seconds = ?, updated_at = NOW() WHERE config_type = ?`
	result, err := r.db.ExecContext(ctx, query, req.IsEnabled, req.TemplateID, req.DelaySeconds, configType)
	if err != nil {
		return err
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		insertQuery := `INSERT INTO auto_message_configs (config_type, is_enabled, template_id, delay_seconds, created_at, updated_at) 
					   VALUES (?, ?, ?, ?, NOW(), NOW())`
		_, err = r.db.ExecContext(ctx, insertQuery, configType, req.IsEnabled, req.TemplateID, req.DelaySeconds)
	}
	return err
}

// GetSystemConfigs retrieves system configs
func (r *adminRepository) GetSystemConfigs(ctx context.Context) ([]*models.SystemConfig, error) {
	query := `SELECT id, config_key, config_value, description, created_at, updated_at
			  FROM system_configs ORDER BY config_key`
	
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var configs []*models.SystemConfig
	for rows.Next() {
		var c models.SystemConfig
		if err := rows.Scan(&c.ID, &c.ConfigKey, &c.ConfigValue, &c.Description, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		configs = append(configs, &c)
	}
	return configs, nil
}

// UpdateSystemConfig updates a system config
func (r *adminRepository) UpdateSystemConfig(ctx context.Context, configKey, configValue string) error {
	query := `UPDATE system_configs SET config_value = ?, updated_at = NOW() WHERE config_key = ?`
	result, err := r.db.ExecContext(ctx, query, configValue, configKey)
	if err != nil {
		return err
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		insertQuery := `INSERT INTO system_configs (config_key, config_value, created_at, updated_at) VALUES (?, ?, NOW(), NOW())`
		_, err = r.db.ExecContext(ctx, insertQuery, configKey, configValue)
	}
	return err
}

// GetUserStatistics retrieves user statistics
func (r *adminRepository) GetUserStatistics(ctx context.Context) (*models.UserStatistics, error) {
	stats := &models.UserStatistics{}
	
	r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE deleted_at IS NULL OR deleted_at = '0000-00-00 00:00:00'").Scan(&stats.TotalUsers)
	r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE status = 'active' AND (deleted_at IS NULL OR deleted_at = '0000-00-00 00:00:00')").Scan(&stats.ActiveUsers)
	r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE DATE(created_at) = CURDATE()").Scan(&stats.NewUsersToday)
	r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE created_at >= DATE_SUB(NOW(), INTERVAL 7 DAY)").Scan(&stats.NewUsersWeek)
	r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE created_at >= DATE_SUB(NOW(), INTERVAL 30 DAY)").Scan(&stats.NewUsersMonth)
	
	return stats, nil
}

// GetCallStatistics retrieves call statistics
func (r *adminRepository) GetCallStatistics(ctx context.Context) (*models.CallStatistics, error) {
	stats := &models.CallStatistics{}
	
	r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM calls").Scan(&stats.TotalCalls)
	r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM calls WHERE DATE(created_at) = CURDATE()").Scan(&stats.CallsToday)
	r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM calls WHERE created_at >= DATE_SUB(NOW(), INTERVAL 7 DAY)").Scan(&stats.CallsWeek)
	r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM calls WHERE created_at >= DATE_SUB(NOW(), INTERVAL 30 DAY)").Scan(&stats.CallsMonth)
	r.db.QueryRowContext(ctx, "SELECT COALESCE(AVG(duration), 0) FROM calls WHERE duration > 0").Scan(&stats.AvgDuration)
	r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM calls WHERE call_type = 'video'").Scan(&stats.VideoCalls)
	r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM calls WHERE call_type = 'audio'").Scan(&stats.AudioCalls)
	
	return stats, nil
}

// GetBotStatistics retrieves bot statistics
func (r *adminRepository) GetBotStatistics(ctx context.Context) (*models.BotStatistics, error) {
	stats := &models.BotStatistics{}
	
	r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM broadcast_messages").Scan(&stats.TotalBroadcasts)
	r.db.QueryRowContext(ctx, "SELECT COALESCE(SUM(total_count), 0) FROM broadcast_messages").Scan(&stats.TotalMessagesSent)
	r.db.QueryRowContext(ctx, "SELECT COALESCE(SUM(success_count), 0) FROM broadcast_messages").Scan(&stats.SuccessfulDeliveries)
	r.db.QueryRowContext(ctx, "SELECT COALESCE(SUM(failed_count), 0) FROM broadcast_messages").Scan(&stats.FailedDeliveries)
	
	return stats, nil
}

// GetDashboardData retrieves dashboard data
func (r *adminRepository) GetDashboardData(ctx context.Context) (*models.DashboardData, error) {
	data := &models.DashboardData{}
	
	r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE deleted_at IS NULL OR deleted_at = '0000-00-00 00:00:00'").Scan(&data.TotalUsers)
	r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE DATE(created_at) = CURDATE()").Scan(&data.NewUsersToday)
	r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM calls").Scan(&data.TotalCalls)
	r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM calls WHERE DATE(created_at) = CURDATE()").Scan(&data.CallsToday)
	r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM broadcast_messages").Scan(&data.TotalBroadcasts)
	
	return data, nil
}

// GetAllUserIDs retrieves all user IDs
func (r *adminRepository) GetAllUserIDs(ctx context.Context) ([]int64, error) {
	query := `SELECT id FROM users WHERE deleted_at IS NULL OR deleted_at = '0000-00-00 00:00:00'`
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
func (r *adminRepository) GetOnlineUserIDs(ctx context.Context) ([]int64, error) {
	query := `SELECT id FROM users WHERE status = 'online' AND (deleted_at IS NULL OR deleted_at = '0000-00-00 00:00:00')`
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

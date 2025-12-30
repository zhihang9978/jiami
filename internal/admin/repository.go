package admin

import (
	"context"
	"database/sql"
	"time"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Stats represents system statistics
type Stats struct {
	TotalUsers       int64
	ActiveUsers      int64
	TotalMessages    int64
	TotalChats       int64
	TotalChannels    int64
	OnlineUsers      int64
	NewUsersToday    int64
	MessagesToday    int64
}

// UserInfo represents admin view of user
type UserInfo struct {
	ID          int64
	Phone       string
	Username    *string
	FirstName   string
	LastName    string
	Bio         string
	PhotoID     *int64
	IsBot       bool
	IsVerified  bool
	IsRestricted bool
	IsDeleted   bool
	LastOnline  int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// GetStats retrieves system statistics
func (r *Repository) GetStats(ctx context.Context) (*Stats, error) {
	stats := &Stats{}

	// Total users
	r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE is_deleted = 0`).Scan(&stats.TotalUsers)

	// Active users (last 30 days)
	r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE is_deleted = 0 AND last_online > ?`, time.Now().Unix()-30*24*3600).Scan(&stats.ActiveUsers)

	// Total messages
	r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM messages`).Scan(&stats.TotalMessages)

	// Total chats
	r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM chats WHERE is_deactivated = 0`).Scan(&stats.TotalChats)

	// Total channels
	r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM channels`).Scan(&stats.TotalChannels)

	// Online users (last 5 minutes)
	r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE is_deleted = 0 AND last_online > ?`, time.Now().Unix()-300).Scan(&stats.OnlineUsers)

	// New users today
	today := time.Now().Truncate(24 * time.Hour)
	r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE created_at >= ?`, today).Scan(&stats.NewUsersToday)

	// Messages today
	r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM messages WHERE date >= ?`, today.Unix()).Scan(&stats.MessagesToday)

	return stats, nil
}

// GetUsers retrieves users with pagination
func (r *Repository) GetUsers(ctx context.Context, offset, limit int, search string) ([]*UserInfo, int64, error) {
	var total int64
	countQuery := `SELECT COUNT(*) FROM users WHERE is_deleted = 0`
	args := []interface{}{}

	if search != "" {
		countQuery += ` AND (phone LIKE ? OR username LIKE ? OR first_name LIKE ? OR last_name LIKE ?)`
		searchPattern := "%" + search + "%"
		args = append(args, searchPattern, searchPattern, searchPattern, searchPattern)
	}

	r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)

	query := `SELECT id, phone, username, COALESCE(first_name, ''), COALESCE(last_name, ''), COALESCE(bio, ''),
	                 photo_id, is_bot, is_verified, is_restricted, is_deleted, last_online, created_at, updated_at
	          FROM users WHERE is_deleted = 0`

	if search != "" {
		query += ` AND (phone LIKE ? OR username LIKE ? OR first_name LIKE ? OR last_name LIKE ?)`
	}
	query += ` ORDER BY created_at DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []*UserInfo
	for rows.Next() {
		u := &UserInfo{}
		if err := rows.Scan(&u.ID, &u.Phone, &u.Username, &u.FirstName, &u.LastName, &u.Bio,
			&u.PhotoID, &u.IsBot, &u.IsVerified, &u.IsRestricted, &u.IsDeleted, &u.LastOnline,
			&u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}

	return users, total, nil
}

// GetUser retrieves a single user by ID
func (r *Repository) GetUser(ctx context.Context, userID int64) (*UserInfo, error) {
	u := &UserInfo{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, phone, username, COALESCE(first_name, ''), COALESCE(last_name, ''), COALESCE(bio, ''),
		       photo_id, is_bot, is_verified, is_restricted, is_deleted, last_online, created_at, updated_at
		FROM users WHERE id = ?
	`, userID).Scan(&u.ID, &u.Phone, &u.Username, &u.FirstName, &u.LastName, &u.Bio,
		&u.PhotoID, &u.IsBot, &u.IsVerified, &u.IsRestricted, &u.IsDeleted, &u.LastOnline,
		&u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

// CreateUser creates a new user (admin function)
func (r *Repository) CreateUser(ctx context.Context, phone, firstName, lastName string, username *string) (*UserInfo, error) {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO users (phone, first_name, last_name, username, created_at, updated_at)
		VALUES (?, ?, ?, ?, NOW(), NOW())
	`, phone, firstName, lastName, username)
	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()
	return r.GetUser(ctx, id)
}

// UpdateUser updates a user
func (r *Repository) UpdateUser(ctx context.Context, userID int64, firstName, lastName, bio string, username *string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE users SET first_name = ?, last_name = ?, bio = ?, username = ?, updated_at = NOW()
		WHERE id = ?
	`, firstName, lastName, bio, username, userID)
	return err
}

// DeleteUser soft deletes a user
func (r *Repository) DeleteUser(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET is_deleted = 1, updated_at = NOW() WHERE id = ?`, userID)
	return err
}

// BanUser bans a user
func (r *Repository) BanUser(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET is_restricted = 1, updated_at = NOW() WHERE id = ?`, userID)
	return err
}

// UnbanUser unbans a user
func (r *Repository) UnbanUser(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET is_restricted = 0, updated_at = NOW() WHERE id = ?`, userID)
	return err
}

// GetMessages retrieves messages with pagination
func (r *Repository) GetMessages(ctx context.Context, offset, limit int, userID int64) ([]map[string]interface{}, int64, error) {
	var total int64
	countQuery := `SELECT COUNT(*) FROM messages`
	args := []interface{}{}

	if userID > 0 {
		countQuery += ` WHERE from_id = ?`
		args = append(args, userID)
	}

	r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)

	query := `SELECT id, from_id, peer_id, peer_type, message, date, is_out, is_mentioned, is_media_unread, is_silent, is_post
	          FROM messages`
	if userID > 0 {
		query += ` WHERE from_id = ?`
	}
	query += ` ORDER BY date DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var messages []map[string]interface{}
	for rows.Next() {
		var id, fromID, peerID int64
		var peerType, message string
		var date int
		var isOut, isMentioned, isMediaUnread, isSilent, isPost bool

		if err := rows.Scan(&id, &fromID, &peerID, &peerType, &message, &date, &isOut, &isMentioned, &isMediaUnread, &isSilent, &isPost); err != nil {
			return nil, 0, err
		}

		messages = append(messages, map[string]interface{}{
			"id":              id,
			"from_id":         fromID,
			"peer_id":         peerID,
			"peer_type":       peerType,
			"message":         message,
			"date":            date,
			"is_out":          isOut,
			"is_mentioned":    isMentioned,
			"is_media_unread": isMediaUnread,
			"is_silent":       isSilent,
			"is_post":         isPost,
		})
	}

	return messages, total, nil
}

// DeleteMessage deletes a message
func (r *Repository) DeleteMessage(ctx context.Context, messageID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM messages WHERE id = ?`, messageID)
	return err
}

// GetChats retrieves chats with pagination
func (r *Repository) GetChats(ctx context.Context, offset, limit int) ([]map[string]interface{}, int64, error) {
	var total int64
	r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM chats WHERE is_deactivated = 0`).Scan(&total)

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, title, participants_count, date, creator_id, is_deactivated, created_at
		FROM chats WHERE is_deactivated = 0
		ORDER BY created_at DESC LIMIT ? OFFSET ?
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var chats []map[string]interface{}
	for rows.Next() {
		var id, creatorID int64
		var title string
		var participantsCount, date int
		var isDeactivated bool
		var createdAt time.Time

		if err := rows.Scan(&id, &title, &participantsCount, &date, &creatorID, &isDeactivated, &createdAt); err != nil {
			return nil, 0, err
		}

		chats = append(chats, map[string]interface{}{
			"id":                 id,
			"title":              title,
			"participants_count": participantsCount,
			"date":               date,
			"creator_id":         creatorID,
			"is_deactivated":     isDeactivated,
			"created_at":         createdAt,
		})
	}

	return chats, total, nil
}

// GetChannels retrieves channels with pagination
func (r *Repository) GetChannels(ctx context.Context, offset, limit int) ([]map[string]interface{}, int64, error) {
	var total int64
	r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM channels`).Scan(&total)

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, title, username, participants_count, date, creator_id, is_broadcast, is_megagroup, created_at
		FROM channels
		ORDER BY created_at DESC LIMIT ? OFFSET ?
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var channels []map[string]interface{}
	for rows.Next() {
		var id, creatorID int64
		var title string
		var username *string
		var participantsCount, date int
		var isBroadcast, isMegagroup bool
		var createdAt time.Time

		if err := rows.Scan(&id, &title, &username, &participantsCount, &date, &creatorID, &isBroadcast, &isMegagroup, &createdAt); err != nil {
			return nil, 0, err
		}

		channels = append(channels, map[string]interface{}{
			"id":                 id,
			"title":              title,
			"username":           username,
			"participants_count": participantsCount,
			"date":               date,
			"creator_id":         creatorID,
			"is_broadcast":       isBroadcast,
			"is_megagroup":       isMegagroup,
			"created_at":         createdAt,
		})
	}

	return channels, total, nil
}

// GetSessions retrieves active sessions
func (r *Repository) GetSessions(ctx context.Context, userID int64) ([]map[string]interface{}, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, auth_key_id, device_model, platform, system_version, app_version, ip, created_at
		FROM sessions WHERE user_id = ?
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []map[string]interface{}
	for rows.Next() {
		var id, sessionUserID int64
		var authKeyID, deviceModel, platform, systemVersion, appVersion, ip string
		var createdAt time.Time

		if err := rows.Scan(&id, &sessionUserID, &authKeyID, &deviceModel, &platform, &systemVersion, &appVersion, &ip, &createdAt); err != nil {
			return nil, err
		}

		sessions = append(sessions, map[string]interface{}{
			"id":             id,
			"user_id":        sessionUserID,
			"auth_key_id":    authKeyID,
			"device_model":   deviceModel,
			"platform":       platform,
			"system_version": systemVersion,
			"app_version":    appVersion,
			"ip":             ip,
			"created_at":     createdAt,
		})
	}

	return sessions, nil
}

// TerminateSession terminates a session
func (r *Repository) TerminateSession(ctx context.Context, sessionID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM sessions WHERE id = ?`, sessionID)
	return err
}

// TerminateAllSessions terminates all sessions for a user
func (r *Repository) TerminateAllSessions(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM sessions WHERE user_id = ?`, userID)
	return err
}

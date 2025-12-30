package chats

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"
)

type Chat struct {
	ID                 int64
	Title              string
	PhotoID            *int64
	ParticipantsCount  int
	Date               int
	Version            int
	CreatorID          int64
	IsDeactivated      bool
	IsCallActive       bool
	IsCallNotEmpty     bool
	MigratedToChannelID *int64
	AdminRights        json.RawMessage
	DefaultBannedRights json.RawMessage
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type ChatParticipant struct {
	ID          int64
	ChatID      int64
	UserID      int64
	InviterID   *int64
	Date        int
	IsAdmin     bool
	AdminRights json.RawMessage
	CreatedAt   time.Time
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// CreateChat creates a new chat group
func (r *Repository) CreateChat(ctx context.Context, chat *Chat) error {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO chats (title, photo_id, participants_count, date, version, creator_id, 
		                   is_deactivated, admin_rights, default_banned_rights, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`, chat.Title, chat.PhotoID, chat.ParticipantsCount, chat.Date, chat.Version, chat.CreatorID,
		chat.IsDeactivated, chat.AdminRights, chat.DefaultBannedRights)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	chat.ID = id
	return nil
}

// GetChat retrieves a chat by ID
func (r *Repository) GetChat(ctx context.Context, chatID int64) (*Chat, error) {
	chat := &Chat{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, title, photo_id, participants_count, date, version, creator_id,
		       is_deactivated, is_call_active, is_call_not_empty, migrated_to_channel_id,
		       admin_rights, default_banned_rights, created_at, updated_at
		FROM chats
		WHERE id = ?
	`, chatID).Scan(&chat.ID, &chat.Title, &chat.PhotoID, &chat.ParticipantsCount, &chat.Date,
		&chat.Version, &chat.CreatorID, &chat.IsDeactivated, &chat.IsCallActive, &chat.IsCallNotEmpty,
		&chat.MigratedToChannelID, &chat.AdminRights, &chat.DefaultBannedRights, &chat.CreatedAt, &chat.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return chat, nil
}

// UpdateChat updates a chat
func (r *Repository) UpdateChat(ctx context.Context, chat *Chat) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE chats
		SET title = ?, photo_id = ?, participants_count = ?, version = version + 1,
		    is_deactivated = ?, admin_rights = ?, default_banned_rights = ?, updated_at = NOW()
		WHERE id = ?
	`, chat.Title, chat.PhotoID, chat.ParticipantsCount, chat.IsDeactivated,
		chat.AdminRights, chat.DefaultBannedRights, chat.ID)
	return err
}

// DeleteChat deletes a chat
func (r *Repository) DeleteChat(ctx context.Context, chatID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM chats WHERE id = ?`, chatID)
	return err
}

// AddParticipant adds a user to a chat
func (r *Repository) AddParticipant(ctx context.Context, participant *ChatParticipant) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO chat_participants (chat_id, user_id, inviter_id, date, is_admin, admin_rights, created_at)
		VALUES (?, ?, ?, ?, ?, ?, NOW())
		ON DUPLICATE KEY UPDATE inviter_id = VALUES(inviter_id), date = VALUES(date)
	`, participant.ChatID, participant.UserID, participant.InviterID, participant.Date,
		participant.IsAdmin, participant.AdminRights)
	return err
}

// RemoveParticipant removes a user from a chat
func (r *Repository) RemoveParticipant(ctx context.Context, chatID, userID int64) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM chat_participants WHERE chat_id = ? AND user_id = ?
	`, chatID, userID)
	return err
}

// GetParticipant retrieves a participant
func (r *Repository) GetParticipant(ctx context.Context, chatID, userID int64) (*ChatParticipant, error) {
	p := &ChatParticipant{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, chat_id, user_id, inviter_id, date, is_admin, admin_rights, created_at
		FROM chat_participants
		WHERE chat_id = ? AND user_id = ?
	`, chatID, userID).Scan(&p.ID, &p.ChatID, &p.UserID, &p.InviterID, &p.Date, &p.IsAdmin, &p.AdminRights, &p.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return p, nil
}

// GetParticipants retrieves all participants of a chat
func (r *Repository) GetParticipants(ctx context.Context, chatID int64) ([]*ChatParticipant, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, chat_id, user_id, inviter_id, date, is_admin, admin_rights, created_at
		FROM chat_participants
		WHERE chat_id = ?
		ORDER BY date
	`, chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var participants []*ChatParticipant
	for rows.Next() {
		p := &ChatParticipant{}
		if err := rows.Scan(&p.ID, &p.ChatID, &p.UserID, &p.InviterID, &p.Date, &p.IsAdmin, &p.AdminRights, &p.CreatedAt); err != nil {
			return nil, err
		}
		participants = append(participants, p)
	}
	return participants, nil
}

// UpdateParticipantAdmin updates admin status
func (r *Repository) UpdateParticipantAdmin(ctx context.Context, chatID, userID int64, isAdmin bool, adminRights json.RawMessage) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE chat_participants
		SET is_admin = ?, admin_rights = ?
		WHERE chat_id = ? AND user_id = ?
	`, isAdmin, adminRights, chatID, userID)
	return err
}

// IncrementParticipantsCount increments the participants count
func (r *Repository) IncrementParticipantsCount(ctx context.Context, chatID int64, delta int) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE chats SET participants_count = participants_count + ?, updated_at = NOW() WHERE id = ?
	`, delta, chatID)
	return err
}

// GetUserChats retrieves all chats a user is a member of
func (r *Repository) GetUserChats(ctx context.Context, userID int64) ([]*Chat, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT c.id, c.title, c.photo_id, c.participants_count, c.date, c.version, c.creator_id,
		       c.is_deactivated, c.is_call_active, c.is_call_not_empty, c.migrated_to_channel_id,
		       c.admin_rights, c.default_banned_rights, c.created_at, c.updated_at
		FROM chats c
		JOIN chat_participants cp ON c.id = cp.chat_id
		WHERE cp.user_id = ? AND c.is_deactivated = 0
		ORDER BY c.updated_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chats []*Chat
	for rows.Next() {
		chat := &Chat{}
		if err := rows.Scan(&chat.ID, &chat.Title, &chat.PhotoID, &chat.ParticipantsCount, &chat.Date,
			&chat.Version, &chat.CreatorID, &chat.IsDeactivated, &chat.IsCallActive, &chat.IsCallNotEmpty,
			&chat.MigratedToChannelID, &chat.AdminRights, &chat.DefaultBannedRights, &chat.CreatedAt, &chat.UpdatedAt); err != nil {
			return nil, err
		}
		chats = append(chats, chat)
	}
	return chats, nil
}

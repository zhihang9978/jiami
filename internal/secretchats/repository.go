package secretchats

import (
	"context"
	"database/sql"
	"time"
)

type SecretChat struct {
	ID          int64
	InitiatorID int64
	PeerID      int64
	Status      string // WAITING, ACTIVE, CLOSED
	GA          string // DH public key A (base64)
	GB          string // DH public key B (base64)
	KeyHash     string // Key fingerprint
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type SecretMessage struct {
	ID               int64
	SecretChatID     int64
	FromID           int64
	EncryptedMessage string // base64 encoded encrypted data
	Date             int64
	CreatedAt        time.Time
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// CreateSecretChat creates a new secret chat
func (r *Repository) CreateSecretChat(ctx context.Context, chat *SecretChat) error {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO secret_chats (initiator_id, peer_id, status, g_a, g_b, key_hash, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())
	`, chat.InitiatorID, chat.PeerID, chat.Status, chat.GA, chat.GB, chat.KeyHash)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	chat.ID = id
	return nil
}

// GetSecretChat retrieves a secret chat by ID
func (r *Repository) GetSecretChat(ctx context.Context, chatID int64) (*SecretChat, error) {
	chat := &SecretChat{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, initiator_id, peer_id, status, COALESCE(g_a, ''), COALESCE(g_b, ''), COALESCE(key_hash, ''), created_at, updated_at
		FROM secret_chats WHERE id = ?
	`, chatID).Scan(&chat.ID, &chat.InitiatorID, &chat.PeerID, &chat.Status, &chat.GA, &chat.GB, &chat.KeyHash, &chat.CreatedAt, &chat.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return chat, nil
}

// UpdateSecretChat updates a secret chat
func (r *Repository) UpdateSecretChat(ctx context.Context, chat *SecretChat) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE secret_chats SET status = ?, g_a = ?, g_b = ?, key_hash = ?, updated_at = NOW()
		WHERE id = ?
	`, chat.Status, chat.GA, chat.GB, chat.KeyHash, chat.ID)
	return err
}

// GetUserSecretChats retrieves all secret chats for a user
func (r *Repository) GetUserSecretChats(ctx context.Context, userID int64) ([]*SecretChat, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, initiator_id, peer_id, status, COALESCE(g_a, ''), COALESCE(g_b, ''), COALESCE(key_hash, ''), created_at, updated_at
		FROM secret_chats
		WHERE initiator_id = ? OR peer_id = ?
		ORDER BY updated_at DESC
	`, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chats []*SecretChat
	for rows.Next() {
		chat := &SecretChat{}
		if err := rows.Scan(&chat.ID, &chat.InitiatorID, &chat.PeerID, &chat.Status, &chat.GA, &chat.GB, &chat.KeyHash, &chat.CreatedAt, &chat.UpdatedAt); err != nil {
			return nil, err
		}
		chats = append(chats, chat)
	}
	return chats, nil
}

// DeleteSecretChat deletes a secret chat
func (r *Repository) DeleteSecretChat(ctx context.Context, chatID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM secret_chats WHERE id = ?`, chatID)
	return err
}

// CreateSecretMessage creates a new secret message
func (r *Repository) CreateSecretMessage(ctx context.Context, msg *SecretMessage) error {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO secret_messages (secret_chat_id, from_id, encrypted_message, date, created_at)
		VALUES (?, ?, ?, ?, NOW())
	`, msg.SecretChatID, msg.FromID, msg.EncryptedMessage, msg.Date)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	msg.ID = id
	return nil
}

// GetSecretMessages retrieves messages for a secret chat
func (r *Repository) GetSecretMessages(ctx context.Context, chatID int64, limit, offset int) ([]*SecretMessage, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, secret_chat_id, from_id, encrypted_message, date, created_at
		FROM secret_messages
		WHERE secret_chat_id = ?
		ORDER BY date DESC
		LIMIT ? OFFSET ?
	`, chatID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*SecretMessage
	for rows.Next() {
		msg := &SecretMessage{}
		if err := rows.Scan(&msg.ID, &msg.SecretChatID, &msg.FromID, &msg.EncryptedMessage, &msg.Date, &msg.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	return messages, nil
}

// DeleteSecretMessages deletes all messages for a secret chat
func (r *Repository) DeleteSecretMessages(ctx context.Context, chatID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM secret_messages WHERE secret_chat_id = ?`, chatID)
	return err
}

package messages

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/feiji/feiji-backend/internal/models"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// GetNextMessageID gets the next message ID for a peer
func (r *Repository) GetNextMessageID(ctx context.Context, peerID int64, peerType string) (int, error) {
	var maxID sql.NullInt32
	err := r.db.QueryRowContext(ctx, `
		SELECT MAX(message_id) FROM messages WHERE peer_id = ? AND peer_type = ?
	`, peerID, peerType).Scan(&maxID)
	if err != nil {
		return 1, nil
	}
	if maxID.Valid {
		return int(maxID.Int32) + 1, nil
	}
	return 1, nil
}

// CreateMessage creates a new message
func (r *Repository) CreateMessage(ctx context.Context, msg *models.Message) error {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO messages (message_id, from_id, peer_id, peer_type, message, date, random_id, 
		                      reply_to_msg_id, media_type, media_id, entities, is_out)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, msg.MessageID, msg.FromID, msg.PeerID, msg.PeerType, msg.Message, msg.Date, msg.RandomID,
		msg.ReplyToMsgID, msg.MediaType, msg.MediaID, msg.Entities, msg.IsOut)
	if err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}

	id, _ := result.LastInsertId()
	msg.ID = id
	return nil
}

// GetMessageByRandomID retrieves a message by random_id (for deduplication)
func (r *Repository) GetMessageByRandomID(ctx context.Context, randomID int64) (*models.Message, error) {
	msg := &models.Message{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, message_id, from_id, peer_id, peer_type, message, date, random_id,
		       reply_to_msg_id, media_type, media_id, entities, is_out, created_at
		FROM messages WHERE random_id = ?
	`, randomID).Scan(
		&msg.ID, &msg.MessageID, &msg.FromID, &msg.PeerID, &msg.PeerType, &msg.Message,
		&msg.Date, &msg.RandomID, &msg.ReplyToMsgID, &msg.MediaType, &msg.MediaID,
		&msg.Entities, &msg.IsOut, &msg.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return msg, nil
}

// GetHistory retrieves message history for a peer
func (r *Repository) GetHistory(ctx context.Context, userID, peerID int64, peerType string, offsetID, limit int) ([]*models.Message, error) {
	query := `
		SELECT id, message_id, from_id, peer_id, peer_type, message, date, random_id,
		       reply_to_msg_id, fwd_from_id, fwd_date, edit_date, media_type, media_id,
		       entities, is_out, is_mentioned, is_media_unread, is_silent, is_pinned, created_at
		FROM messages 
		WHERE ((peer_id = ? AND peer_type = ?) OR (from_id = ? AND peer_id = ? AND peer_type = 'user'))
	`
	args := []interface{}{peerID, peerType, userID, peerID}

	if offsetID > 0 {
		query += " AND message_id < ?"
		args = append(args, offsetID)
	}

	query += " ORDER BY message_id DESC LIMIT ?"
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get history: %w", err)
	}
	defer rows.Close()

	var messages []*models.Message
	for rows.Next() {
		msg := &models.Message{}
		err := rows.Scan(
			&msg.ID, &msg.MessageID, &msg.FromID, &msg.PeerID, &msg.PeerType, &msg.Message,
			&msg.Date, &msg.RandomID, &msg.ReplyToMsgID, &msg.FwdFromID, &msg.FwdDate,
			&msg.EditDate, &msg.MediaType, &msg.MediaID, &msg.Entities, &msg.IsOut,
			&msg.IsMentioned, &msg.IsMediaUnread, &msg.IsSilent, &msg.IsPinned, &msg.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

// UpdateOrCreateDialog updates or creates a dialog
func (r *Repository) UpdateOrCreateDialog(ctx context.Context, userID, peerID int64, peerType string, topMessage int) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO dialogs (user_id, peer_id, peer_type, top_message, unread_count)
		VALUES (?, ?, ?, ?, 0)
		ON DUPLICATE KEY UPDATE top_message = ?, updated_at = NOW()
	`, userID, peerID, peerType, topMessage, topMessage)
	return err
}

// IncrementUnreadCount increments the unread count for a dialog
func (r *Repository) IncrementUnreadCount(ctx context.Context, userID, peerID int64, peerType string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE dialogs SET unread_count = unread_count + 1, updated_at = NOW()
		WHERE user_id = ? AND peer_id = ? AND peer_type = ?
	`, userID, peerID, peerType)
	return err
}

// GetDialogs retrieves dialogs for a user
func (r *Repository) GetDialogs(ctx context.Context, userID int64, offsetDate, limit int) ([]*models.Dialog, error) {
	query := `
		SELECT id, user_id, peer_id, peer_type, top_message, read_inbox_max_id, read_outbox_max_id,
		       unread_count, unread_mentions_count, unread_reactions_count, pts, draft, folder_id,
		       is_pinned, pinned_order, mute_until, notify_sound, show_previews, updated_at
		FROM dialogs WHERE user_id = ?
	`
	args := []interface{}{userID}

	if offsetDate > 0 {
		query += " AND UNIX_TIMESTAMP(updated_at) < ?"
		args = append(args, offsetDate)
	}

	query += " ORDER BY is_pinned DESC, updated_at DESC LIMIT ?"
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get dialogs: %w", err)
	}
	defer rows.Close()

	var dialogs []*models.Dialog
	for rows.Next() {
		d := &models.Dialog{}
		err := rows.Scan(
			&d.ID, &d.UserID, &d.PeerID, &d.PeerType, &d.TopMessage, &d.ReadInboxMaxID,
			&d.ReadOutboxMaxID, &d.UnreadCount, &d.UnreadMentionsCount, &d.UnreadReactionsCount,
			&d.Pts, &d.Draft, &d.FolderID, &d.IsPinned, &d.PinnedOrder, &d.MuteUntil,
			&d.NotifySound, &d.ShowPreviews, &d.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		dialogs = append(dialogs, d)
	}

	return dialogs, nil
}

// MarkAsRead marks messages as read
func (r *Repository) MarkAsRead(ctx context.Context, userID, peerID int64, peerType string, maxID int) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE dialogs SET read_inbox_max_id = ?, unread_count = 0
		WHERE user_id = ? AND peer_id = ? AND peer_type = ? AND read_inbox_max_id < ?
	`, maxID, userID, peerID, peerType, maxID)
	return err
}

// DeleteMessage deletes a message
func (r *Repository) DeleteMessage(ctx context.Context, userID int64, messageIDs []int) error {
	if len(messageIDs) == 0 {
		return nil
	}

	// Build placeholders
	placeholders := ""
	args := make([]interface{}, len(messageIDs)+1)
	args[0] = userID
	for i, id := range messageIDs {
		if i > 0 {
			placeholders += ","
		}
		placeholders += "?"
		args[i+1] = id
	}

	_, err := r.db.ExecContext(ctx, fmt.Sprintf(`
		DELETE FROM messages WHERE from_id = ? AND message_id IN (%s)
	`, placeholders), args...)
	return err
}

// EditMessage edits a message
func (r *Repository) EditMessage(ctx context.Context, userID int64, messageID int, newMessage string, entities json.RawMessage) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE messages SET message = ?, entities = ?, edit_date = ?
		WHERE from_id = ? AND message_id = ?
	`, newMessage, entities, time.Now().Unix(), userID, messageID)
	return err
}

// GetMessage retrieves a single message
func (r *Repository) GetMessage(ctx context.Context, peerID int64, peerType string, messageID int) (*models.Message, error) {
	msg := &models.Message{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, message_id, from_id, peer_id, peer_type, message, date, random_id,
		       reply_to_msg_id, fwd_from_id, fwd_date, edit_date, media_type, media_id,
		       entities, is_out, is_mentioned, is_media_unread, is_silent, is_pinned, created_at
		FROM messages WHERE peer_id = ? AND peer_type = ? AND message_id = ?
	`, peerID, peerType, messageID).Scan(
		&msg.ID, &msg.MessageID, &msg.FromID, &msg.PeerID, &msg.PeerType, &msg.Message,
		&msg.Date, &msg.RandomID, &msg.ReplyToMsgID, &msg.FwdFromID, &msg.FwdDate,
		&msg.EditDate, &msg.MediaType, &msg.MediaID, &msg.Entities, &msg.IsOut,
		&msg.IsMentioned, &msg.IsMediaUnread, &msg.IsSilent, &msg.IsPinned, &msg.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return msg, nil
}

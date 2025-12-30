package messages

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/feiji/feiji-backend/internal/models"
	"github.com/feiji/feiji-backend/internal/store"
)

type Service struct {
	repo  *Repository
	redis *store.RedisStore
}

func NewService(repo *Repository, redis *store.RedisStore) *Service {
	return &Service{
		repo:  repo,
		redis: redis,
	}
}

// SendMessage sends a message to a peer
func (s *Service) SendMessage(ctx context.Context, fromID, peerID int64, peerType, message string, randomID int64, replyToMsgID *int) (*models.Message, error) {
	// Check for duplicate message (idempotency)
	if randomID != 0 {
		existing, err := s.repo.GetMessageByRandomID(ctx, randomID)
		if err != nil {
			return nil, fmt.Errorf("failed to check duplicate: %w", err)
		}
		if existing != nil {
			return existing, nil
		}
	}

	// Get next message ID
	messageID, err := s.repo.GetNextMessageID(ctx, peerID, peerType)
	if err != nil {
		return nil, fmt.Errorf("failed to get next message id: %w", err)
	}

	// Create message
	msg := &models.Message{
		MessageID: messageID,
		FromID:    fromID,
		PeerID:    peerID,
		PeerType:  peerType,
		Message:   sql.NullString{String: message, Valid: message != ""},
		Date:      int(time.Now().Unix()),
		RandomID:  sql.NullInt64{Int64: randomID, Valid: randomID != 0},
		IsOut:     true,
	}

	if replyToMsgID != nil {
		msg.ReplyToMsgID = sql.NullInt32{Int32: int32(*replyToMsgID), Valid: true}
	}

	if err := s.repo.CreateMessage(ctx, msg); err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	// Update sender's dialog
	if err := s.repo.UpdateOrCreateDialog(ctx, fromID, peerID, peerType, messageID); err != nil {
		return nil, fmt.Errorf("failed to update sender dialog: %w", err)
	}

	// Update recipient's dialog and increment unread count
	if peerType == "user" {
		if err := s.repo.UpdateOrCreateDialog(ctx, peerID, fromID, "user", messageID); err != nil {
			return nil, fmt.Errorf("failed to update recipient dialog: %w", err)
		}
		if err := s.repo.IncrementUnreadCount(ctx, peerID, fromID, "user"); err != nil {
			return nil, fmt.Errorf("failed to increment unread count: %w", err)
		}
	}

	// Increment PTS for the recipient
	if peerType == "user" {
		s.redis.IncrPts(ctx, peerID)
	}

	return msg, nil
}

// GetHistory retrieves message history
func (s *Service) GetHistory(ctx context.Context, userID, peerID int64, peerType string, offsetID, limit int) ([]*models.Message, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	messages, err := s.repo.GetHistory(ctx, userID, peerID, peerType, offsetID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get history: %w", err)
	}

	return messages, nil
}

// GetDialogs retrieves user's dialogs
func (s *Service) GetDialogs(ctx context.Context, userID int64, offsetDate, limit int) ([]*models.Dialog, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	dialogs, err := s.repo.GetDialogs(ctx, userID, offsetDate, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get dialogs: %w", err)
	}

	return dialogs, nil
}

// MarkAsRead marks messages as read
func (s *Service) MarkAsRead(ctx context.Context, userID, peerID int64, peerType string, maxID int) error {
	return s.repo.MarkAsRead(ctx, userID, peerID, peerType, maxID)
}

// DeleteMessages deletes messages
func (s *Service) DeleteMessages(ctx context.Context, userID int64, messageIDs []int, revoke bool) error {
	return s.repo.DeleteMessage(ctx, userID, messageIDs)
}

// EditMessage edits a message
func (s *Service) EditMessage(ctx context.Context, userID int64, peerID int64, peerType string, messageID int, newMessage string, entities json.RawMessage) (*models.Message, error) {
	if err := s.repo.EditMessage(ctx, userID, messageID, newMessage, entities); err != nil {
		return nil, fmt.Errorf("failed to edit message: %w", err)
	}

	return s.repo.GetMessage(ctx, peerID, peerType, messageID)
}

// ForwardMessages forwards messages
func (s *Service) ForwardMessages(ctx context.Context, fromID, fromPeerID int64, fromPeerType string, toPeerID int64, toPeerType string, messageIDs []int) ([]*models.Message, error) {
	var forwardedMessages []*models.Message

	for _, msgID := range messageIDs {
		// Get original message
		original, err := s.repo.GetMessage(ctx, fromPeerID, fromPeerType, msgID)
		if err != nil || original == nil {
			continue
		}

		// Get next message ID for destination
		newMsgID, err := s.repo.GetNextMessageID(ctx, toPeerID, toPeerType)
		if err != nil {
			continue
		}

		// Create forwarded message
		msg := &models.Message{
			MessageID: newMsgID,
			FromID:    fromID,
			PeerID:    toPeerID,
			PeerType:  toPeerType,
			Message:   original.Message,
			Date:      int(time.Now().Unix()),
			FwdFromID: sql.NullInt64{Int64: original.FromID, Valid: true},
			FwdDate:   sql.NullInt32{Int32: int32(original.Date), Valid: true},
			MediaType: original.MediaType,
			MediaID:   original.MediaID,
			Entities:  original.Entities,
			IsOut:     true,
		}

		if err := s.repo.CreateMessage(ctx, msg); err != nil {
			continue
		}

		forwardedMessages = append(forwardedMessages, msg)

		// Update dialogs
		s.repo.UpdateOrCreateDialog(ctx, fromID, toPeerID, toPeerType, newMsgID)
		if toPeerType == "user" {
			s.repo.UpdateOrCreateDialog(ctx, toPeerID, fromID, "user", newMsgID)
			s.repo.IncrementUnreadCount(ctx, toPeerID, fromID, "user")
		}
	}

	return forwardedMessages, nil
}

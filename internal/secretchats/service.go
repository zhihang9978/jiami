package secretchats

import (
	"context"
	"fmt"
	"time"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// CreateSecretChat creates a new secret chat
func (s *Service) CreateSecretChat(ctx context.Context, initiatorID, peerID int64, gA string) (*SecretChat, error) {
	chat := &SecretChat{
		InitiatorID: initiatorID,
		PeerID:      peerID,
		Status:      "WAITING",
		GA:          gA,
	}

	if err := s.repo.CreateSecretChat(ctx, chat); err != nil {
		return nil, fmt.Errorf("failed to create secret chat: %w", err)
	}

	return chat, nil
}

// UpdateSecretChatStatus updates the status of a secret chat
func (s *Service) UpdateSecretChatStatus(ctx context.Context, chatID int64, userID int64, status string, gB string) (*SecretChat, error) {
	chat, err := s.repo.GetSecretChat(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret chat: %w", err)
	}
	if chat == nil {
		return nil, fmt.Errorf("secret chat not found")
	}

	// Verify user is part of the chat
	if chat.InitiatorID != userID && chat.PeerID != userID {
		return nil, fmt.Errorf("not authorized")
	}

	chat.Status = status
	if gB != "" {
		chat.GB = gB
	}

	if err := s.repo.UpdateSecretChat(ctx, chat); err != nil {
		return nil, fmt.Errorf("failed to update secret chat: %w", err)
	}

	return chat, nil
}

// GetSecretChat retrieves a secret chat by ID
func (s *Service) GetSecretChat(ctx context.Context, chatID int64, userID int64) (*SecretChat, error) {
	chat, err := s.repo.GetSecretChat(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret chat: %w", err)
	}
	if chat == nil {
		return nil, fmt.Errorf("secret chat not found")
	}

	// Verify user is part of the chat
	if chat.InitiatorID != userID && chat.PeerID != userID {
		return nil, fmt.Errorf("not authorized")
	}

	return chat, nil
}

// GetUserSecretChats retrieves all secret chats for a user
func (s *Service) GetUserSecretChats(ctx context.Context, userID int64) ([]*SecretChat, error) {
	return s.repo.GetUserSecretChats(ctx, userID)
}

// CloseSecretChat closes a secret chat
func (s *Service) CloseSecretChat(ctx context.Context, chatID int64, userID int64) error {
	chat, err := s.repo.GetSecretChat(ctx, chatID)
	if err != nil {
		return fmt.Errorf("failed to get secret chat: %w", err)
	}
	if chat == nil {
		return fmt.Errorf("secret chat not found")
	}

	// Verify user is part of the chat
	if chat.InitiatorID != userID && chat.PeerID != userID {
		return fmt.Errorf("not authorized")
	}

	chat.Status = "CLOSED"
	if err := s.repo.UpdateSecretChat(ctx, chat); err != nil {
		return fmt.Errorf("failed to close secret chat: %w", err)
	}

	return nil
}

// SendSecretMessage sends an encrypted message
func (s *Service) SendSecretMessage(ctx context.Context, chatID int64, fromID int64, encryptedMessage string, randomID int64) (*SecretMessage, error) {
	chat, err := s.repo.GetSecretChat(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret chat: %w", err)
	}
	if chat == nil {
		return nil, fmt.Errorf("secret chat not found")
	}

	// Verify user is part of the chat
	if chat.InitiatorID != fromID && chat.PeerID != fromID {
		return nil, fmt.Errorf("not authorized")
	}

	// Verify chat is active
	if chat.Status != "ACTIVE" {
		return nil, fmt.Errorf("secret chat is not active")
	}

	msg := &SecretMessage{
		SecretChatID:     chatID,
		FromID:           fromID,
		EncryptedMessage: encryptedMessage,
		Date:             time.Now().Unix(),
	}

	if err := s.repo.CreateSecretMessage(ctx, msg); err != nil {
		return nil, fmt.Errorf("failed to create secret message: %w", err)
	}

	return msg, nil
}

// GetSecretMessages retrieves messages for a secret chat
func (s *Service) GetSecretMessages(ctx context.Context, chatID int64, userID int64, limit int) ([]*SecretMessage, error) {
	chat, err := s.repo.GetSecretChat(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret chat: %w", err)
	}
	if chat == nil {
		return nil, fmt.Errorf("secret chat not found")
	}

	// Verify user is part of the chat
	if chat.InitiatorID != userID && chat.PeerID != userID {
		return nil, fmt.Errorf("not authorized")
	}

	if limit <= 0 {
		limit = 50
	}

	return s.repo.GetSecretMessages(ctx, chatID, limit, 0)
}

// ToTL converts SecretChat to TL format
func (c *SecretChat) ToTL() map[string]interface{} {
	return map[string]interface{}{
		"secret_chat_id": c.ID,
		"initiator_id":   c.InitiatorID,
		"peer_id":        c.PeerID,
		"status":         c.Status,
		"g_a":            c.GA,
		"g_b":            c.GB,
		"key_hash":       c.KeyHash,
		"created_at":     c.CreatedAt.Unix(),
	}
}

// ToTL converts SecretMessage to TL format
func (m *SecretMessage) ToTL() map[string]interface{} {
	return map[string]interface{}{
		"message_id":        m.ID,
		"secret_chat_id":    m.SecretChatID,
		"from_id":           m.FromID,
		"encrypted_message": m.EncryptedMessage,
		"date":              m.Date,
	}
}

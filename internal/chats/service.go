package chats

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

// CreateChat creates a new chat group
func (s *Service) CreateChat(ctx context.Context, creatorID int64, title string, userIDs []int64) (*Chat, error) {
	chat := &Chat{
		Title:             title,
		ParticipantsCount: len(userIDs) + 1, // Include creator
		Date:              int(time.Now().Unix()),
		Version:           1,
		CreatorID:         creatorID,
		IsDeactivated:     false,
	}

	if err := s.repo.CreateChat(ctx, chat); err != nil {
		return nil, fmt.Errorf("failed to create chat: %w", err)
	}

	// Add creator as admin
	creatorParticipant := &ChatParticipant{
		ChatID:  chat.ID,
		UserID:  creatorID,
		Date:    int(time.Now().Unix()),
		IsAdmin: true,
	}
	if err := s.repo.AddParticipant(ctx, creatorParticipant); err != nil {
		return nil, fmt.Errorf("failed to add creator: %w", err)
	}

	// Add other users
	for _, userID := range userIDs {
		if userID == creatorID {
			continue
		}
		inviterID := creatorID
		participant := &ChatParticipant{
			ChatID:    chat.ID,
			UserID:    userID,
			InviterID: &inviterID,
			Date:      int(time.Now().Unix()),
			IsAdmin:   false,
		}
		if err := s.repo.AddParticipant(ctx, participant); err != nil {
			return nil, fmt.Errorf("failed to add participant: %w", err)
		}
	}

	return chat, nil
}

// GetChat retrieves a chat by ID
func (s *Service) GetChat(ctx context.Context, chatID int64) (*Chat, error) {
	return s.repo.GetChat(ctx, chatID)
}

// EditChatTitle updates the chat title
func (s *Service) EditChatTitle(ctx context.Context, chatID int64, userID int64, title string) (*Chat, error) {
	chat, err := s.repo.GetChat(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat: %w", err)
	}
	if chat == nil {
		return nil, fmt.Errorf("chat not found")
	}

	// Check if user is admin or creator
	participant, err := s.repo.GetParticipant(ctx, chatID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get participant: %w", err)
	}
	if participant == nil {
		return nil, fmt.Errorf("user is not a member")
	}
	if !participant.IsAdmin && chat.CreatorID != userID {
		return nil, fmt.Errorf("permission denied")
	}

	chat.Title = title
	if err := s.repo.UpdateChat(ctx, chat); err != nil {
		return nil, fmt.Errorf("failed to update chat: %w", err)
	}

	return chat, nil
}

// EditChatPhoto updates the chat photo
func (s *Service) EditChatPhoto(ctx context.Context, chatID int64, userID int64, photoID int64) (*Chat, error) {
	chat, err := s.repo.GetChat(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat: %w", err)
	}
	if chat == nil {
		return nil, fmt.Errorf("chat not found")
	}

	// Check if user is admin or creator
	participant, err := s.repo.GetParticipant(ctx, chatID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get participant: %w", err)
	}
	if participant == nil {
		return nil, fmt.Errorf("user is not a member")
	}
	if !participant.IsAdmin && chat.CreatorID != userID {
		return nil, fmt.Errorf("permission denied")
	}

	chat.PhotoID = &photoID
	if err := s.repo.UpdateChat(ctx, chat); err != nil {
		return nil, fmt.Errorf("failed to update chat: %w", err)
	}

	return chat, nil
}

// AddChatUser adds a user to a chat
func (s *Service) AddChatUser(ctx context.Context, chatID, inviterID, userID int64, fwdLimit int) (*Chat, error) {
	chat, err := s.repo.GetChat(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat: %w", err)
	}
	if chat == nil {
		return nil, fmt.Errorf("chat not found")
	}

	// Check if inviter is a member
	inviter, err := s.repo.GetParticipant(ctx, chatID, inviterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get inviter: %w", err)
	}
	if inviter == nil {
		return nil, fmt.Errorf("inviter is not a member")
	}

	// Check if user is already a member
	existing, err := s.repo.GetParticipant(ctx, chatID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("user is already a member")
	}

	// Add participant
	participant := &ChatParticipant{
		ChatID:    chatID,
		UserID:    userID,
		InviterID: &inviterID,
		Date:      int(time.Now().Unix()),
		IsAdmin:   false,
	}
	if err := s.repo.AddParticipant(ctx, participant); err != nil {
		return nil, fmt.Errorf("failed to add participant: %w", err)
	}

	// Update participants count
	if err := s.repo.IncrementParticipantsCount(ctx, chatID, 1); err != nil {
		return nil, fmt.Errorf("failed to update count: %w", err)
	}

	chat.ParticipantsCount++
	return chat, nil
}

// DeleteChatUser removes a user from a chat
func (s *Service) DeleteChatUser(ctx context.Context, chatID, requesterID, userID int64) (*Chat, error) {
	chat, err := s.repo.GetChat(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat: %w", err)
	}
	if chat == nil {
		return nil, fmt.Errorf("chat not found")
	}

	// Check if requester is admin or the user themselves
	if requesterID != userID {
		requester, err := s.repo.GetParticipant(ctx, chatID, requesterID)
		if err != nil {
			return nil, fmt.Errorf("failed to get requester: %w", err)
		}
		if requester == nil {
			return nil, fmt.Errorf("requester is not a member")
		}
		if !requester.IsAdmin && chat.CreatorID != requesterID {
			return nil, fmt.Errorf("permission denied")
		}
	}

	// Remove participant
	if err := s.repo.RemoveParticipant(ctx, chatID, userID); err != nil {
		return nil, fmt.Errorf("failed to remove participant: %w", err)
	}

	// Update participants count
	if err := s.repo.IncrementParticipantsCount(ctx, chatID, -1); err != nil {
		return nil, fmt.Errorf("failed to update count: %w", err)
	}

	chat.ParticipantsCount--
	return chat, nil
}

// LeaveChat removes the user from a chat
func (s *Service) LeaveChat(ctx context.Context, chatID, userID int64) (*Chat, error) {
	return s.DeleteChatUser(ctx, chatID, userID, userID)
}

// GetFullChat retrieves full chat info including participants
func (s *Service) GetFullChat(ctx context.Context, chatID int64) (*FullChat, error) {
	chat, err := s.repo.GetChat(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat: %w", err)
	}
	if chat == nil {
		return nil, fmt.Errorf("chat not found")
	}

	participants, err := s.repo.GetParticipants(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get participants: %w", err)
	}

	return &FullChat{
		Chat:         chat,
		Participants: participants,
	}, nil
}

// EditChatAdmin updates admin status for a user
func (s *Service) EditChatAdmin(ctx context.Context, chatID, requesterID, userID int64, isAdmin bool) (*Chat, error) {
	chat, err := s.repo.GetChat(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat: %w", err)
	}
	if chat == nil {
		return nil, fmt.Errorf("chat not found")
	}

	// Only creator can edit admins
	if chat.CreatorID != requesterID {
		return nil, fmt.Errorf("only creator can edit admins")
	}

	// Check if user is a member
	participant, err := s.repo.GetParticipant(ctx, chatID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get participant: %w", err)
	}
	if participant == nil {
		return nil, fmt.Errorf("user is not a member")
	}

	if err := s.repo.UpdateParticipantAdmin(ctx, chatID, userID, isAdmin, nil); err != nil {
		return nil, fmt.Errorf("failed to update admin: %w", err)
	}

	return chat, nil
}

// GetUserChats retrieves all chats a user is a member of
func (s *Service) GetUserChats(ctx context.Context, userID int64) ([]*Chat, error) {
	return s.repo.GetUserChats(ctx, userID)
}

// FullChat represents full chat info
type FullChat struct {
	Chat         *Chat
	Participants []*ChatParticipant
}

// ToTL converts Chat to TL format
func (c *Chat) ToTL() map[string]interface{} {
	return map[string]interface{}{
		"_":                    "chat",
		"id":                   c.ID,
		"title":                c.Title,
		"photo":                c.getPhotoTL(),
		"participants_count":   c.ParticipantsCount,
		"date":                 c.Date,
		"version":              c.Version,
		"creator":              c.CreatorID == 0,
		"deactivated":          c.IsDeactivated,
		"call_active":          c.IsCallActive,
		"call_not_empty":       c.IsCallNotEmpty,
		"migrated_to":          c.MigratedToChannelID,
		"admin_rights":         c.AdminRights,
		"default_banned_rights": c.DefaultBannedRights,
	}
}

func (c *Chat) getPhotoTL() map[string]interface{} {
	if c.PhotoID == nil {
		return map[string]interface{}{"_": "chatPhotoEmpty"}
	}
	return map[string]interface{}{
		"_":        "chatPhoto",
		"photo_id": *c.PhotoID,
	}
}

// ToTL converts FullChat to TL format
func (fc *FullChat) ToTL() map[string]interface{} {
	participants := make([]map[string]interface{}, len(fc.Participants))
	for i, p := range fc.Participants {
		participants[i] = p.ToTL()
	}

	return map[string]interface{}{
		"_":            "messages.chatFull",
		"full_chat":    fc.Chat.ToTL(),
		"chats":        []interface{}{fc.Chat.ToTL()},
		"users":        []interface{}{},
		"participants": participants,
	}
}

// ToTL converts ChatParticipant to TL format
func (p *ChatParticipant) ToTL() map[string]interface{} {
	result := map[string]interface{}{
		"_":       "chatParticipant",
		"user_id": p.UserID,
		"date":    p.Date,
	}
	if p.InviterID != nil {
		result["inviter_id"] = *p.InviterID
	}
	if p.IsAdmin {
		result["_"] = "chatParticipantAdmin"
	}
	return result
}

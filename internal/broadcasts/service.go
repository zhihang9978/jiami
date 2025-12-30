package broadcasts

import (
	"context"
	"fmt"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// CreateBroadcast creates a new broadcast
func (s *Service) CreateBroadcast(ctx context.Context, creatorID int64, title, content string) (*Broadcast, error) {
	broadcast := &Broadcast{
		CreatorID: creatorID,
		Title:     title,
		Content:   content,
		Status:    "DRAFT",
	}

	if err := s.repo.CreateBroadcast(ctx, broadcast); err != nil {
		return nil, fmt.Errorf("failed to create broadcast: %w", err)
	}

	return broadcast, nil
}

// GetBroadcast retrieves a broadcast by ID
func (s *Service) GetBroadcast(ctx context.Context, broadcastID int64) (*Broadcast, error) {
	broadcast, err := s.repo.GetBroadcast(ctx, broadcastID)
	if err != nil {
		return nil, fmt.Errorf("failed to get broadcast: %w", err)
	}
	if broadcast == nil {
		return nil, fmt.Errorf("broadcast not found")
	}
	return broadcast, nil
}

// UpdateBroadcast updates a broadcast
func (s *Service) UpdateBroadcast(ctx context.Context, broadcastID int64, title, content string) (*Broadcast, error) {
	broadcast, err := s.repo.GetBroadcast(ctx, broadcastID)
	if err != nil {
		return nil, fmt.Errorf("failed to get broadcast: %w", err)
	}
	if broadcast == nil {
		return nil, fmt.Errorf("broadcast not found")
	}

	if title != "" {
		broadcast.Title = title
	}
	if content != "" {
		broadcast.Content = content
	}

	if err := s.repo.UpdateBroadcast(ctx, broadcast); err != nil {
		return nil, fmt.Errorf("failed to update broadcast: %w", err)
	}

	return broadcast, nil
}

// DeleteBroadcast deletes a broadcast
func (s *Service) DeleteBroadcast(ctx context.Context, broadcastID int64) error {
	return s.repo.DeleteBroadcast(ctx, broadcastID)
}

// GetBroadcasts retrieves all broadcasts with pagination
func (s *Service) GetBroadcasts(ctx context.Context, page, pageSize int) ([]*Broadcast, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize

	return s.repo.GetBroadcasts(ctx, pageSize, offset)
}

// GetUserBroadcasts retrieves broadcasts created by a user
func (s *Service) GetUserBroadcasts(ctx context.Context, userID int64, page, pageSize int) ([]*Broadcast, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize

	return s.repo.GetUserBroadcasts(ctx, userID, pageSize, offset)
}

// SendBroadcast marks a broadcast as sent
func (s *Service) SendBroadcast(ctx context.Context, broadcastID int64) (*Broadcast, error) {
	broadcast, err := s.repo.GetBroadcast(ctx, broadcastID)
	if err != nil {
		return nil, fmt.Errorf("failed to get broadcast: %w", err)
	}
	if broadcast == nil {
		return nil, fmt.Errorf("broadcast not found")
	}

	if err := s.repo.MarkAsSent(ctx, broadcastID); err != nil {
		return nil, fmt.Errorf("failed to send broadcast: %w", err)
	}

	// Refresh broadcast data
	return s.repo.GetBroadcast(ctx, broadcastID)
}

// ToTL converts Broadcast to TL format
func (b *Broadcast) ToTL() map[string]interface{} {
	result := map[string]interface{}{
		"broadcast_id": b.ID,
		"creator_id":   b.CreatorID,
		"title":        b.Title,
		"content":      b.Content,
		"status":       b.Status,
		"created_at":   b.CreatedAt.Unix(),
	}
	if b.SentAt != nil {
		result["sent_at"] = b.SentAt.Unix()
	}
	return result
}

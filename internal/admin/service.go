package admin

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

// GetStats retrieves system statistics
func (s *Service) GetStats(ctx context.Context) (*Stats, error) {
	return s.repo.GetStats(ctx)
}

// GetUsers retrieves users with pagination
func (s *Service) GetUsers(ctx context.Context, page, pageSize int, search string) ([]*UserInfo, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return s.repo.GetUsers(ctx, offset, pageSize, search)
}

// GetUser retrieves a single user
func (s *Service) GetUser(ctx context.Context, userID int64) (*UserInfo, error) {
	return s.repo.GetUser(ctx, userID)
}

// CreateUser creates a new user
func (s *Service) CreateUser(ctx context.Context, phone, firstName, lastName string, username *string) (*UserInfo, error) {
	return s.repo.CreateUser(ctx, phone, firstName, lastName, username)
}

// UpdateUser updates a user
func (s *Service) UpdateUser(ctx context.Context, userID int64, firstName, lastName, bio string, username *string) error {
	user, err := s.repo.GetUser(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}
	return s.repo.UpdateUser(ctx, userID, firstName, lastName, bio, username)
}

// DeleteUser soft deletes a user
func (s *Service) DeleteUser(ctx context.Context, userID int64) error {
	return s.repo.DeleteUser(ctx, userID)
}

// BanUser bans a user
func (s *Service) BanUser(ctx context.Context, userID int64) error {
	return s.repo.BanUser(ctx, userID)
}

// UnbanUser unbans a user
func (s *Service) UnbanUser(ctx context.Context, userID int64) error {
	return s.repo.UnbanUser(ctx, userID)
}

// GetMessages retrieves messages with pagination
func (s *Service) GetMessages(ctx context.Context, page, pageSize int, userID int64) ([]map[string]interface{}, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return s.repo.GetMessages(ctx, offset, pageSize, userID)
}

// DeleteMessage deletes a message
func (s *Service) DeleteMessage(ctx context.Context, messageID int64) error {
	return s.repo.DeleteMessage(ctx, messageID)
}

// GetChats retrieves chats with pagination
func (s *Service) GetChats(ctx context.Context, page, pageSize int) ([]map[string]interface{}, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return s.repo.GetChats(ctx, offset, pageSize)
}

// GetChannels retrieves channels with pagination
func (s *Service) GetChannels(ctx context.Context, page, pageSize int) ([]map[string]interface{}, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return s.repo.GetChannels(ctx, offset, pageSize)
}

// GetSessions retrieves active sessions for a user
func (s *Service) GetSessions(ctx context.Context, userID int64) ([]map[string]interface{}, error) {
	return s.repo.GetSessions(ctx, userID)
}

// TerminateSession terminates a session
func (s *Service) TerminateSession(ctx context.Context, sessionID int64) error {
	return s.repo.TerminateSession(ctx, sessionID)
}

// TerminateAllSessions terminates all sessions for a user
func (s *Service) TerminateAllSessions(ctx context.Context, userID int64) error {
	return s.repo.TerminateAllSessions(ctx, userID)
}

// ToTL converts Stats to TL format
func (s *Stats) ToTL() map[string]interface{} {
	return map[string]interface{}{
		"total_users":     s.TotalUsers,
		"active_users":    s.ActiveUsers,
		"total_messages":  s.TotalMessages,
		"total_chats":     s.TotalChats,
		"total_channels":  s.TotalChannels,
		"online_users":    s.OnlineUsers,
		"new_users_today": s.NewUsersToday,
		"messages_today":  s.MessagesToday,
	}
}

// ToTL converts UserInfo to TL format
func (u *UserInfo) ToTL() map[string]interface{} {
	result := map[string]interface{}{
		"id":           u.ID,
		"phone":        u.Phone,
		"first_name":   u.FirstName,
		"last_name":    u.LastName,
		"bio":          u.Bio,
		"is_bot":       u.IsBot,
		"is_verified":  u.IsVerified,
		"is_restricted": u.IsRestricted,
		"is_deleted":   u.IsDeleted,
		"last_online":  u.LastOnline,
		"created_at":   u.CreatedAt,
		"updated_at":   u.UpdatedAt,
	}
	if u.Username != nil {
		result["username"] = *u.Username
	}
	if u.PhotoID != nil {
		result["photo_id"] = *u.PhotoID
	}
	return result
}

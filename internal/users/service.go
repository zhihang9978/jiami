package users

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/feiji/feiji-backend/internal/models"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// GetUser retrieves a user by ID
func (s *Service) GetUser(ctx context.Context, userID int64) (*models.User, error) {
	return s.repo.GetUserByID(ctx, userID)
}

// GetUsers retrieves multiple users by IDs
func (s *Service) GetUsers(ctx context.Context, userIDs []int64) ([]*models.User, error) {
	return s.repo.GetUsersByIDs(ctx, userIDs)
}

// GetFullUser retrieves full user info
func (s *Service) GetFullUser(ctx context.Context, userID int64) (*FullUser, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	return &FullUser{
		User:         user,
		About:        user.Bio.String,
		CommonChats:  0,
		CanPinMessage: true,
		CanCall:      true,
	}, nil
}

// UpdateProfile updates user profile
func (s *Service) UpdateProfile(ctx context.Context, userID int64, firstName, lastName, about string) (*models.User, error) {
	if firstName == "" {
		return nil, fmt.Errorf("first name is required")
	}

	if err := s.repo.UpdateProfile(ctx, userID, firstName, lastName, about); err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	return s.repo.GetUserByID(ctx, userID)
}

// UpdateUsername updates user's username
func (s *Service) UpdateUsername(ctx context.Context, userID int64, username string) (*models.User, error) {
	// Validate username
	if username != "" {
		if len(username) < 5 || len(username) > 32 {
			return nil, fmt.Errorf("username must be between 5 and 32 characters")
		}
		if !regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`).MatchString(username) {
			return nil, fmt.Errorf("username must start with a letter and contain only letters, numbers, and underscores")
		}

		// Check availability
		available, err := s.repo.CheckUsernameAvailable(ctx, username)
		if err != nil {
			return nil, fmt.Errorf("failed to check username: %w", err)
		}
		if !available {
			return nil, fmt.Errorf("username is already taken")
		}
	}

	if err := s.repo.UpdateUsername(ctx, userID, username); err != nil {
		return nil, fmt.Errorf("failed to update username: %w", err)
	}

	return s.repo.GetUserByID(ctx, userID)
}

// CheckUsername checks if a username is available
func (s *Service) CheckUsername(ctx context.Context, username string) (bool, error) {
	if username == "" {
		return false, fmt.Errorf("username is required")
	}

	username = strings.ToLower(username)

	if len(username) < 5 || len(username) > 32 {
		return false, nil
	}
	if !regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`).MatchString(username) {
		return false, nil
	}

	return s.repo.CheckUsernameAvailable(ctx, username)
}

// UpdatePhoto updates user's photo
func (s *Service) UpdatePhoto(ctx context.Context, userID int64, photoID int64) (*models.User, error) {
	if err := s.repo.UpdatePhoto(ctx, userID, photoID); err != nil {
		return nil, fmt.Errorf("failed to update photo: %w", err)
	}
	return s.repo.GetUserByID(ctx, userID)
}

// DeletePhoto removes user's photo
func (s *Service) DeletePhoto(ctx context.Context, userID int64) error {
	return s.repo.DeletePhoto(ctx, userID)
}

// UpdateStatus updates user's online status
func (s *Service) UpdateStatus(ctx context.Context, userID int64, offline bool) error {
	if offline {
		return s.repo.UpdateStatus(ctx, userID, "offline", int(0))
	}
	return s.repo.UpdateStatus(ctx, userID, "online", int(0))
}

// SearchUsers searches for users
func (s *Service) SearchUsers(ctx context.Context, query string, limit int) ([]*models.User, error) {
	if query == "" {
		return []*models.User{}, nil
	}
	return s.repo.SearchUsers(ctx, query, limit)
}

// ResolveUsername resolves a username to a user
func (s *Service) ResolveUsername(ctx context.Context, username string) (*models.User, error) {
	return s.repo.GetUserByUsername(ctx, username)
}

// Types
type FullUser struct {
	User          *models.User `json:"user"`
	About         string       `json:"about"`
	CommonChats   int          `json:"common_chats_count"`
	CanPinMessage bool         `json:"can_pin_message"`
	CanCall       bool         `json:"can_call"`
}

// ToTL converts to TL format
func (f *FullUser) ToTL() map[string]interface{} {
	result := map[string]interface{}{
		"_": "userFull",
		"full_user": map[string]interface{}{
			"_":                   "userFull",
			"id":                  f.User.ID,
			"about":               f.About,
			"common_chats_count":  f.CommonChats,
			"can_pin_message":     f.CanPinMessage,
			"phone_calls_available": f.CanCall,
			"phone_calls_private": false,
			"video_calls_available": f.CanCall,
		},
		"users": []interface{}{f.User.ToTLUser()},
		"chats": []interface{}{},
	}
	return result
}

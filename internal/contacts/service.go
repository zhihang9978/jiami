package contacts

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

// GetContacts retrieves all contacts for a user
func (s *Service) GetContacts(ctx context.Context, userID int64) ([]*Contact, error) {
	return s.repo.GetContacts(ctx, userID)
}

// ImportContacts imports contacts from phone numbers
func (s *Service) ImportContacts(ctx context.Context, userID int64, contacts []ImportContact) (*ImportResult, error) {
	result := &ImportResult{
		Imported: make([]ImportedContact, 0),
		Users:    make([]*Contact, 0),
	}

	for _, c := range contacts {
		// Find user by phone
		user, err := s.repo.GetUserByPhone(ctx, c.Phone)
		if err != nil {
			continue
		}
		if user == nil {
			// User not found, skip
			continue
		}

		// Add contact
		firstName := c.FirstName
		if firstName == "" {
			firstName = user.FirstName
		}
		lastName := c.LastName
		if lastName == "" {
			lastName = user.LastName
		}

		if err := s.repo.AddContact(ctx, userID, user.ID, firstName, lastName); err != nil {
			continue
		}

		// Update mutual status
		s.repo.UpdateMutualStatus(ctx, userID, user.ID)

		result.Imported = append(result.Imported, ImportedContact{
			UserID:   user.ID,
			ClientID: c.ClientID,
		})

		// Get updated contact info
		contactInfo := &Contact{
			ContactID:     user.ID,
			FirstName:     firstName,
			LastName:      lastName,
			Phone:         user.Phone,
			AccessHash:    user.AccessHash,
			StatusType:    user.StatusType,
			StatusExpires: user.StatusExpires,
		}
		if user.Username.Valid {
			contactInfo.Username = user.Username.String
		}
		result.Users = append(result.Users, contactInfo)
	}

	return result, nil
}

// AddContact adds a single contact
func (s *Service) AddContact(ctx context.Context, userID, contactID int64, firstName, lastName string) error {
	if err := s.repo.AddContact(ctx, userID, contactID, firstName, lastName); err != nil {
		return fmt.Errorf("failed to add contact: %w", err)
	}
	return s.repo.UpdateMutualStatus(ctx, userID, contactID)
}

// DeleteContacts deletes contacts
func (s *Service) DeleteContacts(ctx context.Context, userID int64, contactIDs []int64) error {
	for _, contactID := range contactIDs {
		if err := s.repo.DeleteContact(ctx, userID, contactID); err != nil {
			return fmt.Errorf("failed to delete contact %d: %w", contactID, err)
		}
	}
	return nil
}

// BlockContact blocks a user
func (s *Service) BlockContact(ctx context.Context, userID, contactID int64) error {
	return s.repo.BlockContact(ctx, userID, contactID)
}

// UnblockContact unblocks a user
func (s *Service) UnblockContact(ctx context.Context, userID, contactID int64) error {
	return s.repo.UnblockContact(ctx, userID, contactID)
}

// GetBlocked retrieves blocked users
func (s *Service) GetBlocked(ctx context.Context, userID int64) ([]*Contact, error) {
	return s.repo.GetBlockedContacts(ctx, userID)
}

// Search searches for users by query
func (s *Service) Search(ctx context.Context, query string, limit int) ([]*Contact, error) {
	// This would search users by username or name
	// For now, return empty
	return []*Contact{}, nil
}

// Types
type ImportContact struct {
	ClientID  int64  `json:"client_id"`
	Phone     string `json:"phone"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type ImportedContact struct {
	UserID   int64 `json:"user_id"`
	ClientID int64 `json:"client_id"`
}

type ImportResult struct {
	Imported []ImportedContact `json:"imported"`
	Users    []*Contact        `json:"users"`
}

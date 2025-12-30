package contacts

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/feiji/feiji-backend/internal/models"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// GetContacts retrieves all contacts for a user
func (r *Repository) GetContacts(ctx context.Context, userID int64) ([]*Contact, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT c.id, c.user_id, c.contact_id, c.first_name, c.last_name, c.is_mutual, c.is_blocked, c.created_at,
		       u.phone, u.username, u.first_name as user_first_name, u.last_name as user_last_name,
		       u.access_hash, u.status_type, u.status_expires, u.photo_id
		FROM contacts c
		JOIN users u ON c.contact_id = u.id
		WHERE c.user_id = ? AND c.is_blocked = 0
		ORDER BY c.first_name, c.last_name
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get contacts: %w", err)
	}
	defer rows.Close()

	var contacts []*Contact
	for rows.Next() {
		c := &Contact{}
		var username, userFirstName, userLastName sql.NullString
		var photoID sql.NullInt64
		err := rows.Scan(
			&c.ID, &c.UserID, &c.ContactID, &c.FirstName, &c.LastName, &c.IsMutual, &c.IsBlocked, &c.CreatedAt,
			&c.Phone, &username, &userFirstName, &userLastName, &c.AccessHash, &c.StatusType, &c.StatusExpires, &photoID,
		)
		if err != nil {
			return nil, err
		}
		if username.Valid {
			c.Username = username.String
		}
		if userFirstName.Valid {
			c.UserFirstName = userFirstName.String
		}
		if userLastName.Valid {
			c.UserLastName = userLastName.String
		}
		if photoID.Valid {
			c.PhotoID = photoID.Int64
		}
		contacts = append(contacts, c)
	}

	return contacts, nil
}

// AddContact adds a contact
func (r *Repository) AddContact(ctx context.Context, userID, contactID int64, firstName, lastName string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO contacts (user_id, contact_id, first_name, last_name)
		VALUES (?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE first_name = VALUES(first_name), last_name = VALUES(last_name)
	`, userID, contactID, firstName, lastName)
	return err
}

// DeleteContact deletes a contact
func (r *Repository) DeleteContact(ctx context.Context, userID, contactID int64) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM contacts WHERE user_id = ? AND contact_id = ?
	`, userID, contactID)
	return err
}

// BlockContact blocks a contact
func (r *Repository) BlockContact(ctx context.Context, userID, contactID int64) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO contacts (user_id, contact_id, is_blocked)
		VALUES (?, ?, 1)
		ON DUPLICATE KEY UPDATE is_blocked = 1
	`, userID, contactID)
	return err
}

// UnblockContact unblocks a contact
func (r *Repository) UnblockContact(ctx context.Context, userID, contactID int64) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE contacts SET is_blocked = 0 WHERE user_id = ? AND contact_id = ?
	`, userID, contactID)
	return err
}

// GetBlockedContacts retrieves blocked contacts
func (r *Repository) GetBlockedContacts(ctx context.Context, userID int64) ([]*Contact, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT c.id, c.user_id, c.contact_id, c.first_name, c.last_name, c.is_mutual, c.is_blocked, c.created_at,
		       u.phone, u.username, u.first_name as user_first_name, u.last_name as user_last_name,
		       u.access_hash, u.status_type, u.status_expires, u.photo_id
		FROM contacts c
		JOIN users u ON c.contact_id = u.id
		WHERE c.user_id = ? AND c.is_blocked = 1
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get blocked contacts: %w", err)
	}
	defer rows.Close()

	var contacts []*Contact
	for rows.Next() {
		c := &Contact{}
		var username, userFirstName, userLastName sql.NullString
		var photoID sql.NullInt64
		err := rows.Scan(
			&c.ID, &c.UserID, &c.ContactID, &c.FirstName, &c.LastName, &c.IsMutual, &c.IsBlocked, &c.CreatedAt,
			&c.Phone, &username, &userFirstName, &userLastName, &c.AccessHash, &c.StatusType, &c.StatusExpires, &photoID,
		)
		if err != nil {
			return nil, err
		}
		if username.Valid {
			c.Username = username.String
		}
		contacts = append(contacts, c)
	}

	return contacts, nil
}

// GetUserByPhone retrieves a user by phone number
func (r *Repository) GetUserByPhone(ctx context.Context, phone string) (*models.User, error) {
	user := &models.User{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, phone, username, first_name, last_name, bio, photo_id, access_hash,
		       status_type, status_expires, is_bot, is_verified, is_premium
		FROM users WHERE phone = ?
	`, phone).Scan(
		&user.ID, &user.Phone, &user.Username, &user.FirstName, &user.LastName,
		&user.Bio, &user.PhotoID, &user.AccessHash, &user.StatusType, &user.StatusExpires,
		&user.IsBot, &user.IsVerified, &user.IsPremium,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

// UpdateMutualStatus updates mutual contact status
func (r *Repository) UpdateMutualStatus(ctx context.Context, userID, contactID int64) error {
	// Check if contact has user in their contacts
	var count int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM contacts WHERE user_id = ? AND contact_id = ?
	`, contactID, userID).Scan(&count)
	if err != nil {
		return err
	}

	isMutual := count > 0

	// Update both sides
	_, err = r.db.ExecContext(ctx, `
		UPDATE contacts SET is_mutual = ? WHERE user_id = ? AND contact_id = ?
	`, isMutual, userID, contactID)
	if err != nil {
		return err
	}

	if isMutual {
		_, err = r.db.ExecContext(ctx, `
			UPDATE contacts SET is_mutual = 1 WHERE user_id = ? AND contact_id = ?
		`, contactID, userID)
	}

	return err
}

// Contact model
type Contact struct {
	ID            int64  `json:"id"`
	UserID        int64  `json:"user_id"`
	ContactID     int64  `json:"contact_id"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	IsMutual      bool   `json:"is_mutual"`
	IsBlocked     bool   `json:"is_blocked"`
	CreatedAt     string `json:"created_at"`
	Phone         string `json:"phone"`
	Username      string `json:"username"`
	UserFirstName string `json:"user_first_name"`
	UserLastName  string `json:"user_last_name"`
	AccessHash    int64  `json:"access_hash"`
	StatusType    string `json:"status_type"`
	StatusExpires int    `json:"status_expires"`
	PhotoID       int64  `json:"photo_id"`
}

// ToTLUser converts contact to TL User format
func (c *Contact) ToTLUser() map[string]interface{} {
	result := map[string]interface{}{
		"_":           "user",
		"id":          c.ContactID,
		"access_hash": c.AccessHash,
		"first_name":  c.FirstName,
		"phone":       c.Phone,
		"contact":     true,
		"mutual_contact": c.IsMutual,
	}

	if c.LastName != "" {
		result["last_name"] = c.LastName
	}
	if c.Username != "" {
		result["username"] = c.Username
	}

	// Status
	switch c.StatusType {
	case "online":
		result["status"] = map[string]interface{}{
			"_":       "userStatusOnline",
			"expires": c.StatusExpires,
		}
	case "offline":
		result["status"] = map[string]interface{}{
			"_":          "userStatusOffline",
			"was_online": c.StatusExpires,
		}
	case "recently":
		result["status"] = map[string]interface{}{
			"_": "userStatusRecently",
		}
	}

	return result
}

// ToTLContact converts to TL Contact format
func (c *Contact) ToTLContact() map[string]interface{} {
	return map[string]interface{}{
		"_":       "contact",
		"user_id": c.ContactID,
		"mutual":  c.IsMutual,
	}
}

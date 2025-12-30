package users

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

// GetUserByID retrieves a user by ID
func (r *Repository) GetUserByID(ctx context.Context, userID int64) (*models.User, error) {
	user := &models.User{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, phone, username, first_name, last_name, bio, photo_id, access_hash,
		       status_type, status_expires, password_hash, is_bot, is_verified, is_premium,
		       is_banned, ban_reason, ban_expires_at, created_at, updated_at, last_login_at, last_login_ip
		FROM users WHERE id = ?
	`, userID).Scan(
		&user.ID, &user.Phone, &user.Username, &user.FirstName, &user.LastName,
		&user.Bio, &user.PhotoID, &user.AccessHash, &user.StatusType, &user.StatusExpires,
		&user.PasswordHash, &user.IsBot, &user.IsVerified, &user.IsPremium,
		&user.IsBanned, &user.BanReason, &user.BanExpiresAt, &user.CreatedAt, &user.UpdatedAt,
		&user.LastLoginAt, &user.LastLoginIP,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}
	return user, nil
}

// GetUsersByIDs retrieves multiple users by IDs
func (r *Repository) GetUsersByIDs(ctx context.Context, userIDs []int64) ([]*models.User, error) {
	if len(userIDs) == 0 {
		return []*models.User{}, nil
	}

	// Build placeholders
	placeholders := ""
	args := make([]interface{}, len(userIDs))
	for i, id := range userIDs {
		if i > 0 {
			placeholders += ","
		}
		placeholders += "?"
		args[i] = id
	}

	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT id, phone, username, first_name, last_name, bio, photo_id, access_hash,
		       status_type, status_expires, is_bot, is_verified, is_premium
		FROM users WHERE id IN (%s)
	`, placeholders), args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(
			&user.ID, &user.Phone, &user.Username, &user.FirstName, &user.LastName,
			&user.Bio, &user.PhotoID, &user.AccessHash, &user.StatusType, &user.StatusExpires,
			&user.IsBot, &user.IsVerified, &user.IsPremium,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

// GetUserByUsername retrieves a user by username
func (r *Repository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	user := &models.User{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, phone, username, first_name, last_name, bio, photo_id, access_hash,
		       status_type, status_expires, is_bot, is_verified, is_premium
		FROM users WHERE username = ?
	`, username).Scan(
		&user.ID, &user.Phone, &user.Username, &user.FirstName, &user.LastName,
		&user.Bio, &user.PhotoID, &user.AccessHash, &user.StatusType, &user.StatusExpires,
		&user.IsBot, &user.IsVerified, &user.IsPremium,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}
	return user, nil
}

// UpdateProfile updates user profile
func (r *Repository) UpdateProfile(ctx context.Context, userID int64, firstName, lastName, about string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE users SET first_name = ?, last_name = ?, bio = ?, updated_at = NOW()
		WHERE id = ?
	`, firstName, lastName, about, userID)
	return err
}

// UpdateUsername updates user's username
func (r *Repository) UpdateUsername(ctx context.Context, userID int64, username string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE users SET username = ?, updated_at = NOW() WHERE id = ?
	`, username, userID)
	return err
}

// UpdatePhoto updates user's photo
func (r *Repository) UpdatePhoto(ctx context.Context, userID int64, photoID int64) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE users SET photo_id = ?, updated_at = NOW() WHERE id = ?
	`, photoID, userID)
	return err
}

// DeletePhoto removes user's photo
func (r *Repository) DeletePhoto(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE users SET photo_id = NULL, updated_at = NOW() WHERE id = ?
	`, userID)
	return err
}

// UpdateStatus updates user's online status
func (r *Repository) UpdateStatus(ctx context.Context, userID int64, statusType string, expires int) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE users SET status_type = ?, status_expires = ?, updated_at = NOW() WHERE id = ?
	`, statusType, expires, userID)
	return err
}

// SearchUsers searches for users by query
func (r *Repository) SearchUsers(ctx context.Context, query string, limit int) ([]*models.User, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	searchPattern := "%" + query + "%"
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, phone, username, first_name, last_name, bio, photo_id, access_hash,
		       status_type, status_expires, is_bot, is_verified, is_premium
		FROM users 
		WHERE (username LIKE ? OR first_name LIKE ? OR last_name LIKE ?)
		  AND is_banned = 0
		LIMIT ?
	`, searchPattern, searchPattern, searchPattern, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(
			&user.ID, &user.Phone, &user.Username, &user.FirstName, &user.LastName,
			&user.Bio, &user.PhotoID, &user.AccessHash, &user.StatusType, &user.StatusExpires,
			&user.IsBot, &user.IsVerified, &user.IsPremium,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

// CheckUsernameAvailable checks if a username is available
func (r *Repository) CheckUsernameAvailable(ctx context.Context, username string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM users WHERE username = ?
	`, username).Scan(&count)
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

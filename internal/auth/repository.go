package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/feiji/feiji-backend/internal/models"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// CreateUser creates a new user
func (r *Repository) CreateUser(ctx context.Context, phone, firstName, lastName string) (*models.User, error) {
	accessHash := generateAccessHash()

	result, err := r.db.ExecContext(ctx, `
		INSERT INTO users (phone, first_name, last_name, access_hash, status_type)
		VALUES (?, ?, ?, ?, 'offline')
	`, phone, firstName, lastName, accessHash)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return &models.User{
		ID:         id,
		Phone:      phone,
		FirstName:  firstName,
		LastName:   lastName,
		AccessHash: accessHash,
		StatusType: "offline",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}, nil
}

// GetUserByPhone retrieves a user by phone number
func (r *Repository) GetUserByPhone(ctx context.Context, phone string) (*models.User, error) {
	user := &models.User{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, phone, username, first_name, last_name, bio, photo_id, access_hash,
		       status_type, status_expires, password_hash, is_bot, is_verified, is_premium,
		       is_banned, ban_reason, ban_expires_at, created_at, updated_at, last_login_at, last_login_ip
		FROM users WHERE phone = ?
	`, phone).Scan(
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
		return nil, fmt.Errorf("failed to get user by phone: %w", err)
	}
	return user, nil
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

// UpdateUserLogin updates user's last login info
func (r *Repository) UpdateUserLogin(ctx context.Context, userID int64, ip string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE users SET last_login_at = NOW(), last_login_ip = ?, status_type = 'online', status_expires = ?
		WHERE id = ?
	`, ip, time.Now().Add(5*time.Minute).Unix(), userID)
	return err
}

// SaveAuthKey saves an auth key
func (r *Repository) SaveAuthKey(ctx context.Context, authKeyID string, authKey []byte, userID *int64) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO auth_keys (auth_key_id, auth_key, user_id)
		VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE auth_key = VALUES(auth_key), user_id = VALUES(user_id)
	`, authKeyID, authKey, userID)
	return err
}

// GetAuthKey retrieves an auth key
func (r *Repository) GetAuthKey(ctx context.Context, authKeyID string) ([]byte, *int64, error) {
	var authKey []byte
	var userID sql.NullInt64

	err := r.db.QueryRowContext(ctx, `
		SELECT auth_key, user_id FROM auth_keys WHERE auth_key_id = ?
	`, authKeyID).Scan(&authKey, &userID)
	if err == sql.ErrNoRows {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}

	if userID.Valid {
		uid := userID.Int64
		return authKey, &uid, nil
	}
	return authKey, nil, nil
}

// BindAuthKeyToUser binds an auth key to a user
func (r *Repository) BindAuthKeyToUser(ctx context.Context, authKeyID string, userID int64) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE auth_keys SET user_id = ? WHERE auth_key_id = ?
	`, userID, authKeyID)
	return err
}

// CreateSession creates a new session
func (r *Repository) CreateSession(ctx context.Context, authKeyID string, userID int64, sessionID string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO sessions (auth_key_id, user_id, session_id, date)
		VALUES (?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE last_activity = NOW()
	`, authKeyID, userID, sessionID, time.Now().Unix())
	return err
}

// GetSession retrieves a session
func (r *Repository) GetSession(ctx context.Context, authKeyID string) (*int64, error) {
	var userID sql.NullInt64
	err := r.db.QueryRowContext(ctx, `
		SELECT user_id FROM sessions WHERE auth_key_id = ? ORDER BY last_activity DESC LIMIT 1
	`, authKeyID).Scan(&userID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if userID.Valid {
		uid := userID.Int64
		return &uid, nil
	}
	return nil, nil
}

// CheckUniversalCode checks if a code is a universal code
func (r *Repository) CheckUniversalCode(ctx context.Context, code string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM universal_codes WHERE code = ? AND is_active = 1
	`, code).Scan(&count)
	if err != nil {
		return false, err
	}
	if count > 0 {
		// Increment usage count
		r.db.ExecContext(ctx, `UPDATE universal_codes SET usage_count = usage_count + 1 WHERE code = ?`, code)
		return true, nil
	}
	return false, nil
}

// GetAPICredentials retrieves API credentials
func (r *Repository) GetAPICredentials(ctx context.Context, apiID int) (string, error) {
	var apiHash string
	err := r.db.QueryRowContext(ctx, `
		SELECT api_hash FROM api_credentials WHERE api_id = ? AND is_active = 1
	`, apiID).Scan(&apiHash)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return apiHash, err
}

// Helper function to generate access hash
func generateAccessHash() int64 {
	max := big.NewInt(1 << 62)
	n, _ := rand.Int(rand.Reader, max)
	return n.Int64()
}

// Helper function to generate random hex string
func generateRandomHex(length int) string {
	bytes := make([]byte, length/2)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

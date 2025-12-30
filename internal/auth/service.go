package auth

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
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

// SendCode sends a verification code to the phone number
func (s *Service) SendCode(ctx context.Context, phone string, apiID int, apiHash string) (*SendCodeResult, error) {
	// Verify API credentials
	storedHash, err := s.repo.GetAPICredentials(ctx, apiID)
	if err != nil {
		return nil, fmt.Errorf("failed to get API credentials: %w", err)
	}
	if storedHash == "" || storedHash != apiHash {
		return nil, fmt.Errorf("invalid API credentials")
	}

	// Generate verification code
	code := generateVerificationCode()

	// Store code in Redis with 5 minute expiration
	if err := s.redis.SetAuthCode(ctx, phone, code, 5*time.Minute); err != nil {
		return nil, fmt.Errorf("failed to store auth code: %w", err)
	}

	// Generate phone code hash
	phoneCodeHash := generatePhoneCodeHash()

	// Check if user exists
	user, err := s.repo.GetUserByPhone(ctx, phone)
	if err != nil {
		return nil, fmt.Errorf("failed to check user: %w", err)
	}

	return &SendCodeResult{
		PhoneCodeHash: phoneCodeHash,
		Type: CodeType{
			Type:   "auth.sentCodeTypeSms",
			Length: 6,
		},
		NextType: "auth.codeTypeSms",
		Timeout:  120,
		IsNew:    user == nil,
	}, nil
}

// SignIn signs in an existing user
func (s *Service) SignIn(ctx context.Context, phone, phoneCodeHash, phoneCode string) (*models.User, error) {
	// Check universal code first
	isUniversal, err := s.repo.CheckUniversalCode(ctx, phoneCode)
	if err != nil {
		return nil, fmt.Errorf("failed to check universal code: %w", err)
	}

	if !isUniversal {
		// Verify the code from Redis
		storedCode, err := s.redis.GetAuthCode(ctx, phone)
		if err != nil {
			return nil, fmt.Errorf("code expired or not found")
		}
		if storedCode != phoneCode {
			return nil, fmt.Errorf("invalid code")
		}
	}

	// Delete the code after successful verification
	s.redis.DeleteAuthCode(ctx, phone)

	// Get user
	user, err := s.repo.GetUserByPhone(ctx, phone)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not registered")
	}

	// Check if user is banned
	if user.IsBanned {
		if user.BanExpiresAt.Valid && user.BanExpiresAt.Time.After(time.Now()) {
			return nil, fmt.Errorf("user is banned until %v: %s", user.BanExpiresAt.Time, user.BanReason.String)
		}
	}

	return user, nil
}

// SignUp registers a new user
func (s *Service) SignUp(ctx context.Context, phone, phoneCodeHash, firstName, lastName string) (*models.User, error) {
	// Check if user already exists
	existingUser, err := s.repo.GetUserByPhone(ctx, phone)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if existingUser != nil {
		return nil, fmt.Errorf("user already exists")
	}

	// Create new user
	user, err := s.repo.CreateUser(ctx, phone, firstName, lastName)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// BindAuthKey binds an auth key to a user
func (s *Service) BindAuthKey(ctx context.Context, authKeyID string, userID int64) error {
	return s.repo.BindAuthKeyToUser(ctx, authKeyID, userID)
}

// CreateSession creates a new session for a user
func (s *Service) CreateSession(ctx context.Context, authKeyID string, userID int64) error {
	sessionID := generateSessionID()
	return s.repo.CreateSession(ctx, authKeyID, userID, sessionID)
}

// GetUserByAuthKey gets the user associated with an auth key
func (s *Service) GetUserByAuthKey(ctx context.Context, authKeyID string) (*models.User, error) {
	userID, err := s.repo.GetSession(ctx, authKeyID)
	if err != nil {
		return nil, err
	}
	if userID == nil {
		return nil, nil
	}
	return s.repo.GetUserByID(ctx, *userID)
}

// UpdateUserOnline updates user's online status
func (s *Service) UpdateUserOnline(ctx context.Context, userID int64) error {
	return s.redis.SetUserOnline(ctx, userID, 5*time.Minute)
}

// Types
type SendCodeResult struct {
	PhoneCodeHash string   `json:"phone_code_hash"`
	Type          CodeType `json:"type"`
	NextType      string   `json:"next_type"`
	Timeout       int      `json:"timeout"`
	IsNew         bool     `json:"is_new"`
}

type CodeType struct {
	Type   string `json:"_"`
	Length int    `json:"length"`
}

// Helper functions
func generateVerificationCode() string {
	max := big.NewInt(1000000)
	n, _ := rand.Int(rand.Reader, max)
	return fmt.Sprintf("%06d", n.Int64())
}

func generatePhoneCodeHash() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return fmt.Sprintf("%x", bytes)
}

func generateSessionID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return fmt.Sprintf("%x", bytes)
}

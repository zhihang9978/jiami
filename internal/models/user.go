package models

import (
	"database/sql"
	"time"
)

type User struct {
	ID            int64          `json:"id" db:"id"`
	Phone         string         `json:"phone" db:"phone"`
	Username      sql.NullString `json:"username" db:"username"`
	FirstName     string         `json:"first_name" db:"first_name"`
	LastName      string         `json:"last_name" db:"last_name"`
	Bio           sql.NullString `json:"bio" db:"bio"`
	PhotoID       sql.NullInt64  `json:"photo_id" db:"photo_id"`
	AccessHash    int64          `json:"access_hash" db:"access_hash"`
	StatusType    string         `json:"status_type" db:"status_type"`
	StatusExpires int            `json:"status_expires" db:"status_expires"`
	PasswordHash  sql.NullString `json:"-" db:"password_hash"`
	IsBot         bool           `json:"is_bot" db:"is_bot"`
	IsVerified    bool           `json:"is_verified" db:"is_verified"`
	IsPremium     bool           `json:"is_premium" db:"is_premium"`
	IsBanned      bool           `json:"is_banned" db:"is_banned"`
	BanReason     sql.NullString `json:"ban_reason" db:"ban_reason"`
	BanExpiresAt  sql.NullTime   `json:"ban_expires_at" db:"ban_expires_at"`
	CreatedAt     time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at" db:"updated_at"`
	LastLoginAt   sql.NullTime   `json:"last_login_at" db:"last_login_at"`
	LastLoginIP   sql.NullString `json:"last_login_ip" db:"last_login_ip"`
}

// ToTLUser converts to TL User format for API responses
func (u *User) ToTLUser() map[string]interface{} {
	result := map[string]interface{}{
		"_":           "user",
		"id":          u.ID,
		"access_hash": u.AccessHash,
		"first_name":  u.FirstName,
		"phone":       u.Phone,
	}

	if u.Username.Valid {
		result["username"] = u.Username.String
	}
	if u.LastName != "" {
		result["last_name"] = u.LastName
	}
	if u.Bio.Valid {
		result["about"] = u.Bio.String
	}
	if u.PhotoID.Valid {
		result["photo"] = map[string]interface{}{
			"_":        "userProfilePhoto",
			"photo_id": u.PhotoID.Int64,
			"dc_id":    1,
		}
	}

	// Status
	switch u.StatusType {
	case "online":
		result["status"] = map[string]interface{}{
			"_":       "userStatusOnline",
			"expires": u.StatusExpires,
		}
	case "offline":
		result["status"] = map[string]interface{}{
			"_":        "userStatusOffline",
			"was_online": u.StatusExpires,
		}
	case "recently":
		result["status"] = map[string]interface{}{
			"_": "userStatusRecently",
		}
	}

	// Flags
	if u.IsBot {
		result["bot"] = true
	}
	if u.IsVerified {
		result["verified"] = true
	}
	if u.IsPremium {
		result["premium"] = true
	}

	return result
}

type UserStatus struct {
	Type      string `json:"type"`
	WasOnline int    `json:"was_online,omitempty"`
	Expires   int    `json:"expires,omitempty"`
}

package channels

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"
)

type Channel struct {
	ID                  int64
	AccessHash          int64
	Title               string
	Username            *string
	PhotoID             *int64
	Date                int
	Version             int
	CreatorID           int64
	ParticipantsCount   int
	IsBroadcast         bool
	IsMegagroup         bool
	IsVerified          bool
	IsRestricted        bool
	IsSignatures        bool
	IsSlowmodeEnabled   bool
	SlowmodeSeconds     int
	About               string
	AdminRights         json.RawMessage
	BannedRights        json.RawMessage
	DefaultBannedRights json.RawMessage
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type ChannelParticipant struct {
	ID              int64
	ChannelID       int64
	UserID          int64
	InviterID       *int64
	Date            int
	ParticipantType string // creator, admin, member, banned, left
	AdminRights     json.RawMessage
	BannedRights    json.RawMessage
	Rank            *string
	CreatedAt       time.Time
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// CreateChannel creates a new channel
func (r *Repository) CreateChannel(ctx context.Context, channel *Channel) error {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO channels (access_hash, title, username, photo_id, date, version, creator_id,
		                      participants_count, is_broadcast, is_megagroup, is_verified, is_restricted,
		                      is_signatures, is_slowmode_enabled, slowmode_seconds, about,
		                      admin_rights, banned_rights, default_banned_rights, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`, channel.AccessHash, channel.Title, channel.Username, channel.PhotoID, channel.Date, channel.Version,
		channel.CreatorID, channel.ParticipantsCount, channel.IsBroadcast, channel.IsMegagroup,
		channel.IsVerified, channel.IsRestricted, channel.IsSignatures, channel.IsSlowmodeEnabled,
		channel.SlowmodeSeconds, channel.About, channel.AdminRights, channel.BannedRights, channel.DefaultBannedRights)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	channel.ID = id
	return nil
}

// GetChannel retrieves a channel by ID
func (r *Repository) GetChannel(ctx context.Context, channelID int64) (*Channel, error) {
	channel := &Channel{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, access_hash, title, username, photo_id, date, version, creator_id,
		       participants_count, is_broadcast, is_megagroup, is_verified, is_restricted,
		       is_signatures, is_slowmode_enabled, slowmode_seconds, about,
		       admin_rights, banned_rights, default_banned_rights, created_at, updated_at
		FROM channels
		WHERE id = ?
	`, channelID).Scan(&channel.ID, &channel.AccessHash, &channel.Title, &channel.Username, &channel.PhotoID,
		&channel.Date, &channel.Version, &channel.CreatorID, &channel.ParticipantsCount,
		&channel.IsBroadcast, &channel.IsMegagroup, &channel.IsVerified, &channel.IsRestricted,
		&channel.IsSignatures, &channel.IsSlowmodeEnabled, &channel.SlowmodeSeconds, &channel.About,
		&channel.AdminRights, &channel.BannedRights, &channel.DefaultBannedRights, &channel.CreatedAt, &channel.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return channel, nil
}

// GetChannelByUsername retrieves a channel by username
func (r *Repository) GetChannelByUsername(ctx context.Context, username string) (*Channel, error) {
	channel := &Channel{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, access_hash, title, username, photo_id, date, version, creator_id,
		       participants_count, is_broadcast, is_megagroup, is_verified, is_restricted,
		       is_signatures, is_slowmode_enabled, slowmode_seconds, about,
		       admin_rights, banned_rights, default_banned_rights, created_at, updated_at
		FROM channels
		WHERE username = ?
	`, username).Scan(&channel.ID, &channel.AccessHash, &channel.Title, &channel.Username, &channel.PhotoID,
		&channel.Date, &channel.Version, &channel.CreatorID, &channel.ParticipantsCount,
		&channel.IsBroadcast, &channel.IsMegagroup, &channel.IsVerified, &channel.IsRestricted,
		&channel.IsSignatures, &channel.IsSlowmodeEnabled, &channel.SlowmodeSeconds, &channel.About,
		&channel.AdminRights, &channel.BannedRights, &channel.DefaultBannedRights, &channel.CreatedAt, &channel.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return channel, nil
}

// UpdateChannel updates a channel
func (r *Repository) UpdateChannel(ctx context.Context, channel *Channel) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE channels
		SET title = ?, username = ?, photo_id = ?, version = version + 1,
		    is_signatures = ?, is_slowmode_enabled = ?, slowmode_seconds = ?, about = ?,
		    admin_rights = ?, banned_rights = ?, default_banned_rights = ?, updated_at = NOW()
		WHERE id = ?
	`, channel.Title, channel.Username, channel.PhotoID, channel.IsSignatures,
		channel.IsSlowmodeEnabled, channel.SlowmodeSeconds, channel.About,
		channel.AdminRights, channel.BannedRights, channel.DefaultBannedRights, channel.ID)
	return err
}

// DeleteChannel deletes a channel
func (r *Repository) DeleteChannel(ctx context.Context, channelID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM channels WHERE id = ?`, channelID)
	return err
}

// AddParticipant adds a user to a channel
func (r *Repository) AddParticipant(ctx context.Context, participant *ChannelParticipant) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO channel_participants (channel_id, user_id, inviter_id, date, participant_type,
		                                  admin_rights, banned_rights, ` + "`rank`" + `, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW())
		ON DUPLICATE KEY UPDATE participant_type = VALUES(participant_type), date = VALUES(date)
	`, participant.ChannelID, participant.UserID, participant.InviterID, participant.Date,
		participant.ParticipantType, participant.AdminRights, participant.BannedRights, participant.Rank)
	return err
}

// RemoveParticipant removes a user from a channel
func (r *Repository) RemoveParticipant(ctx context.Context, channelID, userID int64) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM channel_participants WHERE channel_id = ? AND user_id = ?
	`, channelID, userID)
	return err
}

// GetParticipant retrieves a participant
func (r *Repository) GetParticipant(ctx context.Context, channelID, userID int64) (*ChannelParticipant, error) {
	p := &ChannelParticipant{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, channel_id, user_id, inviter_id, date, participant_type, admin_rights, banned_rights, `+"`rank`"+`, created_at
		FROM channel_participants
		WHERE channel_id = ? AND user_id = ?
	`, channelID, userID).Scan(&p.ID, &p.ChannelID, &p.UserID, &p.InviterID, &p.Date,
		&p.ParticipantType, &p.AdminRights, &p.BannedRights, &p.Rank, &p.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return p, nil
}

// GetParticipants retrieves participants with pagination
func (r *Repository) GetParticipants(ctx context.Context, channelID int64, filter string, offset, limit int) ([]*ChannelParticipant, error) {
	query := `
		SELECT id, channel_id, user_id, inviter_id, date, participant_type, admin_rights, banned_rights, ` + "`rank`" + `, created_at
		FROM channel_participants
		WHERE channel_id = ?`

	args := []interface{}{channelID}

	switch filter {
	case "admins":
		query += ` AND participant_type IN ('creator', 'admin')`
	case "banned":
		query += ` AND participant_type = 'banned'`
	case "bots":
		// Would need to join with users table
	}

	query += ` ORDER BY date LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var participants []*ChannelParticipant
	for rows.Next() {
		p := &ChannelParticipant{}
		if err := rows.Scan(&p.ID, &p.ChannelID, &p.UserID, &p.InviterID, &p.Date,
			&p.ParticipantType, &p.AdminRights, &p.BannedRights, &p.Rank, &p.CreatedAt); err != nil {
			return nil, err
		}
		participants = append(participants, p)
	}
	return participants, nil
}

// UpdateParticipant updates a participant
func (r *Repository) UpdateParticipant(ctx context.Context, participant *ChannelParticipant) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE channel_participants
		SET participant_type = ?, admin_rights = ?, banned_rights = ?, `+"`rank`"+` = ?
		WHERE channel_id = ? AND user_id = ?
	`, participant.ParticipantType, participant.AdminRights, participant.BannedRights,
		participant.Rank, participant.ChannelID, participant.UserID)
	return err
}

// IncrementParticipantsCount increments the participants count
func (r *Repository) IncrementParticipantsCount(ctx context.Context, channelID int64, delta int) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE channels SET participants_count = participants_count + ?, updated_at = NOW() WHERE id = ?
	`, delta, channelID)
	return err
}

// GetUserChannels retrieves all channels a user is a member of
func (r *Repository) GetUserChannels(ctx context.Context, userID int64) ([]*Channel, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT c.id, c.access_hash, c.title, c.username, c.photo_id, c.date, c.version, c.creator_id,
		       c.participants_count, c.is_broadcast, c.is_megagroup, c.is_verified, c.is_restricted,
		       c.is_signatures, c.is_slowmode_enabled, c.slowmode_seconds, c.about,
		       c.admin_rights, c.banned_rights, c.default_banned_rights, c.created_at, c.updated_at
		FROM channels c
		JOIN channel_participants cp ON c.id = cp.channel_id
		WHERE cp.user_id = ? AND cp.participant_type != 'left' AND cp.participant_type != 'banned'
		ORDER BY c.updated_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []*Channel
	for rows.Next() {
		channel := &Channel{}
		if err := rows.Scan(&channel.ID, &channel.AccessHash, &channel.Title, &channel.Username, &channel.PhotoID,
			&channel.Date, &channel.Version, &channel.CreatorID, &channel.ParticipantsCount,
			&channel.IsBroadcast, &channel.IsMegagroup, &channel.IsVerified, &channel.IsRestricted,
			&channel.IsSignatures, &channel.IsSlowmodeEnabled, &channel.SlowmodeSeconds, &channel.About,
			&channel.AdminRights, &channel.BannedRights, &channel.DefaultBannedRights, &channel.CreatedAt, &channel.UpdatedAt); err != nil {
			return nil, err
		}
		channels = append(channels, channel)
	}
	return channels, nil
}

// SearchChannels searches for public channels
func (r *Repository) SearchChannels(ctx context.Context, query string, limit int) ([]*Channel, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, access_hash, title, username, photo_id, date, version, creator_id,
		       participants_count, is_broadcast, is_megagroup, is_verified, is_restricted,
		       is_signatures, is_slowmode_enabled, slowmode_seconds, about,
		       admin_rights, banned_rights, default_banned_rights, created_at, updated_at
		FROM channels
		WHERE username IS NOT NULL AND (title LIKE ? OR username LIKE ?)
		ORDER BY participants_count DESC
		LIMIT ?
	`, "%"+query+"%", "%"+query+"%", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []*Channel
	for rows.Next() {
		channel := &Channel{}
		if err := rows.Scan(&channel.ID, &channel.AccessHash, &channel.Title, &channel.Username, &channel.PhotoID,
			&channel.Date, &channel.Version, &channel.CreatorID, &channel.ParticipantsCount,
			&channel.IsBroadcast, &channel.IsMegagroup, &channel.IsVerified, &channel.IsRestricted,
			&channel.IsSignatures, &channel.IsSlowmodeEnabled, &channel.SlowmodeSeconds, &channel.About,
			&channel.AdminRights, &channel.BannedRights, &channel.DefaultBannedRights, &channel.CreatedAt, &channel.UpdatedAt); err != nil {
			return nil, err
		}
		channels = append(channels, channel)
	}
	return channels, nil
}

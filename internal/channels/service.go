package channels

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// generateAccessHash generates a random access hash
func generateAccessHash() int64 {
	b := make([]byte, 8)
	rand.Read(b)
	var hash int64
	for i := 0; i < 8; i++ {
		hash = (hash << 8) | int64(b[i])
	}
	if hash < 0 {
		hash = -hash
	}
	return hash
}

// CreateChannel creates a new channel or supergroup
func (s *Service) CreateChannel(ctx context.Context, creatorID int64, title string, about string, isBroadcast, isMegagroup bool) (*Channel, error) {
	channel := &Channel{
		AccessHash:        generateAccessHash(),
		Title:             title,
		Date:              int(time.Now().Unix()),
		Version:           1,
		CreatorID:         creatorID,
		ParticipantsCount: 1,
		IsBroadcast:       isBroadcast,
		IsMegagroup:       isMegagroup,
		About:             about,
	}

	if err := s.repo.CreateChannel(ctx, channel); err != nil {
		return nil, fmt.Errorf("failed to create channel: %w", err)
	}

	// Add creator
	creatorParticipant := &ChannelParticipant{
		ChannelID:       channel.ID,
		UserID:          creatorID,
		Date:            int(time.Now().Unix()),
		ParticipantType: "creator",
	}
	if err := s.repo.AddParticipant(ctx, creatorParticipant); err != nil {
		return nil, fmt.Errorf("failed to add creator: %w", err)
	}

	return channel, nil
}

// GetChannel retrieves a channel by ID
func (s *Service) GetChannel(ctx context.Context, channelID int64) (*Channel, error) {
	return s.repo.GetChannel(ctx, channelID)
}

// GetChannelByUsername retrieves a channel by username
func (s *Service) GetChannelByUsername(ctx context.Context, username string) (*Channel, error) {
	return s.repo.GetChannelByUsername(ctx, username)
}

// EditTitle updates the channel title
func (s *Service) EditTitle(ctx context.Context, channelID, userID int64, title string) (*Channel, error) {
	channel, err := s.repo.GetChannel(ctx, channelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get channel: %w", err)
	}
	if channel == nil {
		return nil, fmt.Errorf("channel not found")
	}

	// Check permissions
	if err := s.checkAdminPermission(ctx, channelID, userID); err != nil {
		return nil, err
	}

	channel.Title = title
	if err := s.repo.UpdateChannel(ctx, channel); err != nil {
		return nil, fmt.Errorf("failed to update channel: %w", err)
	}

	return channel, nil
}

// EditAbout updates the channel about
func (s *Service) EditAbout(ctx context.Context, channelID, userID int64, about string) (*Channel, error) {
	channel, err := s.repo.GetChannel(ctx, channelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get channel: %w", err)
	}
	if channel == nil {
		return nil, fmt.Errorf("channel not found")
	}

	// Check permissions
	if err := s.checkAdminPermission(ctx, channelID, userID); err != nil {
		return nil, err
	}

	channel.About = about
	if err := s.repo.UpdateChannel(ctx, channel); err != nil {
		return nil, fmt.Errorf("failed to update channel: %w", err)
	}

	return channel, nil
}

// UpdateUsername updates the channel username
func (s *Service) UpdateUsername(ctx context.Context, channelID, userID int64, username string) (*Channel, error) {
	channel, err := s.repo.GetChannel(ctx, channelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get channel: %w", err)
	}
	if channel == nil {
		return nil, fmt.Errorf("channel not found")
	}

	// Check permissions (only creator)
	participant, err := s.repo.GetParticipant(ctx, channelID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get participant: %w", err)
	}
	if participant == nil || participant.ParticipantType != "creator" {
		return nil, fmt.Errorf("only creator can change username")
	}

	// Check if username is available
	existing, err := s.repo.GetChannelByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to check username: %w", err)
	}
	if existing != nil && existing.ID != channelID {
		return nil, fmt.Errorf("username already taken")
	}

	channel.Username = &username
	if err := s.repo.UpdateChannel(ctx, channel); err != nil {
		return nil, fmt.Errorf("failed to update channel: %w", err)
	}

	return channel, nil
}

// JoinChannel joins a public channel
func (s *Service) JoinChannel(ctx context.Context, channelID, userID int64) (*Channel, error) {
	channel, err := s.repo.GetChannel(ctx, channelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get channel: %w", err)
	}
	if channel == nil {
		return nil, fmt.Errorf("channel not found")
	}

	// Check if already a member
	existing, err := s.repo.GetParticipant(ctx, channelID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check membership: %w", err)
	}
	if existing != nil && existing.ParticipantType != "left" && existing.ParticipantType != "banned" {
		return channel, nil // Already a member
	}

	// Add participant
	participant := &ChannelParticipant{
		ChannelID:       channelID,
		UserID:          userID,
		Date:            int(time.Now().Unix()),
		ParticipantType: "member",
	}
	if err := s.repo.AddParticipant(ctx, participant); err != nil {
		return nil, fmt.Errorf("failed to add participant: %w", err)
	}

	// Update count
	if err := s.repo.IncrementParticipantsCount(ctx, channelID, 1); err != nil {
		return nil, fmt.Errorf("failed to update count: %w", err)
	}

	channel.ParticipantsCount++
	return channel, nil
}

// LeaveChannel leaves a channel
func (s *Service) LeaveChannel(ctx context.Context, channelID, userID int64) (*Channel, error) {
	channel, err := s.repo.GetChannel(ctx, channelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get channel: %w", err)
	}
	if channel == nil {
		return nil, fmt.Errorf("channel not found")
	}

	// Check if creator (creator cannot leave)
	participant, err := s.repo.GetParticipant(ctx, channelID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get participant: %w", err)
	}
	if participant == nil {
		return nil, fmt.Errorf("not a member")
	}
	if participant.ParticipantType == "creator" {
		return nil, fmt.Errorf("creator cannot leave channel")
	}

	// Remove participant
	if err := s.repo.RemoveParticipant(ctx, channelID, userID); err != nil {
		return nil, fmt.Errorf("failed to remove participant: %w", err)
	}

	// Update count
	if err := s.repo.IncrementParticipantsCount(ctx, channelID, -1); err != nil {
		return nil, fmt.Errorf("failed to update count: %w", err)
	}

	channel.ParticipantsCount--
	return channel, nil
}

// InviteToChannel invites a user to a channel
func (s *Service) InviteToChannel(ctx context.Context, channelID, inviterID, userID int64) (*Channel, error) {
	channel, err := s.repo.GetChannel(ctx, channelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get channel: %w", err)
	}
	if channel == nil {
		return nil, fmt.Errorf("channel not found")
	}

	// Check if inviter has permission
	if err := s.checkAdminPermission(ctx, channelID, inviterID); err != nil {
		return nil, err
	}

	// Check if already a member
	existing, err := s.repo.GetParticipant(ctx, channelID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check membership: %w", err)
	}
	if existing != nil && existing.ParticipantType != "left" && existing.ParticipantType != "banned" {
		return channel, nil // Already a member
	}

	// Add participant
	participant := &ChannelParticipant{
		ChannelID:       channelID,
		UserID:          userID,
		InviterID:       &inviterID,
		Date:            int(time.Now().Unix()),
		ParticipantType: "member",
	}
	if err := s.repo.AddParticipant(ctx, participant); err != nil {
		return nil, fmt.Errorf("failed to add participant: %w", err)
	}

	// Update count
	if err := s.repo.IncrementParticipantsCount(ctx, channelID, 1); err != nil {
		return nil, fmt.Errorf("failed to update count: %w", err)
	}

	channel.ParticipantsCount++
	return channel, nil
}

// KickFromChannel kicks a user from a channel
func (s *Service) KickFromChannel(ctx context.Context, channelID, adminID, userID int64) (*Channel, error) {
	channel, err := s.repo.GetChannel(ctx, channelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get channel: %w", err)
	}
	if channel == nil {
		return nil, fmt.Errorf("channel not found")
	}

	// Check if admin has permission
	if err := s.checkAdminPermission(ctx, channelID, adminID); err != nil {
		return nil, err
	}

	// Cannot kick creator
	target, err := s.repo.GetParticipant(ctx, channelID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get target: %w", err)
	}
	if target == nil {
		return nil, fmt.Errorf("user is not a member")
	}
	if target.ParticipantType == "creator" {
		return nil, fmt.Errorf("cannot kick creator")
	}

	// Remove participant
	if err := s.repo.RemoveParticipant(ctx, channelID, userID); err != nil {
		return nil, fmt.Errorf("failed to remove participant: %w", err)
	}

	// Update count
	if err := s.repo.IncrementParticipantsCount(ctx, channelID, -1); err != nil {
		return nil, fmt.Errorf("failed to update count: %w", err)
	}

	channel.ParticipantsCount--
	return channel, nil
}

// GetParticipants retrieves channel participants
func (s *Service) GetParticipants(ctx context.Context, channelID int64, filter string, offset, limit int) ([]*ChannelParticipant, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.repo.GetParticipants(ctx, channelID, filter, offset, limit)
}

// GetFullChannel retrieves full channel info
func (s *Service) GetFullChannel(ctx context.Context, channelID int64) (*FullChannel, error) {
	channel, err := s.repo.GetChannel(ctx, channelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get channel: %w", err)
	}
	if channel == nil {
		return nil, fmt.Errorf("channel not found")
	}

	return &FullChannel{
		Channel: channel,
	}, nil
}

// GetUserChannels retrieves all channels a user is a member of
func (s *Service) GetUserChannels(ctx context.Context, userID int64) ([]*Channel, error) {
	return s.repo.GetUserChannels(ctx, userID)
}

// SearchChannels searches for public channels
func (s *Service) SearchChannels(ctx context.Context, query string, limit int) ([]*Channel, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.repo.SearchChannels(ctx, query, limit)
}

// checkAdminPermission checks if user has admin permission
func (s *Service) checkAdminPermission(ctx context.Context, channelID, userID int64) error {
	participant, err := s.repo.GetParticipant(ctx, channelID, userID)
	if err != nil {
		return fmt.Errorf("failed to get participant: %w", err)
	}
	if participant == nil {
		return fmt.Errorf("not a member")
	}
	if participant.ParticipantType != "creator" && participant.ParticipantType != "admin" {
		return fmt.Errorf("permission denied")
	}
	return nil
}

// FullChannel represents full channel info
type FullChannel struct {
	Channel *Channel
}

// ToTL converts Channel to TL format
func (c *Channel) ToTL() map[string]interface{} {
	result := map[string]interface{}{
		"_":                    "channel",
		"id":                   c.ID,
		"access_hash":          c.AccessHash,
		"title":                c.Title,
		"photo":                c.getPhotoTL(),
		"date":                 c.Date,
		"version":              c.Version,
		"participants_count":   c.ParticipantsCount,
		"broadcast":            c.IsBroadcast,
		"megagroup":            c.IsMegagroup,
		"verified":             c.IsVerified,
		"restricted":           c.IsRestricted,
		"signatures":           c.IsSignatures,
		"slowmode_enabled":     c.IsSlowmodeEnabled,
		"slowmode_seconds":     c.SlowmodeSeconds,
	}
	if c.Username != nil {
		result["username"] = *c.Username
	}
	return result
}

func (c *Channel) getPhotoTL() map[string]interface{} {
	if c.PhotoID == nil {
		return map[string]interface{}{"_": "chatPhotoEmpty"}
	}
	return map[string]interface{}{
		"_":        "chatPhoto",
		"photo_id": *c.PhotoID,
	}
}

// ToTL converts FullChannel to TL format
func (fc *FullChannel) ToTL() map[string]interface{} {
	return map[string]interface{}{
		"_":            "messages.chatFull",
		"full_chat":    fc.Channel.ToTL(),
		"chats":        []interface{}{fc.Channel.ToTL()},
		"users":        []interface{}{},
		"about":        fc.Channel.About,
	}
}

// ToTL converts ChannelParticipant to TL format
func (p *ChannelParticipant) ToTL() map[string]interface{} {
	result := map[string]interface{}{
		"user_id": p.UserID,
		"date":    p.Date,
	}

	switch p.ParticipantType {
	case "creator":
		result["_"] = "channelParticipantCreator"
	case "admin":
		result["_"] = "channelParticipantAdmin"
		if p.InviterID != nil {
			result["promoted_by"] = *p.InviterID
		}
		result["admin_rights"] = p.AdminRights
		if p.Rank != nil {
			result["rank"] = *p.Rank
		}
	case "banned":
		result["_"] = "channelParticipantBanned"
		result["banned_rights"] = p.BannedRights
	default:
		result["_"] = "channelParticipant"
	}

	return result
}

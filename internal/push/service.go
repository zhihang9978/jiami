package push

import (
	"context"
	"fmt"
	"time"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// RegisterDevice registers a device for push notifications
func (s *Service) RegisterDevice(ctx context.Context, userID int64, token string, tokenType int, deviceModel, systemVersion, appVersion string, appSandbox bool, secret []byte, noMuted bool) (*Device, error) {
	device := &Device{
		UserID:        userID,
		Token:         token,
		TokenType:     tokenType,
		DeviceModel:   deviceModel,
		SystemVersion: systemVersion,
		AppVersion:    appVersion,
		AppSandbox:    appSandbox,
		Secret:        secret,
		NoMuted:       noMuted,
	}

	if err := s.repo.RegisterDevice(ctx, device); err != nil {
		return nil, fmt.Errorf("failed to register device: %w", err)
	}

	return device, nil
}

// UnregisterDevice removes a device
func (s *Service) UnregisterDevice(ctx context.Context, token string) error {
	return s.repo.UnregisterDevice(ctx, token)
}

// GetUserDevices retrieves all devices for a user
func (s *Service) GetUserDevices(ctx context.Context, userID int64) ([]*Device, error) {
	return s.repo.GetUserDevices(ctx, userID)
}

// GetNotificationSettings retrieves notification settings for a peer
func (s *Service) GetNotificationSettings(ctx context.Context, userID, peerID int64, peerType string) (*NotificationSettings, error) {
	settings, err := s.repo.GetNotificationSettings(ctx, userID, peerID, peerType)
	if err != nil {
		return nil, err
	}
	if settings == nil {
		// Return default settings
		return &NotificationSettings{
			UserID:       userID,
			PeerID:       peerID,
			PeerType:     peerType,
			ShowPreviews: true,
			Silent:       false,
			MuteUntil:    0,
			Sound:        "default",
		}, nil
	}
	return settings, nil
}

// UpdateNotificationSettings updates notification settings
func (s *Service) UpdateNotificationSettings(ctx context.Context, settings *NotificationSettings) error {
	return s.repo.UpdateNotificationSettings(ctx, settings)
}

// ResetNotificationSettings resets notification settings to default
func (s *Service) ResetNotificationSettings(ctx context.Context, userID, peerID int64, peerType string) error {
	return s.repo.ResetNotificationSettings(ctx, userID, peerID, peerType)
}

// GetAllNotificationSettings retrieves all notification settings for a user
func (s *Service) GetAllNotificationSettings(ctx context.Context, userID int64) ([]*NotificationSettings, error) {
	return s.repo.GetAllNotificationSettings(ctx, userID)
}

// SendPushNotification sends a push notification to a user
func (s *Service) SendPushNotification(ctx context.Context, userID int64, title, body string, data map[string]interface{}) error {
	devices, err := s.repo.GetUserDevices(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get devices: %w", err)
	}

	for _, device := range devices {
		// In a real implementation, this would send to APNS/FCM/etc.
		// For now, we just log the notification
		_ = device
		_ = title
		_ = body
		_ = data
	}

	return nil
}

// MuteUntil mutes notifications for a peer until a specific time
func (s *Service) MuteUntil(ctx context.Context, userID, peerID int64, peerType string, muteUntil int) error {
	settings := &NotificationSettings{
		UserID:       userID,
		PeerID:       peerID,
		PeerType:     peerType,
		ShowPreviews: true,
		Silent:       false,
		MuteUntil:    muteUntil,
		Sound:        "default",
	}
	return s.repo.UpdateNotificationSettings(ctx, settings)
}

// Unmute unmutes notifications for a peer
func (s *Service) Unmute(ctx context.Context, userID, peerID int64, peerType string) error {
	return s.MuteUntil(ctx, userID, peerID, peerType, 0)
}

// IsMuted checks if notifications are muted for a peer
func (s *Service) IsMuted(ctx context.Context, userID, peerID int64, peerType string) (bool, error) {
	settings, err := s.repo.GetNotificationSettings(ctx, userID, peerID, peerType)
	if err != nil {
		return false, err
	}
	if settings == nil {
		return false, nil
	}
	if settings.MuteUntil == 0 {
		return false, nil
	}
	if settings.MuteUntil == 2147483647 { // Max int32 = forever
		return true, nil
	}
	return int64(settings.MuteUntil) > time.Now().Unix(), nil
}

// ToTL converts Device to TL format
func (d *Device) ToTL() map[string]interface{} {
	return map[string]interface{}{
		"_":              "account.device",
		"token":          d.Token,
		"token_type":     d.TokenType,
		"device_model":   d.DeviceModel,
		"system_version": d.SystemVersion,
		"app_version":    d.AppVersion,
		"app_sandbox":    d.AppSandbox,
		"no_muted":       d.NoMuted,
	}
}

// ToTL converts NotificationSettings to TL format
func (ns *NotificationSettings) ToTL() map[string]interface{} {
	return map[string]interface{}{
		"_":                   "peerNotifySettings",
		"show_previews":       ns.ShowPreviews,
		"silent":              ns.Silent,
		"mute_until":          ns.MuteUntil,
		"sound":               ns.Sound,
		"stories_muted":       ns.StoriesMuted,
		"stories_hide_sender": ns.StoriesHideSender,
		"stories_sound":       ns.StoriesSound,
	}
}

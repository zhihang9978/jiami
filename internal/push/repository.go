package push

import (
	"context"
	"database/sql"
	"time"
)

type Device struct {
	ID           int64
	UserID       int64
	Token        string
	TokenType    int // 1=APNS, 2=FCM, 3=MPNS, 4=SimplePush, 5=UbuntuPhone, 6=Blackberry, 7=APNS_VOIP, 8=WebPush, 9=MPNS_VOIP, 10=Tizen, 11=Firefox, 12=HuaweiPush
	DeviceModel  string
	SystemVersion string
	AppVersion   string
	AppSandbox   bool
	Secret       []byte
	OtherUIDs    []int64
	NoMuted      bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type NotificationSettings struct {
	ID              int64
	UserID          int64
	PeerID          int64
	PeerType        string
	ShowPreviews    bool
	Silent          bool
	MuteUntil       int
	Sound           string
	StoriesMuted    bool
	StoriesHideSender bool
	StoriesSound    string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// RegisterDevice registers a device for push notifications
func (r *Repository) RegisterDevice(ctx context.Context, device *Device) error {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO push_devices (user_id, token, token_type, device_model, system_version, 
		                          app_version, app_sandbox, secret, no_muted, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
		ON DUPLICATE KEY UPDATE 
			user_id = VALUES(user_id),
			token_type = VALUES(token_type),
			device_model = VALUES(device_model),
			system_version = VALUES(system_version),
			app_version = VALUES(app_version),
			app_sandbox = VALUES(app_sandbox),
			secret = VALUES(secret),
			no_muted = VALUES(no_muted),
			updated_at = NOW()
	`, device.UserID, device.Token, device.TokenType, device.DeviceModel, device.SystemVersion,
		device.AppVersion, device.AppSandbox, device.Secret, device.NoMuted)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	if id > 0 {
		device.ID = id
	}
	return nil
}

// UnregisterDevice removes a device
func (r *Repository) UnregisterDevice(ctx context.Context, token string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM push_devices WHERE token = ?`, token)
	return err
}

// GetUserDevices retrieves all devices for a user
func (r *Repository) GetUserDevices(ctx context.Context, userID int64) ([]*Device, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, token, token_type, device_model, system_version, 
		       app_version, app_sandbox, secret, no_muted, created_at, updated_at
		FROM push_devices
		WHERE user_id = ?
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []*Device
	for rows.Next() {
		d := &Device{}
		if err := rows.Scan(&d.ID, &d.UserID, &d.Token, &d.TokenType, &d.DeviceModel,
			&d.SystemVersion, &d.AppVersion, &d.AppSandbox, &d.Secret, &d.NoMuted,
			&d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		devices = append(devices, d)
	}
	return devices, nil
}

// GetDeviceByToken retrieves a device by token
func (r *Repository) GetDeviceByToken(ctx context.Context, token string) (*Device, error) {
	d := &Device{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, token, token_type, device_model, system_version, 
		       app_version, app_sandbox, secret, no_muted, created_at, updated_at
		FROM push_devices
		WHERE token = ?
	`, token).Scan(&d.ID, &d.UserID, &d.Token, &d.TokenType, &d.DeviceModel,
		&d.SystemVersion, &d.AppVersion, &d.AppSandbox, &d.Secret, &d.NoMuted,
		&d.CreatedAt, &d.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return d, nil
}

// GetNotificationSettings retrieves notification settings for a peer
func (r *Repository) GetNotificationSettings(ctx context.Context, userID, peerID int64, peerType string) (*NotificationSettings, error) {
	ns := &NotificationSettings{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, peer_id, peer_type, show_previews, silent, mute_until, sound,
		       stories_muted, stories_hide_sender, stories_sound, created_at, updated_at
		FROM notification_settings
		WHERE user_id = ? AND peer_id = ? AND peer_type = ?
	`, userID, peerID, peerType).Scan(&ns.ID, &ns.UserID, &ns.PeerID, &ns.PeerType,
		&ns.ShowPreviews, &ns.Silent, &ns.MuteUntil, &ns.Sound,
		&ns.StoriesMuted, &ns.StoriesHideSender, &ns.StoriesSound, &ns.CreatedAt, &ns.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return ns, nil
}

// UpdateNotificationSettings updates notification settings
func (r *Repository) UpdateNotificationSettings(ctx context.Context, settings *NotificationSettings) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO notification_settings (user_id, peer_id, peer_type, show_previews, silent, 
		                                   mute_until, sound, stories_muted, stories_hide_sender, 
		                                   stories_sound, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
		ON DUPLICATE KEY UPDATE
			show_previews = VALUES(show_previews),
			silent = VALUES(silent),
			mute_until = VALUES(mute_until),
			sound = VALUES(sound),
			stories_muted = VALUES(stories_muted),
			stories_hide_sender = VALUES(stories_hide_sender),
			stories_sound = VALUES(stories_sound),
			updated_at = NOW()
	`, settings.UserID, settings.PeerID, settings.PeerType, settings.ShowPreviews,
		settings.Silent, settings.MuteUntil, settings.Sound, settings.StoriesMuted,
		settings.StoriesHideSender, settings.StoriesSound)
	return err
}

// ResetNotificationSettings resets notification settings to default
func (r *Repository) ResetNotificationSettings(ctx context.Context, userID, peerID int64, peerType string) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM notification_settings WHERE user_id = ? AND peer_id = ? AND peer_type = ?
	`, userID, peerID, peerType)
	return err
}

// GetAllNotificationSettings retrieves all notification settings for a user
func (r *Repository) GetAllNotificationSettings(ctx context.Context, userID int64) ([]*NotificationSettings, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, peer_id, peer_type, show_previews, silent, mute_until, sound,
		       stories_muted, stories_hide_sender, stories_sound, created_at, updated_at
		FROM notification_settings
		WHERE user_id = ?
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var settings []*NotificationSettings
	for rows.Next() {
		ns := &NotificationSettings{}
		if err := rows.Scan(&ns.ID, &ns.UserID, &ns.PeerID, &ns.PeerType,
			&ns.ShowPreviews, &ns.Silent, &ns.MuteUntil, &ns.Sound,
			&ns.StoriesMuted, &ns.StoriesHideSender, &ns.StoriesSound, &ns.CreatedAt, &ns.UpdatedAt); err != nil {
			return nil, err
		}
		settings = append(settings, ns)
	}
	return settings, nil
}

-- Phase 2: Groups, Channels, Search, Push Notifications

-- Chats (Groups) table
CREATE TABLE IF NOT EXISTS chats (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    title VARCHAR(255) NOT NULL,
    photo_id BIGINT DEFAULT NULL,
    participants_count INT DEFAULT 0,
    date INT NOT NULL,
    version INT DEFAULT 1,
    creator_id BIGINT NOT NULL,
    is_deactivated TINYINT(1) DEFAULT 0,
    is_call_active TINYINT(1) DEFAULT 0,
    is_call_not_empty TINYINT(1) DEFAULT 0,
    migrated_to_channel_id BIGINT DEFAULT NULL,
    admin_rights JSON DEFAULT NULL,
    default_banned_rights JSON DEFAULT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_creator_id (creator_id),
    INDEX idx_updated_at (updated_at)
);

-- Chat participants table
CREATE TABLE IF NOT EXISTS chat_participants (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    chat_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    inviter_id BIGINT DEFAULT NULL,
    date INT NOT NULL,
    is_admin TINYINT(1) DEFAULT 0,
    admin_rights JSON DEFAULT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_chat_user (chat_id, user_id),
    INDEX idx_user_id (user_id),
    INDEX idx_chat_id (chat_id)
);

-- Channels table
CREATE TABLE IF NOT EXISTS channels (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    access_hash BIGINT NOT NULL,
    title VARCHAR(255) NOT NULL,
    username VARCHAR(64) DEFAULT NULL,
    photo_id BIGINT DEFAULT NULL,
    date INT NOT NULL,
    version INT DEFAULT 1,
    creator_id BIGINT NOT NULL,
    participants_count INT DEFAULT 0,
    is_broadcast TINYINT(1) DEFAULT 0,
    is_megagroup TINYINT(1) DEFAULT 0,
    is_verified TINYINT(1) DEFAULT 0,
    is_restricted TINYINT(1) DEFAULT 0,
    is_signatures TINYINT(1) DEFAULT 0,
    is_slowmode_enabled TINYINT(1) DEFAULT 0,
    slowmode_seconds INT DEFAULT 0,
    about TEXT DEFAULT NULL,
    admin_rights JSON DEFAULT NULL,
    banned_rights JSON DEFAULT NULL,
    default_banned_rights JSON DEFAULT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_username (username),
    INDEX idx_creator_id (creator_id),
    INDEX idx_updated_at (updated_at)
);

-- Channel participants table
CREATE TABLE IF NOT EXISTS channel_participants (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    channel_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    inviter_id BIGINT DEFAULT NULL,
    date INT NOT NULL,
    participant_type VARCHAR(32) DEFAULT 'member',
    admin_rights JSON DEFAULT NULL,
    banned_rights JSON DEFAULT NULL,
    `rank` VARCHAR(64) DEFAULT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_channel_user (channel_id, user_id),
    INDEX idx_user_id (user_id),
    INDEX idx_channel_id (channel_id),
    INDEX idx_participant_type (participant_type)
);

-- Search history table
CREATE TABLE IF NOT EXISTS search_history (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    query VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_user_query (user_id, query),
    INDEX idx_user_id (user_id),
    INDEX idx_created_at (created_at)
);

-- Push devices table
CREATE TABLE IF NOT EXISTS push_devices (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    token VARCHAR(512) NOT NULL,
    token_type INT NOT NULL,
    device_model VARCHAR(128) DEFAULT NULL,
    system_version VARCHAR(64) DEFAULT NULL,
    app_version VARCHAR(64) DEFAULT NULL,
    app_sandbox TINYINT(1) DEFAULT 0,
    secret BLOB DEFAULT NULL,
    no_muted TINYINT(1) DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_token (token),
    INDEX idx_user_id (user_id)
);

-- Notification settings table
CREATE TABLE IF NOT EXISTS notification_settings (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    peer_id BIGINT NOT NULL,
    peer_type VARCHAR(32) NOT NULL,
    show_previews TINYINT(1) DEFAULT 1,
    silent TINYINT(1) DEFAULT 0,
    mute_until INT DEFAULT 0,
    sound VARCHAR(64) DEFAULT 'default',
    stories_muted TINYINT(1) DEFAULT 0,
    stories_hide_sender TINYINT(1) DEFAULT 0,
    stories_sound VARCHAR(64) DEFAULT 'default',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_user_peer (user_id, peer_id, peer_type),
    INDEX idx_user_id (user_id)
);

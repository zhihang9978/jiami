-- FeijiIM Database Schema
-- Based on BACKEND_COMPLETE_SPECIFICATION.md
-- Version: 1.0
-- Created: 2024-12-30

SET NAMES utf8mb4;
SET CHARACTER SET utf8mb4;

-- =====================================================
-- Core Tables
-- =====================================================

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    phone VARCHAR(20) UNIQUE NOT NULL,
    username VARCHAR(32) UNIQUE,
    first_name VARCHAR(64) NOT NULL DEFAULT '',
    last_name VARCHAR(64) DEFAULT '',
    bio TEXT,
    photo_id BIGINT DEFAULT NULL,
    access_hash BIGINT NOT NULL,
    status_type VARCHAR(20) DEFAULT 'offline',
    status_expires INT DEFAULT 0,
    password_hash VARCHAR(255) DEFAULT NULL,
    is_bot TINYINT(1) DEFAULT 0,
    is_verified TINYINT(1) DEFAULT 0,
    is_premium TINYINT(1) DEFAULT 0,
    is_banned TINYINT(1) DEFAULT 0,
    ban_reason VARCHAR(255) DEFAULT NULL,
    ban_expires_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP NULL,
    last_login_ip VARCHAR(45) DEFAULT NULL,
    INDEX idx_phone (phone),
    INDEX idx_username (username),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Auth keys table (MTProto authentication)
CREATE TABLE IF NOT EXISTS auth_keys (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    auth_key_id VARCHAR(16) NOT NULL UNIQUE,
    auth_key BLOB NOT NULL,
    user_id BIGINT DEFAULT NULL,
    temp_auth_key_id VARCHAR(16) DEFAULT NULL,
    temp_auth_key BLOB DEFAULT NULL,
    temp_expires_at TIMESTAMP NULL,
    device_type VARCHAR(50) DEFAULT NULL,
    app_version VARCHAR(20) DEFAULT NULL,
    system_version VARCHAR(50) DEFAULT NULL,
    ip VARCHAR(45) DEFAULT NULL,
    country VARCHAR(10) DEFAULT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_activity TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_auth_key_id (auth_key_id),
    INDEX idx_user_id (user_id),
    INDEX idx_temp_auth_key_id (temp_auth_key_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Sessions table
CREATE TABLE IF NOT EXISTS sessions (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    auth_key_id VARCHAR(16) NOT NULL,
    user_id BIGINT NOT NULL,
    session_id VARCHAR(16) NOT NULL,
    pts INT DEFAULT 0,
    qts INT DEFAULT 0,
    seq INT DEFAULT 0,
    date INT DEFAULT 0,
    device_type VARCHAR(50) DEFAULT NULL,
    app_version VARCHAR(20) DEFAULT NULL,
    system_version VARCHAR(50) DEFAULT NULL,
    ip VARCHAR(45) DEFAULT NULL,
    country VARCHAR(10) DEFAULT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_activity TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_auth_key_id (auth_key_id),
    INDEX idx_user_id (user_id),
    INDEX idx_session_id (session_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- =====================================================
-- Messaging Tables
-- =====================================================

-- Messages table
CREATE TABLE IF NOT EXISTS messages (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    message_id INT NOT NULL,
    from_id BIGINT NOT NULL,
    peer_id BIGINT NOT NULL,
    peer_type ENUM('user', 'chat', 'channel') NOT NULL DEFAULT 'user',
    message TEXT,
    date INT NOT NULL,
    random_id BIGINT DEFAULT NULL,
    reply_to_msg_id INT DEFAULT NULL,
    fwd_from_id BIGINT DEFAULT NULL,
    fwd_date INT DEFAULT NULL,
    via_bot_id BIGINT DEFAULT NULL,
    edit_date INT DEFAULT NULL,
    media_type VARCHAR(50) DEFAULT NULL,
    media_id BIGINT DEFAULT NULL,
    entities JSON DEFAULT NULL,
    is_out TINYINT(1) DEFAULT 0,
    is_mentioned TINYINT(1) DEFAULT 0,
    is_media_unread TINYINT(1) DEFAULT 0,
    is_silent TINYINT(1) DEFAULT 0,
    is_post TINYINT(1) DEFAULT 0,
    is_pinned TINYINT(1) DEFAULT 0,
    is_noforwards TINYINT(1) DEFAULT 0,
    views INT DEFAULT NULL,
    forwards INT DEFAULT NULL,
    grouped_id BIGINT DEFAULT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_from_id (from_id),
    INDEX idx_peer (peer_id, peer_type),
    INDEX idx_date (date),
    INDEX idx_random_id (random_id),
    INDEX idx_message_id_peer (message_id, peer_id, peer_type),
    UNIQUE KEY uk_random_id (random_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Dialogs table
CREATE TABLE IF NOT EXISTS dialogs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    peer_id BIGINT NOT NULL,
    peer_type ENUM('user', 'chat', 'channel') NOT NULL DEFAULT 'user',
    top_message INT DEFAULT 0,
    read_inbox_max_id INT DEFAULT 0,
    read_outbox_max_id INT DEFAULT 0,
    unread_count INT DEFAULT 0,
    unread_mentions_count INT DEFAULT 0,
    unread_reactions_count INT DEFAULT 0,
    pts INT DEFAULT 0,
    draft TEXT DEFAULT NULL,
    folder_id INT DEFAULT 0,
    is_pinned TINYINT(1) DEFAULT 0,
    pinned_order INT DEFAULT 0,
    mute_until INT DEFAULT 0,
    notify_sound VARCHAR(50) DEFAULT 'default',
    show_previews TINYINT(1) DEFAULT 1,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_peer (peer_id, peer_type),
    INDEX idx_updated_at (updated_at),
    UNIQUE KEY uk_user_peer (user_id, peer_id, peer_type),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- =====================================================
-- Contacts Tables
-- =====================================================

-- Contacts table
CREATE TABLE IF NOT EXISTS contacts (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    contact_id BIGINT NOT NULL,
    first_name VARCHAR(64) DEFAULT '',
    last_name VARCHAR(64) DEFAULT '',
    is_mutual TINYINT(1) DEFAULT 0,
    is_blocked TINYINT(1) DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_contact_id (contact_id),
    UNIQUE KEY uk_user_contact (user_id, contact_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (contact_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- =====================================================
-- Groups/Channels Tables
-- =====================================================

-- Chats table (groups)
CREATE TABLE IF NOT EXISTS chats (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    title VARCHAR(255) NOT NULL,
    photo_id BIGINT DEFAULT NULL,
    participants_count INT DEFAULT 0,
    date INT NOT NULL,
    version INT DEFAULT 0,
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
    FOREIGN KEY (creator_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

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
    INDEX idx_chat_id (chat_id),
    INDEX idx_user_id (user_id),
    UNIQUE KEY uk_chat_user (chat_id, user_id),
    FOREIGN KEY (chat_id) REFERENCES chats(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Channels table
CREATE TABLE IF NOT EXISTS channels (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    access_hash BIGINT NOT NULL,
    title VARCHAR(255) NOT NULL,
    username VARCHAR(32) UNIQUE,
    photo_id BIGINT DEFAULT NULL,
    date INT NOT NULL,
    version INT DEFAULT 0,
    creator_id BIGINT NOT NULL,
    participants_count INT DEFAULT 0,
    is_broadcast TINYINT(1) DEFAULT 0,
    is_megagroup TINYINT(1) DEFAULT 0,
    is_verified TINYINT(1) DEFAULT 0,
    is_restricted TINYINT(1) DEFAULT 0,
    is_signatures TINYINT(1) DEFAULT 0,
    is_slowmode_enabled TINYINT(1) DEFAULT 0,
    slowmode_seconds INT DEFAULT 0,
    about TEXT,
    admin_rights JSON DEFAULT NULL,
    banned_rights JSON DEFAULT NULL,
    default_banned_rights JSON DEFAULT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_username (username),
    INDEX idx_creator_id (creator_id),
    FOREIGN KEY (creator_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Channel participants table
CREATE TABLE IF NOT EXISTS channel_participants (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    channel_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    inviter_id BIGINT DEFAULT NULL,
    date INT NOT NULL,
    participant_type ENUM('creator', 'admin', 'member', 'banned', 'left') DEFAULT 'member',
    admin_rights JSON DEFAULT NULL,
    banned_rights JSON DEFAULT NULL,
    `rank` VARCHAR(16) DEFAULT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_channel_id (channel_id),
    INDEX idx_user_id (user_id),
    UNIQUE KEY uk_channel_user (channel_id, user_id),
    FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- =====================================================
-- Files Tables
-- =====================================================

-- Photos table
CREATE TABLE IF NOT EXISTS photos (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    access_hash BIGINT NOT NULL,
    file_reference BLOB,
    date INT NOT NULL,
    dc_id INT DEFAULT 1,
    has_stickers TINYINT(1) DEFAULT 0,
    sizes JSON NOT NULL,
    video_sizes JSON DEFAULT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Documents table
CREATE TABLE IF NOT EXISTS documents (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    access_hash BIGINT NOT NULL,
    file_reference BLOB,
    date INT NOT NULL,
    dc_id INT DEFAULT 1,
    mime_type VARCHAR(100) NOT NULL,
    size BIGINT NOT NULL,
    file_name VARCHAR(255) DEFAULT NULL,
    file_path VARCHAR(500) NOT NULL,
    thumb_path VARCHAR(500) DEFAULT NULL,
    attributes JSON DEFAULT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_mime_type (mime_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- File parts table (for upload tracking)
CREATE TABLE IF NOT EXISTS file_parts (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    file_id BIGINT NOT NULL,
    file_part INT NOT NULL,
    file_total_parts INT DEFAULT NULL,
    bytes_size INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_file_id (file_id),
    UNIQUE KEY uk_file_part (file_id, file_part)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- =====================================================
-- Updates Tables
-- =====================================================

-- Updates state table
CREATE TABLE IF NOT EXISTS updates_state (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL UNIQUE,
    pts INT DEFAULT 0,
    qts INT DEFAULT 0,
    seq INT DEFAULT 0,
    date INT DEFAULT 0,
    unread_count INT DEFAULT 0,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Pending updates table
CREATE TABLE IF NOT EXISTS pending_updates (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    update_type VARCHAR(50) NOT NULL,
    update_data JSON NOT NULL,
    pts INT DEFAULT 0,
    pts_count INT DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_pts (pts),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- =====================================================
-- Push Notifications Tables
-- =====================================================

-- Push tokens table
CREATE TABLE IF NOT EXISTS push_tokens (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    auth_key_id VARCHAR(16) NOT NULL,
    token_type ENUM('fcm', 'apns', 'web') NOT NULL,
    token VARCHAR(500) NOT NULL,
    app_sandbox TINYINT(1) DEFAULT 0,
    secret BLOB DEFAULT NULL,
    other_uids JSON DEFAULT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_token (token(255)),
    UNIQUE KEY uk_auth_key_token (auth_key_id, token_type),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- =====================================================
-- Calls Tables
-- =====================================================

-- Calls table
CREATE TABLE IF NOT EXISTS calls (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    call_id VARCHAR(32) NOT NULL UNIQUE,
    caller_id BIGINT NOT NULL,
    callee_id BIGINT NOT NULL,
    call_type ENUM('audio', 'video') DEFAULT 'audio',
    status ENUM('pending', 'ringing', 'accepted', 'active', 'ended', 'missed', 'declined', 'busy') DEFAULT 'pending',
    g_a_hash BLOB DEFAULT NULL,
    g_b BLOB DEFAULT NULL,
    protocol JSON DEFAULT NULL,
    connection JSON DEFAULT NULL,
    e2ee_enabled TINYINT(1) DEFAULT 1,
    e2ee_version VARCHAR(10) DEFAULT '1.0',
    is_using_turn TINYINT(1) DEFAULT 0,
    bitrate_kbps INT DEFAULT NULL,
    duration_seconds INT DEFAULT NULL,
    end_reason VARCHAR(50) DEFAULT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    accepted_at TIMESTAMP NULL,
    ended_at TIMESTAMP NULL,
    INDEX idx_caller_id (caller_id),
    INDEX idx_callee_id (callee_id),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at),
    FOREIGN KEY (caller_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (callee_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- =====================================================
-- Admin Dashboard Tables
-- =====================================================

-- Admins table
CREATE TABLE IF NOT EXISTS admins (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role ENUM('super_admin', 'security_auditor', 'support_agent') DEFAULT 'support_agent',
    permissions JSON DEFAULT NULL,
    is_active TINYINT(1) DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP NULL,
    last_login_ip VARCHAR(45) DEFAULT NULL,
    INDEX idx_email (email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Audit logs table
CREATE TABLE IF NOT EXISTS audit_logs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    operator VARCHAR(255) NOT NULL,
    action_type ENUM('login', 'logout', 'ban', 'unban', 'kick', 'restrict', 'unrestrict', 'reset_session', 'update_config', 'create_user', 'delete_user', 'create_bot', 'delete_bot', 'broadcast') NOT NULL,
    target_user_id BIGINT DEFAULT NULL,
    target_user_phone VARCHAR(20) DEFAULT NULL,
    details JSON DEFAULT NULL,
    ip VARCHAR(45) DEFAULT NULL,
    hash VARCHAR(64) NOT NULL,
    prev_hash VARCHAR(64) DEFAULT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_operator (operator),
    INDEX idx_action_type (action_type),
    INDEX idx_target_user_id (target_user_id),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Bots table
CREATE TABLE IF NOT EXISTS bots (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL UNIQUE,
    bot_token VARCHAR(100) NOT NULL UNIQUE,
    is_official TINYINT(1) DEFAULT 0,
    is_verified TINYINT(1) DEFAULT 0,
    can_broadcast TINYINT(1) DEFAULT 0,
    broadcast_scope ENUM('all', 'premium_users', 'specific_users') DEFAULT 'all',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255) DEFAULT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Broadcasts table
CREATE TABLE IF NOT EXISTS broadcasts (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    bot_id BIGINT NOT NULL,
    message TEXT NOT NULL,
    status ENUM('pending', 'in_progress', 'completed', 'cancelled', 'paused') DEFAULT 'pending',
    target_scope ENUM('all', 'premium_users', 'specific_users') DEFAULT 'all',
    target_user_ids JSON DEFAULT NULL,
    total_users INT DEFAULT 0,
    sent_count INT DEFAULT 0,
    success_count INT DEFAULT 0,
    failed_count INT DEFAULT 0,
    schedule_time TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255) DEFAULT NULL,
    started_at TIMESTAMP NULL,
    completed_at TIMESTAMP NULL,
    INDEX idx_bot_id (bot_id),
    INDEX idx_status (status),
    FOREIGN KEY (bot_id) REFERENCES bots(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- API credentials table
CREATE TABLE IF NOT EXISTS api_credentials (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    api_id INT NOT NULL UNIQUE,
    api_hash VARCHAR(32) NOT NULL,
    platform ENUM('Android', 'iOS', 'Desktop', 'Web') DEFAULT 'Android',
    is_active TINYINT(1) DEFAULT 1,
    usage_count INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255) DEFAULT NULL,
    last_used_at TIMESTAMP NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Universal codes table (for testing)
CREATE TABLE IF NOT EXISTS universal_codes (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    code VARCHAR(10) NOT NULL,
    usage_count INT DEFAULT 0,
    is_active TINYINT(1) DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255) DEFAULT NULL,
    INDEX idx_code (code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Alerts table
CREATE TABLE IF NOT EXISTS alerts (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    level ENUM('HIGH', 'MEDIUM', 'LOW') DEFAULT 'MEDIUM',
    alert_type VARCHAR(50) NOT NULL,
    user_id BIGINT DEFAULT NULL,
    message TEXT NOT NULL,
    is_resolved TINYINT(1) DEFAULT 0,
    resolved_at TIMESTAMP NULL,
    resolved_by VARCHAR(255) DEFAULT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_level (level),
    INDEX idx_alert_type (alert_type),
    INDEX idx_is_resolved (is_resolved),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- System config table
CREATE TABLE IF NOT EXISTS system_config (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    config_key VARCHAR(100) NOT NULL UNIQUE,
    config_value TEXT NOT NULL,
    value_type ENUM('boolean', 'number', 'string', 'json') DEFAULT 'string',
    description TEXT DEFAULT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    updated_by VARCHAR(255) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- User restrictions table
CREATE TABLE IF NOT EXISTS user_restrictions (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    restrictions JSON NOT NULL,
    reason VARCHAR(255) DEFAULT NULL,
    duration_hours INT DEFAULT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255) DEFAULT NULL,
    expires_at TIMESTAMP NULL,
    INDEX idx_user_id (user_id),
    INDEX idx_expires_at (expires_at),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- =====================================================
-- Initial Data
-- =====================================================

-- Insert default admin
INSERT INTO admins (email, password_hash, role, permissions) VALUES 
('admin@feiji.im', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'super_admin', '["view_users", "manage_users", "view_stats", "manage_bots", "manage_broadcasts", "manage_config"]');
-- Default password: admin123

-- Insert default API credentials
INSERT INTO api_credentials (api_id, api_hash, platform, created_by) VALUES 
(2040000, 'A3406DE8D171bb422bb6DDF60E3E70A5', 'Android', 'system');

-- Insert default universal code for testing
INSERT INTO universal_codes (code, created_by) VALUES 
('123456', 'system');

-- Insert default system config
INSERT INTO system_config (config_key, config_value, value_type, description) VALUES 
('push_enabled', 'true', 'boolean', 'Enable push notifications'),
('broadcast_rate_limit', '100', 'number', 'Broadcast messages per second'),
('max_file_size_mb', '2000', 'number', 'Maximum file size in MB'),
('heartbeat_interval_seconds', '60', 'number', 'WebSocket heartbeat interval'),
('call_enabled', 'true', 'boolean', 'Enable audio/video calls');

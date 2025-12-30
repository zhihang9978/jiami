-- Phase 5: Secret Chats and Broadcasts tables
-- Based on BACKEND_COMPLETE_SPECIFICATION.md sections 3.11.5 (broadcasts) and E2EE requirements

-- Secret Chats table (E2E encrypted chats)
CREATE TABLE IF NOT EXISTS secret_chats (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    initiator_id BIGINT NOT NULL,
    peer_id BIGINT NOT NULL,
    status VARCHAR(20) DEFAULT 'WAITING',  -- WAITING, ACTIVE, CLOSED
    g_a TEXT,  -- DH public key A (base64)
    g_b TEXT,  -- DH public key B (base64)
    key_hash VARCHAR(64),  -- Key fingerprint
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_initiator_id (initiator_id),
    INDEX idx_peer_id (peer_id),
    INDEX idx_status (status)
);

-- Secret Messages table (encrypted messages)
CREATE TABLE IF NOT EXISTS secret_messages (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    secret_chat_id BIGINT NOT NULL,
    from_id BIGINT NOT NULL,
    encrypted_message TEXT NOT NULL,  -- base64 encoded encrypted data
    date BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (secret_chat_id) REFERENCES secret_chats(id) ON DELETE CASCADE,
    INDEX idx_secret_chat_id (secret_chat_id),
    INDEX idx_from_id (from_id),
    INDEX idx_date (date)
);

-- Broadcasts table (system broadcasts as per section 3.11.5)
CREATE TABLE IF NOT EXISTS broadcasts (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    creator_id BIGINT NOT NULL,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    status VARCHAR(20) DEFAULT 'DRAFT',  -- DRAFT, SENT, SCHEDULED, IN_PROGRESS, COMPLETED, CANCELLED, PAUSED
    sent_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_creator_id (creator_id),
    INDEX idx_status (status)
);

-- Admin users table (section 3.11.13)
CREATE TABLE IF NOT EXISTS admins (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) DEFAULT 'support_agent',  -- super_admin, security_auditor, support_agent
    permissions JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP NULL,
    last_login_ip VARCHAR(45),
    INDEX idx_email (email),
    INDEX idx_role (role)
);

-- Audit logs table (section 3.11.8)
CREATE TABLE IF NOT EXISTS audit_logs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    operator VARCHAR(255) NOT NULL,
    action_type VARCHAR(50) NOT NULL,  -- ban, unban, kick, restrict, reset_session, update_config, login
    target_user_id BIGINT,
    target_user_phone VARCHAR(20),
    details JSON,
    ip VARCHAR(45),
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    hash VARCHAR(64),  -- SHA256 hash for tamper protection
    prev_hash VARCHAR(64),
    INDEX idx_operator (operator),
    INDEX idx_action_type (action_type),
    INDEX idx_target_user_id (target_user_id),
    INDEX idx_timestamp (timestamp)
);

-- Bots table (section 3.11.4)
CREATE TABLE IF NOT EXISTS bots (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(255) NOT NULL UNIQUE,
    first_name VARCHAR(255) NOT NULL,
    bot_token VARCHAR(255),
    is_official BOOLEAN DEFAULT FALSE,
    is_verified BOOLEAN DEFAULT FALSE,
    can_broadcast BOOLEAN DEFAULT FALSE,
    broadcast_scope VARCHAR(50) DEFAULT 'all',  -- all, premium_users, specific_users
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255),
    INDEX idx_username (username),
    INDEX idx_is_official (is_official)
);

-- API credentials table (section 3.11.6)
CREATE TABLE IF NOT EXISTS api_credentials (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    api_id INT NOT NULL UNIQUE,
    api_hash VARCHAR(64) NOT NULL,
    platform VARCHAR(50) NOT NULL,  -- Android, iOS, Desktop, Web
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255),
    last_used_at TIMESTAMP NULL,
    usage_count INT DEFAULT 0,
    INDEX idx_api_id (api_id),
    INDEX idx_platform (platform)
);

-- Universal codes table (section 3.11.7)
CREATE TABLE IF NOT EXISTS universal_codes (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    code VARCHAR(10) NOT NULL UNIQUE,
    usage_count INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255),
    INDEX idx_code (code)
);

-- Security alerts table (section 3.11.11)
CREATE TABLE IF NOT EXISTS security_alerts (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    level VARCHAR(10) NOT NULL,  -- HIGH, MEDIUM, LOW
    type VARCHAR(50) NOT NULL,  -- high_frequency_message, login_failure, etc.
    user_id BIGINT,
    message TEXT NOT NULL,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_resolved BOOLEAN DEFAULT FALSE,
    INDEX idx_level (level),
    INDEX idx_type (type),
    INDEX idx_user_id (user_id),
    INDEX idx_is_resolved (is_resolved)
);

-- System config table (section 3.11.12)
CREATE TABLE IF NOT EXISTS system_config (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    config_key VARCHAR(100) NOT NULL UNIQUE,
    config_value TEXT NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_config_key (config_key)
);

-- Insert default system config values
INSERT INTO system_config (config_key, config_value) VALUES
    ('push_enabled', 'true'),
    ('broadcast_rate_limit', '100'),
    ('max_file_size_mb', '2000'),
    ('heartbeat_interval_seconds', '60'),
    ('call_enabled', 'true')
ON DUPLICATE KEY UPDATE config_value = VALUES(config_value);

-- Add E2EE fields to calls table (section 3.12.6)
ALTER TABLE calls ADD COLUMN IF NOT EXISTS e2ee_enabled BOOLEAN DEFAULT TRUE;
ALTER TABLE calls ADD COLUMN IF NOT EXISTS e2ee_version VARCHAR(10) DEFAULT '1.0';

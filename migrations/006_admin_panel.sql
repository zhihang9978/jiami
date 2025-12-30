-- Admin Panel Database Migration
-- Based on FeijiIM Complete Admin Panel Development Requirements v1.0

-- 1. Admins table
CREATE TABLE IF NOT EXISTS admins (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  username VARCHAR(50) UNIQUE NOT NULL COMMENT 'Username',
  password_hash VARCHAR(255) NOT NULL COMMENT 'Password hash (bcrypt)',
  role VARCHAR(20) DEFAULT 'super_admin' COMMENT 'Role: super_admin',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  last_login_at TIMESTAMP NULL COMMENT 'Last login time',
  last_login_ip VARCHAR(50) NULL COMMENT 'Last login IP'
);

-- Insert default super admin (password should be changed after first login)
-- Default password hash is for 'changeme' - MUST be changed in production
INSERT INTO admins (username, password_hash, role) VALUES
('admin', '$2a$10$placeholder.hash.change.in.production.setup', 'super_admin')
ON DUPLICATE KEY UPDATE username = username;

-- 2. Admin notifications table
CREATE TABLE IF NOT EXISTS admin_notifications (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  admin_id BIGINT NULL COMMENT 'Admin ID (NULL means all admins)',
  category VARCHAR(50) NOT NULL COMMENT 'Notification category',
  title VARCHAR(200) NOT NULL COMMENT 'Title',
  message TEXT NOT NULL COMMENT 'Message content',
  data JSON NULL COMMENT 'Additional data',
  priority VARCHAR(20) DEFAULT 'normal' COMMENT 'Priority: low/normal/high/urgent',
  is_read BOOLEAN DEFAULT FALSE COMMENT 'Is read',
  link VARCHAR(500) NULL COMMENT 'Jump link',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_admin_id (admin_id),
  INDEX idx_category (category),
  INDEX idx_is_read (is_read),
  INDEX idx_created_at (created_at)
);

-- 3. Audit logs table
CREATE TABLE IF NOT EXISTS audit_logs (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  admin_id BIGINT NOT NULL COMMENT 'Operator ID',
  module VARCHAR(50) NOT NULL COMMENT 'Module',
  action VARCHAR(50) NOT NULL COMMENT 'Action type',
  target_type VARCHAR(50) NULL COMMENT 'Target type',
  target_id BIGINT NULL COMMENT 'Target ID',
  ip_address VARCHAR(50) NULL COMMENT 'IP address',
  user_agent TEXT NULL COMMENT 'User Agent',
  details JSON NULL COMMENT 'Details',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_admin_id (admin_id),
  INDEX idx_module (module),
  INDEX idx_action (action),
  INDEX idx_created_at (created_at)
);

-- 4. Login logs table
CREATE TABLE IF NOT EXISTS login_logs (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  admin_id BIGINT NOT NULL COMMENT 'Admin ID',
  ip_address VARCHAR(50) NOT NULL COMMENT 'IP address',
  user_agent TEXT NULL COMMENT 'User Agent',
  device VARCHAR(100) NULL COMMENT 'Device',
  browser VARCHAR(100) NULL COMMENT 'Browser',
  status VARCHAR(20) NOT NULL COMMENT 'Status: success/failed',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_admin_id (admin_id),
  INDEX idx_status (status),
  INDEX idx_created_at (created_at)
);

-- 5. Extend users table for admin features
ALTER TABLE users ADD COLUMN IF NOT EXISTS custom_code VARCHAR(10) DEFAULT NULL COMMENT 'Custom verification code';
ALTER TABLE users ADD COLUMN IF NOT EXISTS code_expires_at TIMESTAMP NULL COMMENT 'Code expiration time';
ALTER TABLE users ADD COLUMN IF NOT EXISTS allow_call BOOLEAN DEFAULT TRUE COMMENT 'Allow call';
ALTER TABLE users ADD COLUMN IF NOT EXISTS allow_video_call BOOLEAN DEFAULT TRUE COMMENT 'Allow video call';
ALTER TABLE users ADD COLUMN IF NOT EXISTS remark TEXT NULL COMMENT 'Remark';
ALTER TABLE users ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP NULL COMMENT 'Deleted at (soft delete)';
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_bot BOOLEAN DEFAULT FALSE COMMENT 'Is bot';

-- 6. Create official Bot user
INSERT INTO users (id, phone, username, first_name, status, is_bot) VALUES
(1, '10000000000', 'feiji_official', 'FeijiIM Official', 'active', TRUE)
ON DUPLICATE KEY UPDATE is_bot = TRUE;

-- 7. Broadcast messages table
CREATE TABLE IF NOT EXISTS broadcast_messages (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  admin_id BIGINT NOT NULL COMMENT 'Creator ID',
  message_type VARCHAR(20) NOT NULL COMMENT 'Message type: text/image/video/file',
  message_content TEXT NOT NULL COMMENT 'Message content',
  media_url VARCHAR(500) NULL COMMENT 'Media file URL',
  target_type VARCHAR(20) NOT NULL COMMENT 'Target type: all/online/custom',
  target_user_ids TEXT NULL COMMENT 'Target user IDs (JSON)',
  target_filters JSON NULL COMMENT 'Filter conditions',
  total_users INT DEFAULT 0 COMMENT 'Total target users',
  success_count INT DEFAULT 0 COMMENT 'Success count',
  failed_count INT DEFAULT 0 COMMENT 'Failed count',
  status VARCHAR(20) DEFAULT 'draft' COMMENT 'Status: draft/pending/sending/completed/partial_failed/cancelled',
  scheduled_at TIMESTAMP NULL COMMENT 'Scheduled send time',
  sent_at TIMESTAMP NULL COMMENT 'Actual send time',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_admin_id (admin_id),
  INDEX idx_status (status),
  INDEX idx_created_at (created_at)
);

-- 8. Broadcast details table
CREATE TABLE IF NOT EXISTS broadcast_details (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  broadcast_id BIGINT NOT NULL COMMENT 'Broadcast ID',
  user_id BIGINT NOT NULL COMMENT 'User ID',
  status VARCHAR(20) NOT NULL COMMENT 'Status: success/failed',
  error_message TEXT NULL COMMENT 'Error message',
  sent_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_broadcast_id (broadcast_id),
  INDEX idx_user_id (user_id),
  INDEX idx_status (status),
  FOREIGN KEY (broadcast_id) REFERENCES broadcast_messages(id) ON DELETE CASCADE
);

-- 9. Message templates table
CREATE TABLE IF NOT EXISTS message_templates (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  name VARCHAR(100) NOT NULL COMMENT 'Template name',
  message_type VARCHAR(20) NOT NULL COMMENT 'Message type',
  message_content TEXT NOT NULL COMMENT 'Message content',
  media_url VARCHAR(500) NULL COMMENT 'Media file URL',
  usage_count INT DEFAULT 0 COMMENT 'Usage count',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- Insert default templates
INSERT INTO message_templates (name, message_type, message_content) VALUES
('Welcome New User', 'text', 'Welcome to FeijiIM! We are a secure and fast instant messaging application...'),
('System Maintenance', 'text', 'System will be under maintenance on {date} {time}, expected duration: {duration}...'),
('New Feature Release', 'text', 'FeijiIM new version released! New features: ...'),
('Holiday Greetings', 'text', 'FeijiIM team wishes you a happy {holiday}!')
ON DUPLICATE KEY UPDATE name = name;

-- 10. Auto message configs table
CREATE TABLE IF NOT EXISTS auto_message_configs (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  type VARCHAR(50) NOT NULL UNIQUE COMMENT 'Type: welcome/first_login/inactive_reminder',
  message_content TEXT NOT NULL COMMENT 'Message content',
  is_enabled BOOLEAN DEFAULT TRUE COMMENT 'Is enabled',
  trigger_condition JSON NULL COMMENT 'Trigger condition',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- Insert default auto message configs
INSERT INTO auto_message_configs (type, message_content, is_enabled) VALUES
('welcome', 'Welcome to FeijiIM!', TRUE),
('first_login', 'Thank you for your first login to FeijiIM!', TRUE),
('inactive_reminder', 'Long time no see! We miss you.', FALSE)
ON DUPLICATE KEY UPDATE type = type;

-- 11. System configs table
CREATE TABLE IF NOT EXISTS system_configs (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  config_key VARCHAR(100) NOT NULL UNIQUE COMMENT 'Config key',
  config_value TEXT NOT NULL COMMENT 'Config value',
  config_type VARCHAR(20) DEFAULT 'string' COMMENT 'Config type: string/number/boolean/json',
  description VARCHAR(500) NULL COMMENT 'Description',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- Insert default system configs
INSERT INTO system_configs (config_key, config_value, config_type, description) VALUES
('system_name', 'FeijiIM', 'string', 'System name'),
('system_version', '1.0.0', 'string', 'System version'),
('api_url', 'https://im.zhihang.icu', 'string', 'API URL'),
('ws_url', 'wss://api.zhihang.icu/ws', 'string', 'WebSocket URL'),
('turn_server', 'turn:YOUR_SERVER_IP:3478', 'string', 'TURN server'),
('stun_server', 'stun:YOUR_SERVER_IP:3478', 'string', 'STUN server'),
('turn_username', 'YOUR_TURN_USERNAME', 'string', 'TURN username'),
('turn_password', 'YOUR_TURN_PASSWORD', 'string', 'TURN password'),
('turn_realm', 'YOUR_DOMAIN', 'string', 'TURN realm'),
('call_answer_timeout', '20', 'number', 'Call answer timeout (seconds)'),
('call_ring_timeout', '90', 'number', 'Call ring timeout (seconds)'),
('max_call_duration', '0', 'number', 'Max call duration (0 = unlimited)'),
('allow_video_call', 'true', 'boolean', 'Allow video call'),
('e2ee_enabled', 'true', 'boolean', 'E2EE enabled'),
('e2ee_version', '1.0', 'string', 'E2EE version'),
('ws_heartbeat_interval', '30', 'number', 'WebSocket heartbeat interval (seconds)'),
('ws_reconnect_interval', '5', 'number', 'WebSocket reconnect interval (seconds)'),
('ws_max_reconnect', '10', 'number', 'WebSocket max reconnect attempts')
ON DUPLICATE KEY UPDATE config_key = config_key;

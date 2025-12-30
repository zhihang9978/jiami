-- Phase 1 Additional Tables
-- Version: 1.1

-- Files table (for completed uploads)
CREATE TABLE IF NOT EXISTS files (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    file_id BIGINT NOT NULL UNIQUE,
    access_hash BIGINT NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    mime_type VARCHAR(100) DEFAULT 'application/octet-stream',
    size BIGINT NOT NULL,
    file_path VARCHAR(500) NOT NULL,
    thumbnail_path VARCHAR(500) DEFAULT NULL,
    width INT DEFAULT 0,
    height INT DEFAULT 0,
    duration INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_file_id (file_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- File ID sequence table
CREATE TABLE IF NOT EXISTS file_id_seq (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    stub CHAR(1) NOT NULL DEFAULT 'a',
    UNIQUE KEY uk_stub (stub)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Update file_parts to include data column (ignore error if column already exists)
-- ALTER TABLE file_parts ADD COLUMN data LONGBLOB DEFAULT NULL;
-- ALTER TABLE file_parts MODIFY COLUMN part_num INT NOT NULL;

-- Update state table (rename if needed)
CREATE TABLE IF NOT EXISTS update_state (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL UNIQUE,
    pts INT DEFAULT 1,
    qts INT DEFAULT 0,
    seq INT DEFAULT 1,
    date BIGINT DEFAULT 0,
    unread_count INT DEFAULT 0,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Updates table (for storing updates history)
CREATE TABLE IF NOT EXISTS updates (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    update_type VARCHAR(50) NOT NULL,
    pts INT DEFAULT 0,
    qts INT DEFAULT 0,
    seq INT DEFAULT 0,
    data JSON DEFAULT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_pts (user_id, pts),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Initialize file_id_seq
INSERT IGNORE INTO file_id_seq (stub) VALUES ('a');

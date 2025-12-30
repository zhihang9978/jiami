-- Phase 3-4 Additions: Calls table for VoIP functionality

-- Calls table for VoIP
CREATE TABLE IF NOT EXISTS calls (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    access_hash BIGINT NOT NULL,
    caller_id BIGINT NOT NULL,
    callee_id BIGINT NOT NULL,
    date INT NOT NULL,
    duration INT DEFAULT 0,
    is_video BOOLEAN DEFAULT FALSE,
    is_outgoing BOOLEAN DEFAULT TRUE,
    state VARCHAR(20) DEFAULT 'pending',
    protocol JSON,
    connections JSON,
    encryption_key BLOB,
    key_fingerprint BIGINT DEFAULT 0,
    ga_hash BLOB,
    gb BLOB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_caller_id (caller_id),
    INDEX idx_callee_id (callee_id),
    INDEX idx_state (state),
    INDEX idx_date (date)
);

-- Call ratings table (optional, for quality feedback)
CREATE TABLE IF NOT EXISTS call_ratings (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    call_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    rating INT NOT NULL,
    comment TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (call_id) REFERENCES calls(id) ON DELETE CASCADE,
    INDEX idx_call_id (call_id),
    INDEX idx_user_id (user_id)
);

-- Call debug logs table (optional, for debugging)
CREATE TABLE IF NOT EXISTS call_debug_logs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    call_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    debug_data JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (call_id) REFERENCES calls(id) ON DELETE CASCADE,
    INDEX idx_call_id (call_id)
);

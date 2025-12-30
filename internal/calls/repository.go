package calls

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"
)

type Call struct {
	ID              int64
	AccessHash      int64
	CallerID        int64
	CalleeID        int64
	Date            int
	Duration        int
	IsVideo         bool
	IsOutgoing      bool
	State           string // pending, ringing, accepted, ended, missed, declined, busy
	Protocol        json.RawMessage
	Connections     json.RawMessage
	EncryptionKey   []byte
	KeyFingerprint  int64
	GAHash          []byte
	GB              []byte
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// CreateCall creates a new call
func (r *Repository) CreateCall(ctx context.Context, call *Call) error {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO calls (access_hash, caller_id, callee_id, date, duration, is_video, is_outgoing, state,
		                   protocol, connections, encryption_key, key_fingerprint, ga_hash, gb, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`, call.AccessHash, call.CallerID, call.CalleeID, call.Date, call.Duration, call.IsVideo, call.IsOutgoing,
		call.State, call.Protocol, call.Connections, call.EncryptionKey, call.KeyFingerprint, call.GAHash, call.GB)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	call.ID = id
	return nil
}

// GetCall retrieves a call by ID
func (r *Repository) GetCall(ctx context.Context, callID int64) (*Call, error) {
	call := &Call{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, access_hash, caller_id, callee_id, date, duration, is_video, is_outgoing, state,
		       protocol, connections, encryption_key, key_fingerprint, ga_hash, gb, created_at, updated_at
		FROM calls WHERE id = ?
	`, callID).Scan(&call.ID, &call.AccessHash, &call.CallerID, &call.CalleeID, &call.Date, &call.Duration,
		&call.IsVideo, &call.IsOutgoing, &call.State, &call.Protocol, &call.Connections,
		&call.EncryptionKey, &call.KeyFingerprint, &call.GAHash, &call.GB, &call.CreatedAt, &call.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return call, nil
}

// UpdateCall updates a call
func (r *Repository) UpdateCall(ctx context.Context, call *Call) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE calls SET state = ?, duration = ?, connections = ?, encryption_key = ?, 
		                 key_fingerprint = ?, ga_hash = ?, gb = ?, updated_at = NOW()
		WHERE id = ?
	`, call.State, call.Duration, call.Connections, call.EncryptionKey,
		call.KeyFingerprint, call.GAHash, call.GB, call.ID)
	return err
}

// GetUserCalls retrieves calls for a user
func (r *Repository) GetUserCalls(ctx context.Context, userID int64, offset, limit int) ([]*Call, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, access_hash, caller_id, callee_id, date, duration, is_video, is_outgoing, state,
		       protocol, connections, encryption_key, key_fingerprint, ga_hash, gb, created_at, updated_at
		FROM calls
		WHERE caller_id = ? OR callee_id = ?
		ORDER BY date DESC
		LIMIT ? OFFSET ?
	`, userID, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var calls []*Call
	for rows.Next() {
		call := &Call{}
		if err := rows.Scan(&call.ID, &call.AccessHash, &call.CallerID, &call.CalleeID, &call.Date, &call.Duration,
			&call.IsVideo, &call.IsOutgoing, &call.State, &call.Protocol, &call.Connections,
			&call.EncryptionKey, &call.KeyFingerprint, &call.GAHash, &call.GB, &call.CreatedAt, &call.UpdatedAt); err != nil {
			return nil, err
		}
		calls = append(calls, call)
	}
	return calls, nil
}

// GetActiveCall retrieves an active call between two users
func (r *Repository) GetActiveCall(ctx context.Context, userID1, userID2 int64) (*Call, error) {
	call := &Call{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, access_hash, caller_id, callee_id, date, duration, is_video, is_outgoing, state,
		       protocol, connections, encryption_key, key_fingerprint, ga_hash, gb, created_at, updated_at
		FROM calls
		WHERE ((caller_id = ? AND callee_id = ?) OR (caller_id = ? AND callee_id = ?))
		AND state IN ('pending', 'ringing', 'accepted')
		ORDER BY date DESC LIMIT 1
	`, userID1, userID2, userID2, userID1).Scan(&call.ID, &call.AccessHash, &call.CallerID, &call.CalleeID,
		&call.Date, &call.Duration, &call.IsVideo, &call.IsOutgoing, &call.State, &call.Protocol, &call.Connections,
		&call.EncryptionKey, &call.KeyFingerprint, &call.GAHash, &call.GB, &call.CreatedAt, &call.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return call, nil
}

// GetPendingCallsForUser retrieves pending incoming calls for a user
func (r *Repository) GetPendingCallsForUser(ctx context.Context, userID int64) ([]*Call, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, access_hash, caller_id, callee_id, date, duration, is_video, is_outgoing, state,
		       protocol, connections, encryption_key, key_fingerprint, ga_hash, gb, created_at, updated_at
		FROM calls
		WHERE callee_id = ? AND state IN ('pending', 'ringing')
		ORDER BY date DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var calls []*Call
	for rows.Next() {
		call := &Call{}
		if err := rows.Scan(&call.ID, &call.AccessHash, &call.CallerID, &call.CalleeID, &call.Date, &call.Duration,
			&call.IsVideo, &call.IsOutgoing, &call.State, &call.Protocol, &call.Connections,
			&call.EncryptionKey, &call.KeyFingerprint, &call.GAHash, &call.GB, &call.CreatedAt, &call.UpdatedAt); err != nil {
			return nil, err
		}
		calls = append(calls, call)
	}
	return calls, nil
}

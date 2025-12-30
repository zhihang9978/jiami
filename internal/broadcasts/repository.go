package broadcasts

import (
	"context"
	"database/sql"
	"time"
)

type Broadcast struct {
	ID        int64
	CreatorID int64
	Title     string
	Content   string
	Status    string // DRAFT, SENT, SCHEDULED
	SentAt    *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// CreateBroadcast creates a new broadcast
func (r *Repository) CreateBroadcast(ctx context.Context, broadcast *Broadcast) error {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO broadcasts (creator_id, title, content, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, NOW(), NOW())
	`, broadcast.CreatorID, broadcast.Title, broadcast.Content, broadcast.Status)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	broadcast.ID = id
	return nil
}

// GetBroadcast retrieves a broadcast by ID
func (r *Repository) GetBroadcast(ctx context.Context, broadcastID int64) (*Broadcast, error) {
	broadcast := &Broadcast{}
	var sentAt sql.NullTime
	err := r.db.QueryRowContext(ctx, `
		SELECT id, creator_id, title, content, status, sent_at, created_at, updated_at
		FROM broadcasts WHERE id = ?
	`, broadcastID).Scan(&broadcast.ID, &broadcast.CreatorID, &broadcast.Title, &broadcast.Content,
		&broadcast.Status, &sentAt, &broadcast.CreatedAt, &broadcast.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if sentAt.Valid {
		broadcast.SentAt = &sentAt.Time
	}
	return broadcast, nil
}

// UpdateBroadcast updates a broadcast
func (r *Repository) UpdateBroadcast(ctx context.Context, broadcast *Broadcast) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE broadcasts SET title = ?, content = ?, status = ?, sent_at = ?, updated_at = NOW()
		WHERE id = ?
	`, broadcast.Title, broadcast.Content, broadcast.Status, broadcast.SentAt, broadcast.ID)
	return err
}

// DeleteBroadcast deletes a broadcast
func (r *Repository) DeleteBroadcast(ctx context.Context, broadcastID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM broadcasts WHERE id = ?`, broadcastID)
	return err
}

// GetBroadcasts retrieves all broadcasts with pagination
func (r *Repository) GetBroadcasts(ctx context.Context, limit, offset int) ([]*Broadcast, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, creator_id, title, content, status, sent_at, created_at, updated_at
		FROM broadcasts
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var broadcasts []*Broadcast
	for rows.Next() {
		broadcast := &Broadcast{}
		var sentAt sql.NullTime
		if err := rows.Scan(&broadcast.ID, &broadcast.CreatorID, &broadcast.Title, &broadcast.Content,
			&broadcast.Status, &sentAt, &broadcast.CreatedAt, &broadcast.UpdatedAt); err != nil {
			return nil, err
		}
		if sentAt.Valid {
			broadcast.SentAt = &sentAt.Time
		}
		broadcasts = append(broadcasts, broadcast)
	}
	return broadcasts, nil
}

// GetUserBroadcasts retrieves broadcasts created by a user
func (r *Repository) GetUserBroadcasts(ctx context.Context, userID int64, limit, offset int) ([]*Broadcast, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, creator_id, title, content, status, sent_at, created_at, updated_at
		FROM broadcasts
		WHERE creator_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var broadcasts []*Broadcast
	for rows.Next() {
		broadcast := &Broadcast{}
		var sentAt sql.NullTime
		if err := rows.Scan(&broadcast.ID, &broadcast.CreatorID, &broadcast.Title, &broadcast.Content,
			&broadcast.Status, &sentAt, &broadcast.CreatedAt, &broadcast.UpdatedAt); err != nil {
			return nil, err
		}
		if sentAt.Valid {
			broadcast.SentAt = &sentAt.Time
		}
		broadcasts = append(broadcasts, broadcast)
	}
	return broadcasts, nil
}

// MarkAsSent marks a broadcast as sent
func (r *Repository) MarkAsSent(ctx context.Context, broadcastID int64) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE broadcasts SET status = 'SENT', sent_at = NOW(), updated_at = NOW()
		WHERE id = ?
	`, broadcastID)
	return err
}

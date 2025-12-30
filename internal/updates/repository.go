package updates

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"
)

type Update struct {
	ID        int64
	UserID    int64
	UpdateType string
	Pts       int
	Qts       int
	Seq       int
	Data      json.RawMessage
	CreatedAt time.Time
}

type State struct {
	UserID    int64
	Pts       int
	Qts       int
	Seq       int
	Date      int64
	UnreadCount int
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// GetState retrieves the current state for a user
func (r *Repository) GetState(ctx context.Context, userID int64) (*State, error) {
	state := &State{UserID: userID}
	err := r.db.QueryRowContext(ctx, `
		SELECT pts, qts, seq, date, unread_count
		FROM update_state
		WHERE user_id = ?
	`, userID).Scan(&state.Pts, &state.Qts, &state.Seq, &state.Date, &state.UnreadCount)
	
	if err == sql.ErrNoRows {
		// Create initial state
		state.Pts = 1
		state.Qts = 0
		state.Seq = 1
		state.Date = time.Now().Unix()
		state.UnreadCount = 0
		
		_, err = r.db.ExecContext(ctx, `
			INSERT INTO update_state (user_id, pts, qts, seq, date, unread_count)
			VALUES (?, ?, ?, ?, ?, ?)
		`, userID, state.Pts, state.Qts, state.Seq, state.Date, state.UnreadCount)
		if err != nil {
			return nil, err
		}
		return state, nil
	}
	if err != nil {
		return nil, err
	}
	return state, nil
}

// UpdateState updates the state for a user
func (r *Repository) UpdateState(ctx context.Context, state *State) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE update_state
		SET pts = ?, qts = ?, seq = ?, date = ?, unread_count = ?
		WHERE user_id = ?
	`, state.Pts, state.Qts, state.Seq, state.Date, state.UnreadCount, state.UserID)
	return err
}

// IncrementPts increments pts and returns the new value
func (r *Repository) IncrementPts(ctx context.Context, userID int64) (int, error) {
	result, err := r.db.ExecContext(ctx, `
		UPDATE update_state
		SET pts = pts + 1, date = ?
		WHERE user_id = ?
	`, time.Now().Unix(), userID)
	if err != nil {
		return 0, err
	}
	
	affected, _ := result.RowsAffected()
	if affected == 0 {
		// Create state if not exists
		_, err := r.GetState(ctx, userID)
		if err != nil {
			return 0, err
		}
		return 1, nil
	}
	
	// Get current pts
	var pts int
	err = r.db.QueryRowContext(ctx, `SELECT pts FROM update_state WHERE user_id = ?`, userID).Scan(&pts)
	return pts, err
}

// SaveUpdate saves an update for a user
func (r *Repository) SaveUpdate(ctx context.Context, update *Update) error {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO updates (user_id, update_type, pts, qts, seq, data, created_at)
		VALUES (?, ?, ?, ?, ?, ?, NOW())
	`, update.UserID, update.UpdateType, update.Pts, update.Qts, update.Seq, update.Data)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	update.ID = id
	return nil
}

// GetDifference retrieves updates since a given pts
func (r *Repository) GetDifference(ctx context.Context, userID int64, pts, qts int, limit int) ([]*Update, error) {
	if limit <= 0 {
		limit = 100
	}
	
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, update_type, pts, qts, seq, data, created_at
		FROM updates
		WHERE user_id = ? AND pts > ?
		ORDER BY pts ASC
		LIMIT ?
	`, userID, pts, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var updates []*Update
	for rows.Next() {
		update := &Update{}
		if err := rows.Scan(&update.ID, &update.UserID, &update.UpdateType, &update.Pts, 
			&update.Qts, &update.Seq, &update.Data, &update.CreatedAt); err != nil {
			return nil, err
		}
		updates = append(updates, update)
	}
	return updates, nil
}

// CleanOldUpdates removes updates older than a given time
func (r *Repository) CleanOldUpdates(ctx context.Context, before time.Time) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM updates WHERE created_at < ?`, before)
	return err
}

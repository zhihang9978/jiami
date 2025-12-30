package search

import (
	"context"
	"database/sql"
	"time"
)

type SearchResult struct {
	Type      string // user, chat, channel, message
	ID        int64
	Title     string
	Username  *string
	PhotoID   *int64
	Date      int
	PeerID    int64
	PeerType  string
	MessageID int64
	Text      string
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// SearchUsers searches for users by name or username
func (r *Repository) SearchUsers(ctx context.Context, query string, limit int) ([]*SearchResult, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, COALESCE(first_name, ''), COALESCE(last_name, ''), username, photo_id
		FROM users
		WHERE (first_name LIKE ? OR last_name LIKE ? OR username LIKE ?)
		AND is_deleted = 0
		ORDER BY 
			CASE WHEN username = ? THEN 0
			     WHEN username LIKE ? THEN 1
			     ELSE 2 END,
			first_name
		LIMIT ?
	`, "%"+query+"%", "%"+query+"%", "%"+query+"%", query, query+"%", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*SearchResult
	for rows.Next() {
		var id int64
		var firstName, lastName string
		var username *string
		var photoID *int64
		if err := rows.Scan(&id, &firstName, &lastName, &username, &photoID); err != nil {
			return nil, err
		}
		title := firstName
		if lastName != "" {
			title += " " + lastName
		}
		results = append(results, &SearchResult{
			Type:     "user",
			ID:       id,
			Title:    title,
			Username: username,
			PhotoID:  photoID,
		})
	}
	return results, nil
}

// SearchChats searches for chats by title
func (r *Repository) SearchChats(ctx context.Context, userID int64, query string, limit int) ([]*SearchResult, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT c.id, c.title, c.photo_id, c.date
		FROM chats c
		JOIN chat_participants cp ON c.id = cp.chat_id
		WHERE cp.user_id = ? AND c.title LIKE ? AND c.is_deactivated = 0
		ORDER BY c.updated_at DESC
		LIMIT ?
	`, userID, "%"+query+"%", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*SearchResult
	for rows.Next() {
		r := &SearchResult{Type: "chat"}
		if err := rows.Scan(&r.ID, &r.Title, &r.PhotoID, &r.Date); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, nil
}

// SearchChannels searches for channels by title or username
func (r *Repository) SearchChannels(ctx context.Context, query string, limit int) ([]*SearchResult, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, title, username, photo_id, date
		FROM channels
		WHERE (title LIKE ? OR username LIKE ?)
		ORDER BY participants_count DESC
		LIMIT ?
	`, "%"+query+"%", "%"+query+"%", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*SearchResult
	for rows.Next() {
		r := &SearchResult{Type: "channel"}
		if err := rows.Scan(&r.ID, &r.Title, &r.Username, &r.PhotoID, &r.Date); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, nil
}

// SearchMessages searches for messages containing query
func (r *Repository) SearchMessages(ctx context.Context, userID int64, query string, peerID int64, peerType string, offsetID int64, limit int) ([]*SearchResult, error) {
	baseQuery := `
		SELECT m.id, m.peer_id, m.peer_type, m.message, m.date
		FROM messages m
		WHERE m.from_id = ? OR m.peer_id = ?`
	args := []interface{}{userID, userID}

	if query != "" {
		baseQuery += ` AND m.message LIKE ?`
		args = append(args, "%"+query+"%")
	}

	if peerID > 0 {
		baseQuery += ` AND m.peer_id = ? AND m.peer_type = ?`
		args = append(args, peerID, peerType)
	}

	if offsetID > 0 {
		baseQuery += ` AND m.id < ?`
		args = append(args, offsetID)
	}

	baseQuery += ` ORDER BY m.date DESC LIMIT ?`
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*SearchResult
	for rows.Next() {
		r := &SearchResult{Type: "message"}
		if err := rows.Scan(&r.MessageID, &r.PeerID, &r.PeerType, &r.Text, &r.Date); err != nil {
			return nil, err
		}
		r.ID = r.MessageID
		results = append(results, r)
	}
	return results, nil
}

// SearchGlobal performs a global search across all types
func (r *Repository) SearchGlobal(ctx context.Context, userID int64, query string, limit int) (*GlobalSearchResult, error) {
	result := &GlobalSearchResult{}

	// Search users
	users, err := r.SearchUsers(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	result.Users = users

	// Search chats
	chats, err := r.SearchChats(ctx, userID, query, limit)
	if err != nil {
		return nil, err
	}
	result.Chats = chats

	// Search channels
	channels, err := r.SearchChannels(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	result.Channels = channels

	// Search messages
	messages, err := r.SearchMessages(ctx, userID, query, 0, "", 0, limit)
	if err != nil {
		return nil, err
	}
	result.Messages = messages

	return result, nil
}

// SearchHashtag searches for messages with a specific hashtag
func (r *Repository) SearchHashtag(ctx context.Context, hashtag string, offsetID int64, limit int) ([]*SearchResult, error) {
	query := `
		SELECT m.id, m.peer_id, m.peer_type, m.message, m.date
		FROM messages m
		WHERE m.message LIKE ?`
	args := []interface{}{"%" + hashtag + "%"}

	if offsetID > 0 {
		query += ` AND m.id < ?`
		args = append(args, offsetID)
	}

	query += ` ORDER BY m.date DESC LIMIT ?`
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*SearchResult
	for rows.Next() {
		r := &SearchResult{Type: "message"}
		if err := rows.Scan(&r.MessageID, &r.PeerID, &r.PeerType, &r.Text, &r.Date); err != nil {
			return nil, err
		}
		r.ID = r.MessageID
		results = append(results, r)
	}
	return results, nil
}

// GetRecentSearch retrieves recent search queries
func (r *Repository) GetRecentSearch(ctx context.Context, userID int64, limit int) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT query FROM search_history
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT ?
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var queries []string
	for rows.Next() {
		var q string
		if err := rows.Scan(&q); err != nil {
			return nil, err
		}
		queries = append(queries, q)
	}
	return queries, nil
}

// SaveSearchQuery saves a search query to history
func (r *Repository) SaveSearchQuery(ctx context.Context, userID int64, query string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO search_history (user_id, query, created_at)
		VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE created_at = VALUES(created_at)
	`, userID, query, time.Now())
	return err
}

// ClearRecentSearch clears search history
func (r *Repository) ClearRecentSearch(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM search_history WHERE user_id = ?`, userID)
	return err
}

// GlobalSearchResult contains all search results
type GlobalSearchResult struct {
	Users    []*SearchResult
	Chats    []*SearchResult
	Channels []*SearchResult
	Messages []*SearchResult
}

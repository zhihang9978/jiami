package search

import (
	"context"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// SearchGlobal performs a global search across all types
func (s *Service) SearchGlobal(ctx context.Context, userID int64, query string, limit int) (*GlobalSearchResult, error) {
	if limit <= 0 {
		limit = 20
	}

	// Save search query to history
	_ = s.repo.SaveSearchQuery(ctx, userID, query)

	return s.repo.SearchGlobal(ctx, userID, query, limit)
}

// SearchUsers searches for users
func (s *Service) SearchUsers(ctx context.Context, query string, limit int) ([]*SearchResult, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.repo.SearchUsers(ctx, query, limit)
}

// SearchChats searches for chats
func (s *Service) SearchChats(ctx context.Context, userID int64, query string, limit int) ([]*SearchResult, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.repo.SearchChats(ctx, userID, query, limit)
}

// SearchChannels searches for channels
func (s *Service) SearchChannels(ctx context.Context, query string, limit int) ([]*SearchResult, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.repo.SearchChannels(ctx, query, limit)
}

// SearchMessages searches for messages
func (s *Service) SearchMessages(ctx context.Context, userID int64, query string, peerID int64, peerType string, offsetID int64, limit int) ([]*SearchResult, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.repo.SearchMessages(ctx, userID, query, peerID, peerType, offsetID, limit)
}

// SearchHashtag searches for messages with a specific hashtag
func (s *Service) SearchHashtag(ctx context.Context, hashtag string, offsetID int64, limit int) ([]*SearchResult, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.repo.SearchHashtag(ctx, hashtag, offsetID, limit)
}

// GetRecentSearch retrieves recent search queries
func (s *Service) GetRecentSearch(ctx context.Context, userID int64) ([]string, error) {
	return s.repo.GetRecentSearch(ctx, userID, 20)
}

// ClearRecentSearch clears search history
func (s *Service) ClearRecentSearch(ctx context.Context, userID int64) error {
	return s.repo.ClearRecentSearch(ctx, userID)
}

// ToTL converts GlobalSearchResult to TL format
func (r *GlobalSearchResult) ToTL() map[string]interface{} {
	users := make([]map[string]interface{}, len(r.Users))
	for i, u := range r.Users {
		users[i] = u.ToUserTL()
	}

	chats := make([]map[string]interface{}, len(r.Chats))
	for i, c := range r.Chats {
		chats[i] = c.ToChatTL()
	}

	channels := make([]map[string]interface{}, len(r.Channels))
	for i, c := range r.Channels {
		channels[i] = c.ToChannelTL()
	}

	messages := make([]map[string]interface{}, len(r.Messages))
	for i, m := range r.Messages {
		messages[i] = m.ToMessageTL()
	}

	return map[string]interface{}{
		"_":        "messages.searchGlobal",
		"users":    users,
		"chats":    chats,
		"channels": channels,
		"messages": messages,
		"count":    len(users) + len(chats) + len(channels) + len(messages),
	}
}

// ToUserTL converts SearchResult to user TL format
func (r *SearchResult) ToUserTL() map[string]interface{} {
	result := map[string]interface{}{
		"_":  "user",
		"id": r.ID,
	}
	if r.Title != "" {
		result["first_name"] = r.Title
	}
	if r.Username != nil {
		result["username"] = *r.Username
	}
	if r.PhotoID != nil {
		result["photo"] = map[string]interface{}{
			"_":        "userProfilePhoto",
			"photo_id": *r.PhotoID,
		}
	} else {
		result["photo"] = map[string]interface{}{"_": "userProfilePhotoEmpty"}
	}
	return result
}

// ToChatTL converts SearchResult to chat TL format
func (r *SearchResult) ToChatTL() map[string]interface{} {
	result := map[string]interface{}{
		"_":     "chat",
		"id":    r.ID,
		"title": r.Title,
		"date":  r.Date,
	}
	if r.PhotoID != nil {
		result["photo"] = map[string]interface{}{
			"_":        "chatPhoto",
			"photo_id": *r.PhotoID,
		}
	} else {
		result["photo"] = map[string]interface{}{"_": "chatPhotoEmpty"}
	}
	return result
}

// ToChannelTL converts SearchResult to channel TL format
func (r *SearchResult) ToChannelTL() map[string]interface{} {
	result := map[string]interface{}{
		"_":     "channel",
		"id":    r.ID,
		"title": r.Title,
		"date":  r.Date,
	}
	if r.Username != nil {
		result["username"] = *r.Username
	}
	if r.PhotoID != nil {
		result["photo"] = map[string]interface{}{
			"_":        "chatPhoto",
			"photo_id": *r.PhotoID,
		}
	} else {
		result["photo"] = map[string]interface{}{"_": "chatPhotoEmpty"}
	}
	return result
}

// ToMessageTL converts SearchResult to message TL format
func (r *SearchResult) ToMessageTL() map[string]interface{} {
	return map[string]interface{}{
		"_":         "message",
		"id":        r.MessageID,
		"peer_id":   r.PeerID,
		"peer_type": r.PeerType,
		"message":   r.Text,
		"date":      r.Date,
	}
}

package updates

import (
	"context"
	"encoding/json"
	"time"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// GetState retrieves the current state for a user
func (s *Service) GetState(ctx context.Context, userID int64) (*State, error) {
	return s.repo.GetState(ctx, userID)
}

// GetDifference retrieves updates since a given pts
func (s *Service) GetDifference(ctx context.Context, userID int64, pts, qts, date int) (*DifferenceResult, error) {
	state, err := s.repo.GetState(ctx, userID)
	if err != nil {
		return nil, err
	}
	
	// If pts is current, no updates
	if pts >= state.Pts {
		return &DifferenceResult{
			Type:  "updates.differenceEmpty",
			Date:  state.Date,
			Seq:   state.Seq,
		}, nil
	}
	
	// Get updates
	updates, err := s.repo.GetDifference(ctx, userID, pts, qts, 100)
	if err != nil {
		return nil, err
	}
	
	if len(updates) == 0 {
		return &DifferenceResult{
			Type:  "updates.differenceEmpty",
			Date:  state.Date,
			Seq:   state.Seq,
		}, nil
	}
	
	// Check if there are more updates
	if len(updates) >= 100 {
		return &DifferenceResult{
			Type:           "updates.differenceSlice",
			NewMessages:    extractMessages(updates),
			NewEncryptedMessages: []interface{}{},
			OtherUpdates:   extractOtherUpdates(updates),
			Chats:          []interface{}{},
			Users:          []interface{}{},
			IntermediateState: &State{
				Pts:  updates[len(updates)-1].Pts,
				Qts:  state.Qts,
				Seq:  state.Seq,
				Date: state.Date,
			},
		}, nil
	}
	
	return &DifferenceResult{
		Type:           "updates.difference",
		NewMessages:    extractMessages(updates),
		NewEncryptedMessages: []interface{}{},
		OtherUpdates:   extractOtherUpdates(updates),
		Chats:          []interface{}{},
		Users:          []interface{}{},
		State:          state,
	}, nil
}

// PushUpdate creates and saves an update for a user
func (s *Service) PushUpdate(ctx context.Context, userID int64, updateType string, data interface{}) (*Update, error) {
	// Increment pts
	pts, err := s.repo.IncrementPts(ctx, userID)
	if err != nil {
		return nil, err
	}
	
	// Serialize data
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	
	update := &Update{
		UserID:     userID,
		UpdateType: updateType,
		Pts:        pts,
		Qts:        0,
		Seq:        0,
		Data:       jsonData,
	}
	
	if err := s.repo.SaveUpdate(ctx, update); err != nil {
		return nil, err
	}
	
	return update, nil
}

// CleanOldUpdates removes updates older than 24 hours
func (s *Service) CleanOldUpdates(ctx context.Context) error {
	return s.repo.CleanOldUpdates(ctx, time.Now().Add(-24*time.Hour))
}

// DifferenceResult represents the result of getDifference
type DifferenceResult struct {
	Type                 string
	NewMessages          []interface{}
	NewEncryptedMessages []interface{}
	OtherUpdates         []interface{}
	Chats                []interface{}
	Users                []interface{}
	State                *State
	IntermediateState    *State
	Date                 int64
	Seq                  int
}

// ToTL converts DifferenceResult to TL format
func (d *DifferenceResult) ToTL() map[string]interface{} {
	result := map[string]interface{}{
		"_": d.Type,
	}
	
	switch d.Type {
	case "updates.differenceEmpty":
		result["date"] = d.Date
		result["seq"] = d.Seq
	case "updates.difference":
		result["new_messages"] = d.NewMessages
		result["new_encrypted_messages"] = d.NewEncryptedMessages
		result["other_updates"] = d.OtherUpdates
		result["chats"] = d.Chats
		result["users"] = d.Users
		result["state"] = d.State.ToTL()
	case "updates.differenceSlice":
		result["new_messages"] = d.NewMessages
		result["new_encrypted_messages"] = d.NewEncryptedMessages
		result["other_updates"] = d.OtherUpdates
		result["chats"] = d.Chats
		result["users"] = d.Users
		result["intermediate_state"] = d.IntermediateState.ToTL()
	}
	
	return result
}

// ToTL converts State to TL format
func (s *State) ToTL() map[string]interface{} {
	return map[string]interface{}{
		"_":            "updates.state",
		"pts":          s.Pts,
		"qts":          s.Qts,
		"seq":          s.Seq,
		"date":         s.Date,
		"unread_count": s.UnreadCount,
	}
}

// Helper functions
func extractMessages(updates []*Update) []interface{} {
	var messages []interface{}
	for _, u := range updates {
		if u.UpdateType == "updateNewMessage" || u.UpdateType == "updateEditMessage" {
			var data map[string]interface{}
			if err := json.Unmarshal(u.Data, &data); err == nil {
				if msg, ok := data["message"]; ok {
					messages = append(messages, msg)
				}
			}
		}
	}
	return messages
}

func extractOtherUpdates(updates []*Update) []interface{} {
	var others []interface{}
	for _, u := range updates {
		if u.UpdateType != "updateNewMessage" && u.UpdateType != "updateEditMessage" {
			var data map[string]interface{}
			if err := json.Unmarshal(u.Data, &data); err == nil {
				data["_"] = u.UpdateType
				data["pts"] = u.Pts
				others = append(others, data)
			}
		}
	}
	return others
}

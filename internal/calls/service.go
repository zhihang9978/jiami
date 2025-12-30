package calls

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"time"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// generateAccessHash generates a random access hash
func generateAccessHash() int64 {
	b := make([]byte, 8)
	rand.Read(b)
	var hash int64
	for i := 0; i < 8; i++ {
		hash = (hash << 8) | int64(b[i])
	}
	if hash < 0 {
		hash = -hash
	}
	return hash
}

// RequestCall initiates a new call
func (s *Service) RequestCall(ctx context.Context, callerID, calleeID int64, isVideo bool, protocol json.RawMessage, gaHash []byte) (*Call, error) {
	// Check if there's already an active call
	existing, err := s.repo.GetActiveCall(ctx, callerID, calleeID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing call: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("call already in progress")
	}

	call := &Call{
		AccessHash: generateAccessHash(),
		CallerID:   callerID,
		CalleeID:   calleeID,
		Date:       int(time.Now().Unix()),
		Duration:   0,
		IsVideo:    isVideo,
		IsOutgoing: true,
		State:      "pending",
		Protocol:   protocol,
		GAHash:     gaHash,
	}

	if err := s.repo.CreateCall(ctx, call); err != nil {
		return nil, fmt.Errorf("failed to create call: %w", err)
	}

	return call, nil
}

// AcceptCall accepts an incoming call
func (s *Service) AcceptCall(ctx context.Context, callID int64, userID int64, gb []byte, keyFingerprint int64) (*Call, error) {
	call, err := s.repo.GetCall(ctx, callID)
	if err != nil {
		return nil, fmt.Errorf("failed to get call: %w", err)
	}
	if call == nil {
		return nil, fmt.Errorf("call not found")
	}

	// Verify user is the callee
	if call.CalleeID != userID {
		return nil, fmt.Errorf("not authorized to accept this call")
	}

	// Verify call is in pending/ringing state
	if call.State != "pending" && call.State != "ringing" {
		return nil, fmt.Errorf("call cannot be accepted in state: %s", call.State)
	}

	call.State = "accepted"
	call.GB = gb
	call.KeyFingerprint = keyFingerprint

	if err := s.repo.UpdateCall(ctx, call); err != nil {
		return nil, fmt.Errorf("failed to update call: %w", err)
	}

	return call, nil
}

// DiscardCall ends or declines a call
func (s *Service) DiscardCall(ctx context.Context, callID int64, userID int64, reason string, duration int) (*Call, error) {
	call, err := s.repo.GetCall(ctx, callID)
	if err != nil {
		return nil, fmt.Errorf("failed to get call: %w", err)
	}
	if call == nil {
		return nil, fmt.Errorf("call not found")
	}

	// Verify user is part of the call
	if call.CallerID != userID && call.CalleeID != userID {
		return nil, fmt.Errorf("not authorized to discard this call")
	}

	// Determine final state based on reason and current state
	switch reason {
	case "hangup":
		call.State = "ended"
	case "busy":
		call.State = "busy"
	case "decline":
		call.State = "declined"
	case "missed":
		call.State = "missed"
	default:
		call.State = "ended"
	}

	call.Duration = duration

	if err := s.repo.UpdateCall(ctx, call); err != nil {
		return nil, fmt.Errorf("failed to update call: %w", err)
	}

	return call, nil
}

// ConfirmCall confirms call parameters (DH exchange)
func (s *Service) ConfirmCall(ctx context.Context, callID int64, userID int64, keyFingerprint int64) (*Call, error) {
	call, err := s.repo.GetCall(ctx, callID)
	if err != nil {
		return nil, fmt.Errorf("failed to get call: %w", err)
	}
	if call == nil {
		return nil, fmt.Errorf("call not found")
	}

	// Verify user is the caller
	if call.CallerID != userID {
		return nil, fmt.Errorf("not authorized to confirm this call")
	}

	call.KeyFingerprint = keyFingerprint

	if err := s.repo.UpdateCall(ctx, call); err != nil {
		return nil, fmt.Errorf("failed to update call: %w", err)
	}

	return call, nil
}

// SetCallRating sets the rating for a call
func (s *Service) SetCallRating(ctx context.Context, callID int64, userID int64, rating int, comment string) error {
	call, err := s.repo.GetCall(ctx, callID)
	if err != nil {
		return fmt.Errorf("failed to get call: %w", err)
	}
	if call == nil {
		return fmt.Errorf("call not found")
	}

	// Verify user is part of the call
	if call.CallerID != userID && call.CalleeID != userID {
		return fmt.Errorf("not authorized to rate this call")
	}

	// In a real implementation, we would store the rating
	// For now, we just validate the request
	if rating < 1 || rating > 5 {
		return fmt.Errorf("rating must be between 1 and 5")
	}

	return nil
}

// SaveCallDebug saves debug information for a call
func (s *Service) SaveCallDebug(ctx context.Context, callID int64, userID int64, debugData json.RawMessage) error {
	call, err := s.repo.GetCall(ctx, callID)
	if err != nil {
		return fmt.Errorf("failed to get call: %w", err)
	}
	if call == nil {
		return fmt.Errorf("call not found")
	}

	// Verify user is part of the call
	if call.CallerID != userID && call.CalleeID != userID {
		return fmt.Errorf("not authorized to save debug for this call")
	}

	// In a real implementation, we would store the debug data
	return nil
}

// GetCall retrieves a call by ID
func (s *Service) GetCall(ctx context.Context, callID int64) (*Call, error) {
	return s.repo.GetCall(ctx, callID)
}

// GetUserCalls retrieves calls for a user
func (s *Service) GetUserCalls(ctx context.Context, userID int64, offset, limit int) ([]*Call, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.repo.GetUserCalls(ctx, userID, offset, limit)
}

// GetPendingCalls retrieves pending incoming calls for a user
func (s *Service) GetPendingCalls(ctx context.Context, userID int64) ([]*Call, error) {
	return s.repo.GetPendingCallsForUser(ctx, userID)
}

// ReceivedCall marks a call as received (ringing)
func (s *Service) ReceivedCall(ctx context.Context, callID int64, userID int64) (*Call, error) {
	call, err := s.repo.GetCall(ctx, callID)
	if err != nil {
		return nil, fmt.Errorf("failed to get call: %w", err)
	}
	if call == nil {
		return nil, fmt.Errorf("call not found")
	}

	// Verify user is the callee
	if call.CalleeID != userID {
		return nil, fmt.Errorf("not authorized")
	}

	if call.State == "pending" {
		call.State = "ringing"
		if err := s.repo.UpdateCall(ctx, call); err != nil {
			return nil, fmt.Errorf("failed to update call: %w", err)
		}
	}

	return call, nil
}

// ToTL converts Call to TL format
func (c *Call) ToTL() map[string]interface{} {
	result := map[string]interface{}{
		"_":               "phoneCall",
		"id":              c.ID,
		"access_hash":     c.AccessHash,
		"date":            c.Date,
		"admin_id":        c.CallerID,
		"participant_id":  c.CalleeID,
		"protocol":        c.Protocol,
		"video":           c.IsVideo,
		"duration":        c.Duration,
		"key_fingerprint": c.KeyFingerprint,
	}

	switch c.State {
	case "pending":
		result["_"] = "phoneCallRequested"
		result["g_a_hash"] = c.GAHash
	case "ringing":
		result["_"] = "phoneCallRequested"
		result["g_a_hash"] = c.GAHash
	case "accepted":
		result["_"] = "phoneCallAccepted"
		result["g_b"] = c.GB
	case "ended", "missed", "declined", "busy":
		result["_"] = "phoneCallDiscarded"
		result["reason"] = c.getDiscardReason()
	}

	return result
}

func (c *Call) getDiscardReason() map[string]interface{} {
	switch c.State {
	case "missed":
		return map[string]interface{}{"_": "phoneCallDiscardReasonMissed"}
	case "declined":
		return map[string]interface{}{"_": "phoneCallDiscardReasonBusy"}
	case "busy":
		return map[string]interface{}{"_": "phoneCallDiscardReasonBusy"}
	default:
		return map[string]interface{}{"_": "phoneCallDiscardReasonHangup"}
	}
}

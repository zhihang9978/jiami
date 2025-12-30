package api

import (
	"net/http"

	"github.com/feiji/feiji-backend/internal/models"
	"github.com/feiji/feiji-backend/internal/users"
	"github.com/gin-gonic/gin"
)

// UsersHandler handles users-related requests
type UsersHandler struct {
	usersService *users.Service
}

func NewUsersHandler(usersService *users.Service) *UsersHandler {
	return &UsersHandler{
		usersService: usersService,
	}
}

// GetUsers handles users.getUsers
func (h *UsersHandler) GetUsers(c *gin.Context) {
	user := getUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		ID []struct {
			UserID     int64 `json:"user_id"`
			AccessHash int64 `json:"access_hash"`
		} `json:"id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDs := make([]int64, len(req.ID))
	for i, id := range req.ID {
		userIDs[i] = id.UserID
	}

	usersList, err := h.usersService.GetUsers(c.Request.Context(), userIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert to TL format
	tlUsers := make([]interface{}, len(usersList))
	for i, u := range usersList {
		tlUsers[i] = u.ToTLUser()
	}

	c.JSON(http.StatusOK, gin.H{
		"_":     "vector",
		"users": tlUsers,
	})
}

// GetFullUser handles users.getFullUser
func (h *UsersHandler) GetFullUser(c *gin.Context) {
	user := getUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		ID struct {
			UserID     int64 `json:"user_id"`
			AccessHash int64 `json:"access_hash"`
		} `json:"id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fullUser, err := h.usersService.GetFullUser(c.Request.Context(), req.ID.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, fullUser.ToTL())
}

// UpdateProfile handles account.updateProfile
func (h *UsersHandler) UpdateProfile(c *gin.Context) {
	user := getUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		About     string `json:"about"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Use existing values if not provided
	firstName := req.FirstName
	if firstName == "" {
		firstName = user.FirstName
	}
	lastName := req.LastName
	if lastName == "" {
		lastName = user.LastName
	}

	updatedUser, err := h.usersService.UpdateProfile(c.Request.Context(), user.ID, firstName, lastName, req.About)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedUser.ToTLUser())
}

// UpdateUsername handles account.updateUsername
func (h *UsersHandler) UpdateUsername(c *gin.Context) {
	user := getUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		Username string `json:"username" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedUser, err := h.usersService.UpdateUsername(c.Request.Context(), user.ID, req.Username)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedUser.ToTLUser())
}

// CheckUsername handles account.checkUsername
func (h *UsersHandler) CheckUsername(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	available, err := h.usersService.CheckUsername(c.Request.Context(), req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if available {
		c.JSON(http.StatusOK, gin.H{"_": "boolTrue"})
	} else {
		c.JSON(http.StatusOK, gin.H{"_": "boolFalse"})
	}
}

// UpdateStatus handles account.updateStatus
func (h *UsersHandler) UpdateStatus(c *gin.Context) {
	user := getUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		Offline bool `json:"offline"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.usersService.UpdateStatus(c.Request.Context(), user.ID, req.Offline); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"_": "boolTrue"})
}

// ResolveUsername handles contacts.resolveUsername
func (h *UsersHandler) ResolveUsername(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resolvedUser, err := h.usersService.ResolveUsername(c.Request.Context(), req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if resolvedUser == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "username not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"_": "contacts.resolvedPeer",
		"peer": gin.H{
			"_":       "peerUser",
			"user_id": resolvedUser.ID,
		},
		"chats": []interface{}{},
		"users": []interface{}{resolvedUser.ToTLUser()},
	})
}

// Search handles contacts.search
func (h *UsersHandler) Search(c *gin.Context) {
	var req struct {
		Q     string `json:"q" binding:"required"`
		Limit int    `json:"limit"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	limit := req.Limit
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	usersList, err := h.usersService.SearchUsers(c.Request.Context(), req.Q, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert to TL format
	tlUsers := make([]interface{}, len(usersList))
	tlResults := make([]interface{}, len(usersList))
	for i, u := range usersList {
		tlUsers[i] = u.ToTLUser()
		tlResults[i] = gin.H{
			"_":       "peerUser",
			"user_id": u.ID,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"_":            "contacts.found",
		"my_results":   []interface{}{},
		"results":      tlResults,
		"chats":        []interface{}{},
		"users":        tlUsers,
	})
}

// Helper function
func getUserFromContext(c *gin.Context) *models.User {
	user, exists := c.Get("user")
	if !exists {
		return nil
	}
	return user.(*models.User)
}

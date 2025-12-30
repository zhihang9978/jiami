package api

import (
	"net/http"

	"github.com/feiji/feiji-backend/internal/contacts"
	"github.com/feiji/feiji-backend/internal/models"
	"github.com/gin-gonic/gin"
)

// ContactsHandler handles contacts-related requests
type ContactsHandler struct {
	contactsService *contacts.Service
}

func NewContactsHandler(contactsService *contacts.Service) *ContactsHandler {
	return &ContactsHandler{
		contactsService: contactsService,
	}
}

// GetContacts handles contacts.getContacts
func (h *ContactsHandler) GetContacts(c *gin.Context) {
	user := getCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	contactsList, err := h.contactsService.GetContacts(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert to TL format
	tlContacts := make([]interface{}, len(contactsList))
	tlUsers := make([]interface{}, len(contactsList))
	for i, contact := range contactsList {
		tlContacts[i] = contact.ToTLContact()
		tlUsers[i] = contact.ToTLUser()
	}

	c.JSON(http.StatusOK, gin.H{
		"_":                "contacts.contacts",
		"contacts":         tlContacts,
		"saved_count":      len(contactsList),
		"users":            tlUsers,
	})
}

// ImportContacts handles contacts.importContacts
func (h *ContactsHandler) ImportContacts(c *gin.Context) {
	user := getCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		Contacts []contacts.ImportContact `json:"contacts" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.contactsService.ImportContacts(c.Request.Context(), user.ID, req.Contacts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert imported contacts
	imported := make([]interface{}, len(result.Imported))
	for i, imp := range result.Imported {
		imported[i] = gin.H{
			"_":         "importedContact",
			"user_id":   imp.UserID,
			"client_id": imp.ClientID,
		}
	}

	// Convert users
	users := make([]interface{}, len(result.Users))
	for i, u := range result.Users {
		users[i] = u.ToTLUser()
	}

	c.JSON(http.StatusOK, gin.H{
		"_":                "contacts.importedContacts",
		"imported":         imported,
		"popular_invites":  []interface{}{},
		"retry_contacts":   []interface{}{},
		"users":            users,
	})
}

// AddContact handles contacts.addContact
func (h *ContactsHandler) AddContact(c *gin.Context) {
	user := getCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		ID struct {
			UserID int64 `json:"user_id"`
		} `json:"id" binding:"required"`
		FirstName string `json:"first_name" binding:"required"`
		LastName  string `json:"last_name"`
		Phone     string `json:"phone"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.contactsService.AddContact(c.Request.Context(), user.ID, req.ID.UserID, req.FirstName, req.LastName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"_": "updates",
		"updates": []interface{}{},
		"users":   []interface{}{},
		"chats":   []interface{}{},
		"date":    0,
		"seq":     0,
	})
}

// DeleteContacts handles contacts.deleteContacts
func (h *ContactsHandler) DeleteContacts(c *gin.Context) {
	user := getCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		ID []struct {
			UserID int64 `json:"user_id"`
		} `json:"id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	contactIDs := make([]int64, len(req.ID))
	for i, id := range req.ID {
		contactIDs[i] = id.UserID
	}

	if err := h.contactsService.DeleteContacts(c.Request.Context(), user.ID, contactIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"_": "updates",
		"updates": []interface{}{},
		"users":   []interface{}{},
		"chats":   []interface{}{},
		"date":    0,
		"seq":     0,
	})
}

// Block handles contacts.block
func (h *ContactsHandler) Block(c *gin.Context) {
	user := getCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		ID struct {
			UserID int64 `json:"user_id"`
		} `json:"id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.contactsService.BlockContact(c.Request.Context(), user.ID, req.ID.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"_": "boolTrue",
	})
}

// Unblock handles contacts.unblock
func (h *ContactsHandler) Unblock(c *gin.Context) {
	user := getCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		ID struct {
			UserID int64 `json:"user_id"`
		} `json:"id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.contactsService.UnblockContact(c.Request.Context(), user.ID, req.ID.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"_": "boolTrue",
	})
}

// GetBlocked handles contacts.getBlocked
func (h *ContactsHandler) GetBlocked(c *gin.Context) {
	user := getCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	blocked, err := h.contactsService.GetBlocked(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert to TL format
	tlBlocked := make([]interface{}, len(blocked))
	tlUsers := make([]interface{}, len(blocked))
	for i, b := range blocked {
		tlBlocked[i] = gin.H{
			"_": "peerBlocked",
			"peer_id": gin.H{
				"_":       "peerUser",
				"user_id": b.ContactID,
			},
			"date": 0,
		}
		tlUsers[i] = b.ToTLUser()
	}

	c.JSON(http.StatusOK, gin.H{
		"_":       "contacts.blocked",
		"blocked": tlBlocked,
		"chats":   []interface{}{},
		"users":   tlUsers,
	})
}

// Helper function
func getCurrentUser(c *gin.Context) *models.User {
	user, exists := c.Get("user")
	if !exists {
		return nil
	}
	return user.(*models.User)
}

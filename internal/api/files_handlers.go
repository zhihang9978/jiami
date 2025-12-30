package api

import (
	"encoding/base64"
	"net/http"

	"github.com/feiji/feiji-backend/internal/files"
	"github.com/feiji/feiji-backend/internal/models"
	"github.com/gin-gonic/gin"
)

// FilesHandler handles file upload/download requests
type FilesHandler struct {
	filesService *files.Service
}

func NewFilesHandler(filesService *files.Service) *FilesHandler {
	return &FilesHandler{
		filesService: filesService,
	}
}

// SaveFilePart handles upload.saveFilePart
func (h *FilesHandler) SaveFilePart(c *gin.Context) {
	user := getFileUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		FileID   int64  `json:"file_id" binding:"required"`
		FilePart int    `json:"file_part" binding:"required"`
		Bytes    string `json:"bytes" binding:"required"` // Base64 encoded
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Decode base64 data
	data, err := base64.StdEncoding.DecodeString(req.Bytes)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid base64 data"})
		return
	}

	success, err := h.filesService.SaveFilePart(c.Request.Context(), req.FileID, req.FilePart, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if success {
		c.JSON(http.StatusOK, gin.H{"_": "boolTrue"})
	} else {
		c.JSON(http.StatusOK, gin.H{"_": "boolFalse"})
	}
}

// SaveBigFilePart handles upload.saveBigFilePart
func (h *FilesHandler) SaveBigFilePart(c *gin.Context) {
	user := getFileUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		FileID      int64  `json:"file_id" binding:"required"`
		FilePart    int    `json:"file_part" binding:"required"`
		FileTotalParts int `json:"file_total_parts" binding:"required"`
		Bytes       string `json:"bytes" binding:"required"` // Base64 encoded
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Decode base64 data
	data, err := base64.StdEncoding.DecodeString(req.Bytes)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid base64 data"})
		return
	}

	success, err := h.filesService.SaveBigFilePart(c.Request.Context(), req.FileID, req.FilePart, req.FileTotalParts, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if success {
		c.JSON(http.StatusOK, gin.H{"_": "boolTrue"})
	} else {
		c.JSON(http.StatusOK, gin.H{"_": "boolFalse"})
	}
}

// GetFile handles upload.getFile
func (h *FilesHandler) GetFile(c *gin.Context) {
	var req struct {
		Location struct {
			FileID     int64 `json:"id"`
			AccessHash int64 `json:"access_hash"`
		} `json:"location" binding:"required"`
		Offset int `json:"offset"`
		Limit  int `json:"limit"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	data, err := h.filesService.GetFileContent(c.Request.Context(), req.Location.FileID, req.Location.AccessHash, req.Offset, req.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"_":    "upload.file",
		"type": gin.H{"_": "storage.filePartial"},
		"mtime": 0,
		"bytes": base64.StdEncoding.EncodeToString(data),
	})
}

// CompleteUpload handles completing a multipart upload
func (h *FilesHandler) CompleteUpload(c *gin.Context) {
	user := getFileUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		FileID     int64  `json:"file_id" binding:"required"`
		FileName   string `json:"file_name" binding:"required"`
		MimeType   string `json:"mime_type"`
		TotalParts int    `json:"total_parts"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.MimeType == "" {
		req.MimeType = "application/octet-stream"
	}

	file, err := h.filesService.CompleteUpload(c.Request.Context(), user.ID, req.FileID, req.FileName, req.MimeType, req.TotalParts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, file.ToTLDocument())
}

// GetNextFileID handles getting a new file ID for upload
func (h *FilesHandler) GetNextFileID(c *gin.Context) {
	fileID, err := h.filesService.GetNextFileID(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"file_id": fileID,
	})
}

// Helper function
func getFileUser(c *gin.Context) *models.User {
	user, exists := c.Get("user")
	if !exists {
		return nil
	}
	return user.(*models.User)
}

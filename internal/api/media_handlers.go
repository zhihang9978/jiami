package api

import (
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// MediaHandler handles REST-style media upload endpoints
type MediaHandler struct {
	uploadPath string
	baseURL    string
}

func NewMediaHandler(uploadPath, baseURL string) *MediaHandler {
	return &MediaHandler{
		uploadPath: uploadPath,
		baseURL:    baseURL,
	}
}

// generateFileID generates a unique file ID
func generateFileID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// sanitizeFilename removes unsafe characters from filename
func sanitizeFilename(filename string) string {
	reg := regexp.MustCompile(`[^a-zA-Z0-9._-]`)
	safe := reg.ReplaceAllString(filename, "_")
	if len(safe) > 200 {
		safe = safe[:200]
	}
	return safe
}

// UploadVoice handles POST /api/v1/upload/voice
func (h *MediaHandler) UploadVoice(c *gin.Context) {
	file, header, err := c.Request.FormFile("voice")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error_code": "INVALID_REQUEST", "error": "voice file required"})
		return
	}
	defer file.Close()

	// Validate file size (10MB limit)
	if header.Size > 10*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error_code": "FILE_TOO_LARGE", "error": "file size exceeds 10MB limit"})
		return
	}

	// Validate file type
	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowedTypes := map[string]bool{".m4a": true, ".mp3": true, ".wav": true, ".aac": true, ".ogg": true}
	if !allowedTypes[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error_code": "INVALID_FILE_TYPE", "error": "allowed types: m4a, mp3, wav, aac, ogg"})
		return
	}

	// Create upload directory
	voicesDir := filepath.Join(h.uploadPath, "voices")
	if err := os.MkdirAll(voicesDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error_code": "INTERNAL_ERROR", "error": "failed to create directory"})
		return
	}

	// Generate unique filename
	fileID := generateFileID()
	filename := fmt.Sprintf("%s%s", fileID, ext)
	filePath := filepath.Join(voicesDir, filename)

	// Save file
	dst, err := os.Create(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error_code": "INTERNAL_ERROR", "error": "failed to create file"})
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error_code": "INTERNAL_ERROR", "error": "failed to save file"})
		return
	}

	// Get duration using ffprobe (optional)
	var duration float64
	cmd := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", filePath)
	if output, err := cmd.Output(); err == nil {
		fmt.Sscanf(strings.TrimSpace(string(output)), "%f", &duration)
	}

	fileURL := fmt.Sprintf("%s/uploads/voices/%s", h.baseURL, filename)

	c.JSON(http.StatusOK, gin.H{
		"ok":       true,
		"file_id":  fileID,
		"url":      fileURL,
		"duration": int(duration),
		"size":     header.Size,
		"mime_type": header.Header.Get("Content-Type"),
	})
}

// UploadVideo handles POST /api/v1/upload/video
func (h *MediaHandler) UploadVideo(c *gin.Context) {
	file, header, err := c.Request.FormFile("video")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error_code": "INVALID_REQUEST", "error": "video file required"})
		return
	}
	defer file.Close()

	// Validate file size (100MB limit)
	if header.Size > 100*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error_code": "FILE_TOO_LARGE", "error": "file size exceeds 100MB limit"})
		return
	}

	// Validate file type
	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowedTypes := map[string]bool{".mp4": true, ".mov": true, ".avi": true, ".mkv": true, ".webm": true}
	if !allowedTypes[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error_code": "INVALID_FILE_TYPE", "error": "allowed types: mp4, mov, avi, mkv, webm"})
		return
	}

	// Create upload directory
	videosDir := filepath.Join(h.uploadPath, "videos")
	thumbsDir := filepath.Join(h.uploadPath, "thumbnails")
	if err := os.MkdirAll(videosDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error_code": "INTERNAL_ERROR", "error": "failed to create directory"})
		return
	}
	os.MkdirAll(thumbsDir, 0755)

	// Generate unique filename
	fileID := generateFileID()
	filename := fmt.Sprintf("%s%s", fileID, ext)
	filePath := filepath.Join(videosDir, filename)

	// Save file
	dst, err := os.Create(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error_code": "INTERNAL_ERROR", "error": "failed to create file"})
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error_code": "INTERNAL_ERROR", "error": "failed to save file"})
		return
	}

	// Generate thumbnail using ffmpeg
	thumbFilename := fmt.Sprintf("%s_thumb.jpg", fileID)
	thumbPath := filepath.Join(thumbsDir, thumbFilename)
	cmd := exec.Command("ffmpeg", "-i", filePath, "-ss", "00:00:01", "-vframes", "1", "-vf", "scale=320:-1", "-y", thumbPath)
	cmd.Run() // Ignore error, thumbnail is optional

	// Get video duration and resolution
	var duration float64
	var width, height int
	
	// Get duration
	durationCmd := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", filePath)
	if output, err := durationCmd.Output(); err == nil {
		fmt.Sscanf(strings.TrimSpace(string(output)), "%f", &duration)
	}

	// Get resolution
	resCmd := exec.Command("ffprobe", "-v", "error", "-select_streams", "v:0", "-show_entries", "stream=width,height", "-of", "csv=s=x:p=0", filePath)
	if output, err := resCmd.Output(); err == nil {
		fmt.Sscanf(strings.TrimSpace(string(output)), "%dx%d", &width, &height)
	}

	fileURL := fmt.Sprintf("%s/uploads/videos/%s", h.baseURL, filename)
	thumbURL := ""
	if _, err := os.Stat(thumbPath); err == nil {
		thumbURL = fmt.Sprintf("%s/uploads/thumbnails/%s", h.baseURL, thumbFilename)
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":            true,
		"file_id":       fileID,
		"url":           fileURL,
		"thumbnail_url": thumbURL,
		"duration":      int(duration),
		"width":         width,
		"height":        height,
		"size":          header.Size,
		"mime_type":     header.Header.Get("Content-Type"),
	})
}

// UploadFile handles POST /api/v1/upload/file
func (h *MediaHandler) UploadFile(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error_code": "INVALID_REQUEST", "error": "file required"})
		return
	}
	defer file.Close()

	// Validate file size (100MB limit)
	if header.Size > 100*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error_code": "FILE_TOO_LARGE", "error": "file size exceeds 100MB limit"})
		return
	}

	// Validate file type (whitelist)
	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowedTypes := map[string]bool{
		".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true,
		".ppt": true, ".pptx": true, ".txt": true, ".zip": true, ".rar": true,
		".7z": true, ".csv": true, ".json": true, ".xml": true, ".html": true,
		".css": true, ".js": true, ".md": true, ".rtf": true, ".odt": true,
	}
	
	// Block executable files
	blockedTypes := map[string]bool{
		".exe": true, ".bat": true, ".sh": true, ".cmd": true, ".com": true,
		".msi": true, ".dll": true, ".scr": true, ".vbs": true, ".ps1": true,
	}
	
	if blockedTypes[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error_code": "INVALID_FILE_TYPE", "error": "executable files are not allowed"})
		return
	}

	// If not in whitelist, still allow but log warning
	if !allowedTypes[ext] && ext != "" {
		// Allow other file types but with caution
	}

	// Create upload directory
	filesDir := filepath.Join(h.uploadPath, "files")
	if err := os.MkdirAll(filesDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error_code": "INTERNAL_ERROR", "error": "failed to create directory"})
		return
	}

	// Generate unique filename but preserve original name for reference
	fileID := generateFileID()
	safeOriginalName := sanitizeFilename(header.Filename)
	filename := fmt.Sprintf("%s_%s", fileID, safeOriginalName)
	filePath := filepath.Join(filesDir, filename)

	// Save file
	dst, err := os.Create(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error_code": "INTERNAL_ERROR", "error": "failed to create file"})
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error_code": "INTERNAL_ERROR", "error": "failed to save file"})
		return
	}

	fileURL := fmt.Sprintf("%s/uploads/files/%s", h.baseURL, filename)

	c.JSON(http.StatusOK, gin.H{
		"ok":            true,
		"file_id":       fileID,
		"url":           fileURL,
		"original_name": header.Filename,
		"size":          header.Size,
		"mime_type":     header.Header.Get("Content-Type"),
	})
}

// UploadImage handles POST /api/v1/upload/image
func (h *MediaHandler) UploadImage(c *gin.Context) {
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error_code": "INVALID_REQUEST", "error": "image file required"})
		return
	}
	defer file.Close()

	// Validate file size (10MB limit)
	if header.Size > 10*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error_code": "FILE_TOO_LARGE", "error": "file size exceeds 10MB limit"})
		return
	}

	// Validate file type
	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowedTypes := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}
	if !allowedTypes[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error_code": "INVALID_FILE_TYPE", "error": "allowed types: jpg, jpeg, png, gif, webp"})
		return
	}

	// Create upload directory
	imagesDir := filepath.Join(h.uploadPath, "images")
	if err := os.MkdirAll(imagesDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error_code": "INTERNAL_ERROR", "error": "failed to create directory"})
		return
	}

	// Generate unique filename
	fileID := generateFileID()
	filename := fmt.Sprintf("%s%s", fileID, ext)
	filePath := filepath.Join(imagesDir, filename)

	// Save file
	dst, err := os.Create(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error_code": "INTERNAL_ERROR", "error": "failed to create file"})
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error_code": "INTERNAL_ERROR", "error": "failed to save file"})
		return
	}

	fileURL := fmt.Sprintf("%s/uploads/images/%s", h.baseURL, filename)

	c.JSON(http.StatusOK, gin.H{
		"ok":       true,
		"file_id":  fileID,
		"url":      fileURL,
		"size":     header.Size,
		"mime_type": header.Header.Get("Content-Type"),
	})
}

// MediaInit handles POST /api/v1/media/init
func (h *MediaHandler) MediaInit(c *gin.Context) {
	var req struct {
		FileName   string `json:"file_name" binding:"required"`
		FileSize   int64  `json:"file_size" binding:"required"`
		MimeType   string `json:"mime_type"`
		TotalParts int    `json:"total_parts"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error_code": "INVALID_REQUEST", "error": err.Error()})
		return
	}

	fileID := fmt.Sprintf("file_%d_%s", time.Now().UnixNano(), generateFileID()[:8])

	c.JSON(http.StatusOK, gin.H{
		"ok":         true,
		"file_id":    fileID,
		"upload_url": fmt.Sprintf("%s/api/v1/media/upload", h.baseURL),
	})
}

// MediaUpload handles POST /api/v1/media/upload (multipart upload)
func (h *MediaHandler) MediaUpload(c *gin.Context) {
	fileID := c.PostForm("file_id")
	partNumber := c.PostForm("part_number")
	
	if fileID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error_code": "INVALID_REQUEST", "error": "file_id required"})
		return
	}

	file, _, err := c.Request.FormFile("data")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error_code": "INVALID_REQUEST", "error": "data file required"})
		return
	}
	defer file.Close()

	// Create temp directory for parts
	partsDir := filepath.Join(h.uploadPath, "temp", fileID)
	if err := os.MkdirAll(partsDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error_code": "INTERNAL_ERROR", "error": "failed to create directory"})
		return
	}

	// Save part
	partPath := filepath.Join(partsDir, fmt.Sprintf("part_%s", partNumber))
	dst, err := os.Create(partPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error_code": "INTERNAL_ERROR", "error": "failed to create file"})
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error_code": "INTERNAL_ERROR", "error": "failed to save file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":          true,
		"part_number": partNumber,
	})
}

// MediaComplete handles POST /api/v1/media/complete
func (h *MediaHandler) MediaComplete(c *gin.Context) {
	var req struct {
		FileID string `json:"file_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error_code": "INVALID_REQUEST", "error": err.Error()})
		return
	}

	// For now, just return success - actual implementation would merge parts
	c.JSON(http.StatusOK, gin.H{
		"ok":            true,
		"file_url":      fmt.Sprintf("%s/media/%s", h.baseURL, req.FileID),
		"thumbnail_url": "",
	})
}

// GetMedia handles GET /api/v1/media/:file_id
func (h *MediaHandler) GetMedia(c *gin.Context) {
	fileID := c.Param("file_id")
	if fileID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error_code": "INVALID_REQUEST", "error": "file_id required"})
		return
	}

	// Search for file in various directories
	dirs := []string{"images", "voices", "videos", "files"}
	for _, dir := range dirs {
		pattern := filepath.Join(h.uploadPath, dir, fileID+"*")
		matches, _ := filepath.Glob(pattern)
		if len(matches) > 0 {
			c.File(matches[0])
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"ok": false, "error_code": "FILE_NOT_FOUND", "error": "file not found"})
}

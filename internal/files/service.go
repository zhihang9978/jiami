package files

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

type Service struct {
	repo       *Repository
	uploadPath string
	baseURL    string
}

func NewService(repo *Repository, uploadPath, baseURL string) *Service {
	return &Service{
		repo:       repo,
		uploadPath: uploadPath,
		baseURL:    baseURL,
	}
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

// SaveFilePart saves a file part for multipart upload
func (s *Service) SaveFilePart(ctx context.Context, fileID int64, partNum int, data []byte) (bool, error) {
	if err := s.repo.SaveFilePart(ctx, fileID, partNum, data); err != nil {
		return false, fmt.Errorf("failed to save file part: %w", err)
	}
	return true, nil
}

// SaveBigFilePart saves a big file part (same as SaveFilePart but for larger files)
func (s *Service) SaveBigFilePart(ctx context.Context, fileID int64, partNum, totalParts int, data []byte) (bool, error) {
	return s.SaveFilePart(ctx, fileID, partNum, data)
}

// CompleteUpload merges all file parts and creates the final file
func (s *Service) CompleteUpload(ctx context.Context, userID, fileID int64, fileName, mimeType string, totalParts int) (*File, error) {
	// Get all parts
	parts, err := s.repo.GetFileParts(ctx, fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get file parts: %w", err)
	}

	if len(parts) == 0 {
		return nil, fmt.Errorf("no file parts found")
	}

	// Merge parts
	var buffer bytes.Buffer
	for _, part := range parts {
		buffer.Write(part.Data)
	}

	// Generate file path
	dateDir := time.Now().Format("2006/01/02")
	randomName := generateRandomFileName()
	ext := filepath.Ext(fileName)
	if ext == "" {
		ext = ".bin"
	}
	relativePath := filepath.Join(dateDir, randomName+ext)
	fullPath := filepath.Join(s.uploadPath, relativePath)

	// Create directory
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(fullPath, buffer.Bytes(), 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	// Create file record
	file := &File{
		UserID:     userID,
		FileID:     fileID,
		AccessHash: generateAccessHash(),
		FileName:   fileName,
		MimeType:   mimeType,
		Size:       int64(buffer.Len()),
		FilePath:   relativePath,
	}

	if err := s.repo.CreateFile(ctx, file); err != nil {
		// Clean up file on error
		os.Remove(fullPath)
		return nil, fmt.Errorf("failed to create file record: %w", err)
	}

	// Clean up parts
	s.repo.DeleteFileParts(ctx, fileID)

	return file, nil
}

// GetFile retrieves file information
func (s *Service) GetFile(ctx context.Context, fileID, accessHash int64) (*File, error) {
	file, err := s.repo.GetFileByAccessHash(ctx, fileID, accessHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %w", err)
	}
	return file, nil
}

// GetFileContent retrieves file content with offset and limit
func (s *Service) GetFileContent(ctx context.Context, fileID, accessHash int64, offset, limit int) ([]byte, error) {
	file, err := s.repo.GetFileByAccessHash(ctx, fileID, accessHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %w", err)
	}
	if file == nil {
		return nil, fmt.Errorf("file not found")
	}

	fullPath := filepath.Join(s.uploadPath, file.FilePath)
	f, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	// Seek to offset
	if offset > 0 {
		if _, err := f.Seek(int64(offset), io.SeekStart); err != nil {
			return nil, fmt.Errorf("failed to seek: %w", err)
		}
	}

	// Read limit bytes
	if limit <= 0 {
		limit = 1024 * 1024 // Default 1MB
	}
	data := make([]byte, limit)
	n, err := f.Read(data)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return data[:n], nil
}

// GetNextFileID generates a new file ID for upload
func (s *Service) GetNextFileID(ctx context.Context) (int64, error) {
	return s.repo.GetNextFileID(ctx)
}

// DeleteFile removes a file
func (s *Service) DeleteFile(ctx context.Context, fileID int64) error {
	file, err := s.repo.GetFile(ctx, fileID)
	if err != nil {
		return fmt.Errorf("failed to get file: %w", err)
	}
	if file == nil {
		return nil
	}

	// Delete physical file
	fullPath := filepath.Join(s.uploadPath, file.FilePath)
	os.Remove(fullPath)

	// Delete record
	return s.repo.DeleteFile(ctx, fileID)
}

// GetUserFiles retrieves all files for a user
func (s *Service) GetUserFiles(ctx context.Context, userID int64, limit, offset int) ([]*File, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.repo.GetUserFiles(ctx, userID, limit, offset)
}

// ToTLDocument converts File to TL document format
func (f *File) ToTLDocument() map[string]interface{} {
	return map[string]interface{}{
		"_":           "document",
		"id":          f.FileID,
		"access_hash": f.AccessHash,
		"file_reference": []byte{},
		"date":        f.CreatedAt.Unix(),
		"mime_type":   f.MimeType,
		"size":        f.Size,
		"attributes": []map[string]interface{}{
			{
				"_":         "documentAttributeFilename",
				"file_name": f.FileName,
			},
		},
		"dc_id":   1,
		"thumbs":  []interface{}{},
	}
}

// ToTLPhoto converts File to TL photo format (for images)
func (f *File) ToTLPhoto() map[string]interface{} {
	return map[string]interface{}{
		"_":           "photo",
		"id":          f.FileID,
		"access_hash": f.AccessHash,
		"file_reference": []byte{},
		"date":        f.CreatedAt.Unix(),
		"sizes": []map[string]interface{}{
			{
				"_":    "photoSize",
				"type": "x",
				"w":    f.Width,
				"h":    f.Height,
				"size": f.Size,
			},
		},
		"dc_id": 1,
	}
}

// generateRandomFileName generates a random file name
func generateRandomFileName() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

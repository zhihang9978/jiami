package files

import (
	"context"
	"database/sql"
	"time"
)

type File struct {
	ID          int64
	UserID      int64
	FileID      int64
	AccessHash  int64
	FileName    string
	MimeType    string
	Size        int64
	FilePath    string
	ThumbnailPath string
	Width       int
	Height      int
	Duration    int
	CreatedAt   time.Time
}

type FilePart struct {
	ID        int64
	FileID    int64
	PartNum   int
	Data      []byte
	CreatedAt time.Time
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// GetNextFileID generates a unique file ID
func (r *Repository) GetNextFileID(ctx context.Context) (int64, error) {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO file_id_seq (stub) VALUES ('a')
		ON DUPLICATE KEY UPDATE id = LAST_INSERT_ID(id + 1)
	`)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// SaveFilePart saves a file part for multipart upload
func (r *Repository) SaveFilePart(ctx context.Context, fileID int64, partNum int, data []byte) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO file_parts (file_id, part_num, data, created_at)
		VALUES (?, ?, ?, NOW())
		ON DUPLICATE KEY UPDATE data = VALUES(data)
	`, fileID, partNum, data)
	return err
}

// GetFileParts retrieves all parts for a file
func (r *Repository) GetFileParts(ctx context.Context, fileID int64) ([]*FilePart, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, file_id, part_num, data, created_at
		FROM file_parts
		WHERE file_id = ?
		ORDER BY part_num
	`, fileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var parts []*FilePart
	for rows.Next() {
		part := &FilePart{}
		if err := rows.Scan(&part.ID, &part.FileID, &part.PartNum, &part.Data, &part.CreatedAt); err != nil {
			return nil, err
		}
		parts = append(parts, part)
	}
	return parts, nil
}

// DeleteFileParts removes all parts for a file
func (r *Repository) DeleteFileParts(ctx context.Context, fileID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM file_parts WHERE file_id = ?`, fileID)
	return err
}

// CreateFile creates a new file record
func (r *Repository) CreateFile(ctx context.Context, file *File) error {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO files (user_id, file_id, access_hash, file_name, mime_type, size, file_path, 
		                   thumbnail_path, width, height, duration, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW())
	`, file.UserID, file.FileID, file.AccessHash, file.FileName, file.MimeType, file.Size,
		file.FilePath, file.ThumbnailPath, file.Width, file.Height, file.Duration)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	file.ID = id
	return nil
}

// GetFile retrieves a file by file_id
func (r *Repository) GetFile(ctx context.Context, fileID int64) (*File, error) {
	file := &File{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, file_id, access_hash, file_name, mime_type, size, file_path,
		       thumbnail_path, width, height, duration, created_at
		FROM files
		WHERE file_id = ?
	`, fileID).Scan(&file.ID, &file.UserID, &file.FileID, &file.AccessHash, &file.FileName,
		&file.MimeType, &file.Size, &file.FilePath, &file.ThumbnailPath, &file.Width,
		&file.Height, &file.Duration, &file.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return file, nil
}

// GetFileByAccessHash retrieves a file by file_id and access_hash
func (r *Repository) GetFileByAccessHash(ctx context.Context, fileID, accessHash int64) (*File, error) {
	file := &File{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, file_id, access_hash, file_name, mime_type, size, file_path,
		       thumbnail_path, width, height, duration, created_at
		FROM files
		WHERE file_id = ? AND access_hash = ?
	`, fileID, accessHash).Scan(&file.ID, &file.UserID, &file.FileID, &file.AccessHash, &file.FileName,
		&file.MimeType, &file.Size, &file.FilePath, &file.ThumbnailPath, &file.Width,
		&file.Height, &file.Duration, &file.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return file, nil
}

// GetUserFiles retrieves all files for a user
func (r *Repository) GetUserFiles(ctx context.Context, userID int64, limit, offset int) ([]*File, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, file_id, access_hash, file_name, mime_type, size, file_path,
		       thumbnail_path, width, height, duration, created_at
		FROM files
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []*File
	for rows.Next() {
		file := &File{}
		if err := rows.Scan(&file.ID, &file.UserID, &file.FileID, &file.AccessHash, &file.FileName,
			&file.MimeType, &file.Size, &file.FilePath, &file.ThumbnailPath, &file.Width,
			&file.Height, &file.Duration, &file.CreatedAt); err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	return files, nil
}

// DeleteFile removes a file record
func (r *Repository) DeleteFile(ctx context.Context, fileID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM files WHERE file_id = ?`, fileID)
	return err
}

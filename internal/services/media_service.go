package services

import (
	"context"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/opencode-ai/polaris/internal/model"
)

type MediaService struct {
	pg       *pgxpool.Pool
	minio    *minio.Client
	bucket   string
	endpoint string
}

func NewMediaService(pg *pgxpool.Pool) (*MediaService, error) {
	endpoint := lookupEnv("MINIO_ENDPOINT", "http://127.0.0.1:9000")
	accessKey := lookupEnv("MINIO_ACCESS_KEY", "tl-images")
	secretKey := lookupEnv("MINIO_SECRET_KEY", "27kCGLzoYXoqr74TyuWvUywlbgzqyw")
	bucket := lookupEnv("MINIO_BUCKET", "threadlight-images")

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
	})
	if err != nil {
		return nil, fmt.Errorf("media_service: failed to create minio client: %w", err)
	}

	return &MediaService{pg: pg, minio: client, bucket: bucket, endpoint: endpoint}, nil
}

func lookupEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

// UploadImage uploads a file to MinIO and records it in the media table.
func (s *MediaService) UploadImage(ctx context.Context, uploaderID int64, file multipart.File, header *multipart.FileHeader) (*model.Media, error) {
	// Generate a unique file path
	ext := filepath.Ext(header.Filename)
	objectName := fmt.Sprintf("uploads/%d/%s%s", uploaderID, time.Now().Format("20060102150405"), ext)

	// Upload to MinIO
	_, err := s.minio.PutObject(ctx, s.bucket, objectName, file, header.Size, minio.PutObjectOptions{
		ContentType: header.Header.Get("Content-Type"),
	})
	if err != nil {
		return nil, fmt.Errorf("media_service: failed to upload to minio: %w", err)
	}

	// Record in database
	media := &model.Media{}
	err = s.pg.QueryRow(ctx,
		`INSERT INTO media (uploader_id, file_path, original_name, mime_type, file_size)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, post_id, uploader_id, file_path, original_name, mime_type, file_size, width, height, created_at`,
		uploaderID, objectName, header.Filename, header.Header.Get("Content-Type"), header.Size,
	).Scan(&media.ID, &media.PostID, &media.UploaderID, &media.FilePath,
		&media.OriginalName, &media.MimeType, &media.FileSize, &media.Width, &media.Height, &media.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("media_service: failed to insert media record: %w", err)
	}

	return media, nil
}

// GetImageURL returns the public URL for an image.
func (s *MediaService) GetImageURL(ctx context.Context, mediaID int64) (string, error) {
	var filePath string
	err := s.pg.QueryRow(ctx,
		`SELECT file_path FROM media WHERE id = $1`, mediaID,
	).Scan(&filePath)
	if err != nil {
		return "", fmt.Errorf("media_service: media not found: %w", err)
	}

	// Construct a direct MinIO URL
	baseURL := strings.TrimSuffix(s.endpoint, "/")
	return fmt.Sprintf("%s/%s/%s", baseURL, s.bucket, filePath), nil
}

// LinkImageToPost associates a media item with a post.
func (s *MediaService) LinkImageToPost(ctx context.Context, mediaID, postID int64) error {
	_, err := s.pg.Exec(ctx,
		`UPDATE media SET post_id = $1 WHERE id = $2`,
		postID, mediaID)
	if err != nil {
		return fmt.Errorf("media_service: failed to link image to post: %w", err)
	}
	return nil
}

// GetPostImages returns all images for a post.
func (s *MediaService) GetPostImages(ctx context.Context, postID int64) ([]model.Media, error) {
	rows, err := s.pg.Query(ctx,
		`SELECT id, post_id, uploader_id, file_path, original_name, mime_type,
		        file_size, width, height, created_at
		 FROM media WHERE post_id = $1
		 ORDER BY created_at DESC`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []model.Media
	for rows.Next() {
		var m model.Media
		if err := rows.Scan(&m.ID, &m.PostID, &m.UploaderID, &m.FilePath,
			&m.OriginalName, &m.MimeType, &m.FileSize, &m.Width, &m.Height, &m.CreatedAt); err != nil {
			return nil, err
		}
		images = append(images, m)
	}
	return images, nil
}

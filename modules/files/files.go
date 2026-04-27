package files

import (
	"context"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Gooowan/matchup/modules/files/gen"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type File struct {
	Bucket      string
	Key         string
	Size        int64
	Reader      io.Reader
	ContentType string
}

type FileService struct {
	PublicEndpoint string
	Client         *minio.Client
	PublicClient   *minio.Client // Client configured with public endpoint for presigned URLs
	DB             *pgxpool.Pool
	Queries        *gen.Queries
}

func NewFileService(db *pgxpool.Pool) (*FileService, error) {
	public_endpoint := os.Getenv("MINIO_PUBLIC_ENDPOINT")
	if public_endpoint == "" {
		return nil, fmt.Errorf("MINIO_PUBLIC_ENDPOINT is not set")
	}

	endpoint := os.Getenv("MINIO_ENDPOINT")
	if endpoint == "" {
		return nil, fmt.Errorf("MINIO_ENDPOINT is not set")
	}

	accessKeyID := os.Getenv("MINIO_ACCESS_KEY")
	if accessKeyID == "" {
		return nil, fmt.Errorf("MINIO_ACCESS_KEY is not set")
	}

	secretAccessKey := os.Getenv("MINIO_SECRET_KEY")
	if secretAccessKey == "" {
		return nil, fmt.Errorf("MINIO_SECRET_KEY is not set")
	}
	useSSL := false

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}

	// Create a second client with the public endpoint for generating presigned URLs
	// Remove protocol prefix from public_endpoint if present
	publicEndpointClean := strings.TrimPrefix(strings.TrimPrefix(public_endpoint, "https://"), "http://")
	publicUseSSL := strings.HasPrefix(public_endpoint, "https://")

	publicMinioClient, err := minio.New(publicEndpointClean, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: publicUseSSL,
	})
	if err != nil {
		return nil, err
	}

	service := &FileService{
		PublicEndpoint: public_endpoint,
		Client:         minioClient,
		PublicClient:   publicMinioClient,
		DB:             db,
		Queries:        gen.New(db),
	}

	ctx := context.Background()

	// Create avatars bucket with public read access
	avatarsBucket := "avatars"
	exists, err := service.Client.BucketExists(ctx, avatarsBucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check avatars bucket existence: %w", err)
	}

	if !exists {
		err = service.Client.MakeBucket(ctx, avatarsBucket, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create avatars bucket: %w", err)
		}

		policy := fmt.Sprintf(`{
			"Version": "2012-10-17",
			"Statement": [
				{
					"Effect": "Allow",
					"Principal": "*",
					"Action": "s3:GetObject",
					"Resource": "arn:aws:s3:::%s/*"
				}
			]
		}`, avatarsBucket)

		err = service.Client.SetBucketPolicy(ctx, avatarsBucket, policy)
		if err != nil {
			return nil, fmt.Errorf("failed to set avatars bucket policy: %w", err)
		}
	}

	// Create materials bucket WITHOUT public read access (access via signed URLs)
	materialsBucket := "materials"
	exists, err = service.Client.BucketExists(ctx, materialsBucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check materials bucket existence: %w", err)
	}

	if !exists {
		err = service.Client.MakeBucket(ctx, materialsBucket, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create materials bucket: %w", err)
		}
		// No public policy - access controlled via signed URLs
	}

	// Create photos bucket with public read access (extra profile photos)
	photosBucket := "photos"
	exists, err = service.Client.BucketExists(ctx, photosBucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check photos bucket existence: %w", err)
	}

	if !exists {
		err = service.Client.MakeBucket(ctx, photosBucket, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create photos bucket: %w", err)
		}

		policy := fmt.Sprintf(`{
			"Version": "2012-10-17",
			"Statement": [
				{
					"Effect": "Allow",
					"Principal": "*",
					"Action": "s3:GetObject",
					"Resource": "arn:aws:s3:::%s/*"
				}
			]
		}`, photosBucket)

		err = service.Client.SetBucketPolicy(ctx, photosBucket, policy)
		if err != nil {
			return nil, fmt.Errorf("failed to set photos bucket policy: %w", err)
		}
	}

	return service, nil
}

func (s *FileService) UploadFile(ctx context.Context, file File) error {
	options := minio.PutObjectOptions{
		ContentType: file.ContentType,
	}

	_, err := s.Client.PutObject(ctx, file.Bucket, file.Key, file.Reader, file.Size, options)
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	return nil
}

func (s *FileService) DeleteFile(ctx context.Context, bucket, key string) error {
	err := s.Client.RemoveObject(ctx, bucket, key, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

func (s *FileService) ExtractKeyFromPath(filePath string) string {
	// Extract the key from a full path like "http://minio:9000/avatars/uuid_timestamp.jpg"
	// Returns "uuid_timestamp.jpg"
	parts := strings.Split(filePath, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

func IsValidImageType(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp":
		return true
	default:
		return false
	}
}

func GetContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		switch ext {
		case ".webp":
			return "image/webp"
		default:
			return "application/octet-stream"
		}
	}
	return contentType
}

func ValidateFileSize(size int64, maxSizeMB int64) error {
	maxSizeBytes := maxSizeMB * 1024 * 1024
	if size > maxSizeBytes {
		return fmt.Errorf("file size %d bytes exceeds maximum allowed size of %d MB", size, maxSizeMB)
	}
	return nil
}

func (s *FileService) GetPresignedURL(ctx context.Context, bucket, key string, expiration time.Duration) (string, error) {
	// Use the public client to generate URLs with the correct public endpoint
	url, err := s.PublicClient.PresignedGetObject(ctx, bucket, key, expiration, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}
	return url.String(), nil
}

package storage

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
)

// S3Storage implements StorageService for AWS S3
type S3Storage struct {
	client       *s3.Client
	uploader     *manager.Uploader
	bucket       string
	region       string
	prefix       string
	cdnURL       string
}

// S3Config holds S3 configuration
type S3Config struct {
	Bucket          string
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	Prefix          string
	CDNURL          string
	UseIAMRole      bool
}

// NewS3Storage creates a new S3 storage service
func NewS3Storage(cfg S3Config) (*S3Storage, error) {
	ctx := context.Background()
	
	var awsCfg aws.Config
	var err error
	
	// Check if static credentials are provided and not empty
	hasStaticCredentials := cfg.AccessKeyID != "" && cfg.SecretAccessKey != ""
	
	if cfg.UseIAMRole || !hasStaticCredentials {
		// Use IAM role credentials or default credential chain (includes ~/.aws/credentials)
		awsCfg, err = config.LoadDefaultConfig(ctx, config.WithRegion(cfg.Region))
	} else {
		// Use provided static credentials
		awsCfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(cfg.Region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				cfg.AccessKeyID,
				cfg.SecretAccessKey,
				"",
			)),
		)
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}
	
	client := s3.NewFromConfig(awsCfg)
	uploader := manager.NewUploader(client)
	
	return &S3Storage{
		client:   client,
		uploader: uploader,
		bucket:   cfg.Bucket,
		region:   cfg.Region,
		prefix:   cfg.Prefix,
		cdnURL:   cfg.CDNURL,
	}, nil
}

// Upload uploads a file to S3
func (s *S3Storage) Upload(ctx context.Context, filename string, content io.Reader, metadata map[string]string) (string, error) {
	// Check if filename already contains date path structure (e.g., starts with prefix/YYYY/MM/DD/)
	// If so, use it as-is; otherwise generate S3 key with date structure
	var key string
	if s.isFullPath(filename) {
		// Already a full path (e.g., from transcoding job), use as-is
		key = filename
	} else {
		// Generate S3 key with date structure for new uploads
		key = s.generateS3Key(filename)
	}
	
	// Prepare S3 metadata
	s3Metadata := make(map[string]string)
	for k, v := range metadata {
		s3Metadata[k] = v
	}
	
	// Upload file
	input := &s3.PutObjectInput{
		Bucket:               aws.String(s.bucket),
		Key:                  aws.String(key),
		Body:                 content,
		Metadata:             s3Metadata,
		ServerSideEncryption: "AES256",
	}
	
	if contentType, ok := metadata[MetadataContentType]; ok {
		input.ContentType = aws.String(contentType)
	}
	
	_, err := s.uploader.Upload(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %w", err)
	}
	
	return key, nil
}

// Download retrieves a file from S3
func (s *S3Storage) Download(ctx context.Context, path string) (io.ReadCloser, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	}
	
	result, err := s.client.GetObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to download from S3: %w", err)
	}
	
	return result.Body, nil
}

// Delete removes a file from S3
func (s *S3Storage) Delete(ctx context.Context, path string) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	}
	
	_, err := s.client.DeleteObject(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete from S3: %w", err)
	}
	
	return nil
}

// Exists checks if a file exists in S3
func (s *S3Storage) Exists(ctx context.Context, path string) (bool, error) {
	input := &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	}
	
	_, err := s.client.HeadObject(ctx, input)
	if err != nil {
		if strings.Contains(err.Error(), "NotFound") {
			return false, nil
		}
		return false, err
	}
	
	return true, nil
}

// GetURL returns a URL for accessing the file
func (s *S3Storage) GetURL(ctx context.Context, path string) (string, error) {
	// If CDN URL is configured, use it
	if s.cdnURL != "" {
		return fmt.Sprintf("%s/%s", strings.TrimSuffix(s.cdnURL, "/"), path), nil
	}
	
	// Otherwise, return S3 URL
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.bucket, s.region, path), nil
}

// generateS3Key generates S3 object key with prefix
func (s *S3Storage) generateS3Key(filename string) string {
	// Use timestamp-based path structure
	now := time.Now()
	key := fmt.Sprintf("%s/%d/%02d/%02d/%s",
		strings.TrimSuffix(s.prefix, "/"),
		now.Year(),
		now.Month(),
		now.Day(),
		filename,
	)
	return strings.TrimPrefix(key, "/")
}

// isFullPath checks if the given path already contains date structure (YYYY/MM/DD)
func (s *S3Storage) isFullPath(path string) bool {
	// Check if path matches pattern: prefix/YYYY/MM/DD/...
	// or just YYYY/MM/DD/... (4 digits / 2 digits / 2 digits)
	parts := strings.Split(path, "/")
	
	// Need at least 4 parts: prefix, year, month, day, filename
	if len(parts) < 4 {
		return false
	}
	
	// Check if any consecutive 3 parts match YYYY/MM/DD pattern
	for i := 0; i < len(parts)-2; i++ {
		year := parts[i]
		month := parts[i+1]
		day := parts[i+2]
		
		// Check if year is 4 digits, month and day are 2 digits
		if len(year) == 4 && len(month) == 2 && len(day) == 2 {
			// Try to parse as numbers
			if _, err := strconv.Atoi(year); err == nil {
				if _, err := strconv.Atoi(month); err == nil {
					if _, err := strconv.Atoi(day); err == nil {
						return true
					}
				}
			}
		}
	}
	
	return false
}

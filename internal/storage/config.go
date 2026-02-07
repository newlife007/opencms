package storage

import (
	"fmt"
	"os"
)

// Config holds storage configuration
type Config struct {
	Type            string // "local" or "s3"
	LocalPath       string
	S3Bucket        string
	S3Region        string
	S3AccessKey     string
	S3SecretKey     string
	S3Prefix        string
	S3CDNURL        string
	S3UseIAMRole    bool
}

// NewStorageFromConfig creates a storage service based on configuration
func NewStorageFromConfig(cfg Config) (StorageService, error) {
	switch cfg.Type {
	case "local":
		return NewLocalStorage(cfg.LocalPath)
	case "s3":
		s3cfg := S3Config{
			Bucket:          cfg.S3Bucket,
			Region:          cfg.S3Region,
			AccessKeyID:     cfg.S3AccessKey,
			SecretAccessKey: cfg.S3SecretKey,
			Prefix:          cfg.S3Prefix,
			CDNURL:          cfg.S3CDNURL,
			UseIAMRole:      cfg.S3UseIAMRole,
		}
		return NewS3Storage(s3cfg)
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", cfg.Type)
	}
}

// LoadConfigFromEnv loads storage configuration from environment variables
func LoadConfigFromEnv() Config {
	storageType := getEnv("STORAGE_TYPE", "local")
	
	return Config{
		Type:         storageType,
		LocalPath:    getEnv("LOCAL_STORAGE_PATH", "./storage"),
		S3Bucket:     getEnv("S3_BUCKET", ""),
		S3Region:     getEnv("S3_REGION", "us-east-1"),
		S3AccessKey:  getEnv("S3_ACCESS_KEY_ID", ""),
		S3SecretKey:  getEnv("S3_SECRET_ACCESS_KEY", ""),
		S3Prefix:     getEnv("S3_PREFIX", "openwan/"),
		S3CDNURL:     getEnv("S3_CDN_URL", ""),
		S3UseIAMRole: getEnv("S3_USE_IAM_ROLE", "false") == "true",
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

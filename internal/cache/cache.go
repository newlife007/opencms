package cache

import (
	"context"
	"time"
)

// CacheService defines the caching interface
type CacheService interface {
	Get(ctx context.Context, key string) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	Clear(ctx context.Context) error
}

// CacheKeyType defines cache key types
type CacheKeyType string

const (
	KeyTypeUserPermissions CacheKeyType = "user:permissions"
	KeyTypeCategoryTree    CacheKeyType = "category:tree"
	KeyTypeCatalogConfig   CacheKeyType = "catalog:config"
	KeyTypeFileMetadata    CacheKeyType = "file:metadata"
)

// TTL values for different cache types
const (
	TTLPermissions = 15 * time.Minute
	TTLCategories  = 1 * time.Hour
	TTLCatalog     = 1 * time.Hour
	TTLFileMetadata = 30 * time.Minute
)

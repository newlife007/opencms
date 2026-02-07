package cache

import "fmt"

// GenerateKey generates a cache key based on type and parameters
func GenerateKey(keyType CacheKeyType, params ...interface{}) string {
	switch keyType {
	case KeyTypeUserPermissions:
		if len(params) > 0 {
			return fmt.Sprintf("user:permissions:%v", params[0])
		}
	case KeyTypeCategoryTree:
		return "category:tree"
	case KeyTypeCatalogConfig:
		if len(params) > 0 {
			return fmt.Sprintf("catalog:config:%v", params[0])
		}
	case KeyTypeFileMetadata:
		if len(params) > 0 {
			return fmt.Sprintf("file:metadata:%v", params[0])
		}
	}
	return string(keyType)
}

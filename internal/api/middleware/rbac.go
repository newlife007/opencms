package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/openwan/media-asset-management/internal/service"
)

var aclService *service.ACLService

// SetACLService sets the ACL service for RBAC middleware
func SetACLService(acl *service.ACLService) {
	aclService = acl
}

// RequireAdmin middleware requires admin role
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if user is authenticated
		_, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Authentication required",
			})
			c.Abort()
			return
		}

		// Check if user is admin (from session)
		isAdmin, exists := c.Get("is_admin")
		if !exists || !isAdmin.(bool) {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "Admin access required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequirePermission middleware requires specific permission
// Permission format can be:
//   1. "namespace.controller.action" (e.g., "files.browse.list")
//   2. Simplified format (e.g., "file.delete") - will be mapped
func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if user is authenticated
		userIDInterface, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Authentication required",
			})
			c.Abort()
			return
		}

		userID := int(userIDInterface.(uint))

		// Admin bypass - admins have all permissions
		isAdmin, _ := c.Get("is_admin")
		if isAdmin != nil && isAdmin.(bool) {
			c.Next()
			return
		}

		// If ACL service is not set, allow for backward compatibility
		if aclService == nil {
			c.Next()
			return
		}

		// Parse permission string into namespace.controller.action
		parts := strings.Split(permission, ".")
		if len(parts) != 3 {
			// Try to map simplified format to full format
			namespace, controller, action := mapSimplifiedPermission(permission)
			if namespace == "" {
				// Invalid permission format, deny access
				c.JSON(http.StatusForbidden, gin.H{
					"success": false,
					"message": "Invalid permission format",
					"error":   "Permission must be in format: namespace.controller.action",
				})
				c.Abort()
				return
			}
			parts = []string{namespace, controller, action}
		}

		namespace := parts[0]
		controller := parts[1]
		action := parts[2]

		// Check permission using ACL service
		hasPermission, err := aclService.HasPermission(c.Request.Context(), userID, namespace, controller, action)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to check permission",
				"error":   err.Error(),
			})
			c.Abort()
			return
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "Permission denied",
				"error":   "You don't have permission: " + permission,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// mapSimplifiedPermission maps simplified permission format to namespace.controller.action
// Examples:
//   "file.delete" -> "files.edit.delete"
//   "file.upload" -> "files.upload.create"
//   "category.create" -> "categories.manage.create"
func mapSimplifiedPermission(simplified string) (namespace, controller, action string) {
	parts := strings.Split(simplified, ".")
	if len(parts) != 2 {
		return "", "", ""
	}

	resource := parts[0]
	operation := parts[1]

	// Map simplified formats
	switch resource {
	case "file":
		namespace = "files"
		switch operation {
		case "list", "view", "search", "download", "preview":
			controller = "browse"
		case "upload":
			controller = "upload"
			action = "create" // Map upload to create action
			return
		case "create", "batch":
			controller = "upload"
		case "edit", "update", "delete", "restore":
			controller = "edit"
		case "catalog":
			controller = "catalog"
			action = "edit"
			return
		case "publish", "approve", "reject", "unpublish":
			controller = "publish"
		default:
			return "", "", ""
		}
		action = operation
	case "category":
		namespace = "categories"
		controller = "manage"
		action = operation
	case "catalog":
		namespace = "catalog"
		controller = "config"
		action = operation
	case "user":
		namespace = "users"
		controller = "manage"
		action = operation
	case "group":
		namespace = "groups"
		controller = "manage"
		action = operation
	case "role":
		namespace = "roles"
		controller = "manage"
		action = operation
	case "permission":
		namespace = "permissions"
		controller = "manage"
		action = operation
	case "admin":
		// Generic admin permission - map to appropriate namespace
		namespace = "system"
		controller = "admin"
		action = operation
	default:
		return "", "", ""
	}

	return namespace, controller, action
}

// RequireRole middleware requires specific role
func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if user is authenticated
		_, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Authentication required",
			})
			c.Abort()
			return
		}

		// For now, allow if authenticated
		// TODO: Implement proper role checking with database
		c.Next()
	}
}

// CheckCategoryAccess middleware checks category access
func CheckCategoryAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if user is authenticated
		_, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Authentication required",
			})
			c.Abort()
			return
		}

		// For now, allow if authenticated
		// TODO: Implement proper category access checking with database
		c.Next()
	}
}

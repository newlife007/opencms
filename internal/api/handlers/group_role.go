package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/openwan/media-asset-management/internal/models"
	"github.com/openwan/media-asset-management/internal/service"
)

// GroupHandler handles group operations
type GroupHandler struct {
	groupService *service.GroupService
}

// NewGroupHandler creates a new group handler
func NewGroupHandler(groupService *service.GroupService) *GroupHandler {
	return &GroupHandler{
		groupService: groupService,
	}
}

// ListGroups returns paginated list of groups
func (h *GroupHandler) ListGroups() gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

		if page < 1 {
			page = 1
		}
		if pageSize < 1 || pageSize > 100 {
			pageSize = 20
		}

		offset := (page - 1) * pageSize
		groups, total, err := h.groupService.ListGroups(c.Request.Context(), pageSize, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to retrieve groups",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    groups,
			"pagination": gin.H{
				"page":        page,
				"page_size":   pageSize,
				"total":       total,
				"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
			},
		})
	}
}

// GetGroup returns group details by ID
func (h *GroupHandler) GetGroup() gin.HandlerFunc {
	return func(c *gin.Context) {
		groupID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid group ID",
			})
			return
		}

		group, err := h.groupService.GetGroupByID(c.Request.Context(), uint(groupID))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "Group not found",
			})
			return
		}

		// Get group categories and roles
		categories, _ := h.groupService.GetGroupCategories(c.Request.Context(), uint(groupID))
		roles, _ := h.groupService.GetGroupRoles(c.Request.Context(), uint(groupID))

		c.JSON(http.StatusOK, gin.H{
			"success":    true,
			"data":       group,
			"categories": categories,
			"roles":      roles,
		})
	}
}

// CreateGroup creates a new group
func (h *GroupHandler) CreateGroup() gin.HandlerFunc {
	return func(c *gin.Context) {
		type CreateGroupRequest struct {
			Name        string `json:"name" binding:"required"`
			Description string `json:"description"`
			Enabled     bool   `json:"enabled"`
		}

		var req CreateGroupRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid request body",
				"error":   err.Error(),
			})
			return
		}

		group := &models.Groups{
			Name:        req.Name,
			Description: req.Description,
			Enabled:     req.Enabled,
		}

		if err := h.groupService.CreateGroup(c.Request.Context(), group); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to create group",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"success": true,
			"message": "Group created successfully",
			"data":    group,
		})
	}
}

// UpdateGroup updates group information
func (h *GroupHandler) UpdateGroup() gin.HandlerFunc {
	return func(c *gin.Context) {
		groupID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid group ID",
			})
			return
		}

		var updates map[string]interface{}
		if err := c.ShouldBindJSON(&updates); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid request body",
				"error":   err.Error(),
			})
			return
		}

		if err := h.groupService.UpdateGroup(c.Request.Context(), uint(groupID), updates); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to update group",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Group updated successfully",
		})
	}
}

// DeleteGroup deletes a group
func (h *GroupHandler) DeleteGroup() gin.HandlerFunc {
	return func(c *gin.Context) {
		groupID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid group ID",
			})
			return
		}

		if err := h.groupService.DeleteGroup(c.Request.Context(), uint(groupID)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to delete group",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Group deleted successfully",
		})
	}
}

// AssignCategories assigns categories to a group
func (h *GroupHandler) AssignCategories() gin.HandlerFunc {
	return func(c *gin.Context) {
		groupID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid group ID",
			})
			return
		}

		type AssignRequest struct {
			CategoryIDs []uint `json:"category_ids" binding:"required"`
		}

		var req AssignRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid request body",
				"error":   err.Error(),
			})
			return
		}

		if err := h.groupService.AssignCategories(c.Request.Context(), uint(groupID), req.CategoryIDs); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to assign categories",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Categories assigned successfully",
		})
	}
}

// AssignRoles assigns roles to a group
func (h *GroupHandler) AssignRoles() gin.HandlerFunc {
	return func(c *gin.Context) {
		groupID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid group ID",
			})
			return
		}

		type AssignRequest struct {
			RoleIDs []uint `json:"role_ids" binding:"required"`
		}

		var req AssignRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid request body",
				"error":   err.Error(),
			})
			return
		}

		if err := h.groupService.AssignRoles(c.Request.Context(), uint(groupID), req.RoleIDs); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to assign roles",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Roles assigned successfully",
		})
	}
}

// RoleHandler handles role operations
type RoleHandler struct {
	roleService *service.RoleService
}

// NewRoleHandler creates a new role handler
func NewRoleHandler(roleService *service.RoleService) *RoleHandler {
	return &RoleHandler{
		roleService: roleService,
	}
}

// ListRoles returns paginated list of roles
func (h *RoleHandler) ListRoles() gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

		if page < 1 {
			page = 1
		}
		if pageSize < 1 || pageSize > 100 {
			pageSize = 20
		}

		offset := (page - 1) * pageSize
		roles, total, err := h.roleService.ListRoles(c.Request.Context(), pageSize, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to retrieve roles",
				"error":   err.Error(),
			})
			return
		}

		// Enhance roles with permission count
		rolesWithCount := make([]gin.H, len(roles))
		for i, role := range roles {
			permissions, _ := h.roleService.GetRolePermissions(c.Request.Context(), uint(role.ID))
			rolesWithCount[i] = gin.H{
				"id":               role.ID,
				"name":             role.Name,
				"description":      role.Description,
				"weight":           role.Weight,
				"enabled":          role.Enabled,
				"is_system": role.IsSystem,
				"permission_count": len(permissions),
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    rolesWithCount,
			"pagination": gin.H{
				"page":        page,
				"page_size":   pageSize,
				"total":       total,
				"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
			},
		})
	}
}

// GetRole returns role details by ID
func (h *RoleHandler) GetRole() gin.HandlerFunc {
	return func(c *gin.Context) {
		roleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid role ID",
			})
			return
		}

		role, err := h.roleService.GetRoleByID(c.Request.Context(), uint(roleID))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "Role not found",
			})
			return
		}

		// Get role permissions
		permissions, _ := h.roleService.GetRolePermissions(c.Request.Context(), uint(roleID))

		c.JSON(http.StatusOK, gin.H{
			"success":     true,
			"data":        role,
			"permissions": permissions,
		})
	}
}

// CreateRole creates a new role
func (h *RoleHandler) CreateRole() gin.HandlerFunc {
	return func(c *gin.Context) {
		var role models.Roles
		if err := c.ShouldBindJSON(&role); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid request body",
				"error":   err.Error(),
			})
			return
		}

		if err := h.roleService.CreateRole(c.Request.Context(), &role); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to create role",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"success": true,
			"message": "Role created successfully",
			"data":    role,
		})
	}
}

// UpdateRole updates role information
func (h *RoleHandler) UpdateRole() gin.HandlerFunc {
	return func(c *gin.Context) {
		roleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid role ID",
			})
			return
		}

		var updates map[string]interface{}
		if err := c.ShouldBindJSON(&updates); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid request body",
				"error":   err.Error(),
			})
			return
		}

		if err := h.roleService.UpdateRole(c.Request.Context(), uint(roleID), updates); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to update role",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Role updated successfully",
		})
	}
}

// DeleteRole deletes a role
func (h *RoleHandler) DeleteRole() gin.HandlerFunc {
	return func(c *gin.Context) {
		roleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid role ID",
			})
			return
		}

		// Check if role is a system role
		role, err := h.roleService.GetRoleByID(c.Request.Context(), uint(roleID))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "Role not found",
			})
			return
		}

		// Prevent deletion of system roles
		if role.IsSystem {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "Cannot delete system role",
				"error":   "System roles (超级管理员、内容管理员、审核员、编辑、查看者) cannot be deleted",
			})
			return
		}

		if err := h.roleService.DeleteRole(c.Request.Context(), uint(roleID)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to delete role",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Role deleted successfully",
		})
	}
}

// AssignPermissions assigns permissions to a role
func (h *RoleHandler) AssignPermissions() gin.HandlerFunc {
	return func(c *gin.Context) {
		roleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid role ID",
			})
			return
		}

		type AssignRequest struct {
			PermissionIDs []uint `json:"permission_ids" binding:"required"`
		}

		var req AssignRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid request body",
				"error":   err.Error(),
			})
			return
		}

		if err := h.roleService.AssignPermissions(c.Request.Context(), uint(roleID), req.PermissionIDs); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to assign permissions",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Permissions assigned successfully",
		})
	}
}

// PermissionHandler handles permission operations
type PermissionHandler struct {
	permissionService *service.PermissionService
}

// NewPermissionHandler creates a new permission handler
func NewPermissionHandler(permissionService *service.PermissionService) *PermissionHandler {
	return &PermissionHandler{
		permissionService: permissionService,
	}
}

// ListPermissions returns all permissions
func (h *PermissionHandler) ListPermissions() gin.HandlerFunc {
	return func(c *gin.Context) {
		permissions, err := h.permissionService.ListPermissions(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to retrieve permissions",
				"error":   err.Error(),
			})
			return
		}

		// Transform permissions to include frontend-expected fields
		transformedPermissions := make([]gin.H, len(permissions))
		for i, perm := range permissions {
			// Generate permission name from namespace.controller.action
			permName := perm.Namespace + "." + perm.Controller + "." + perm.Action
			
			transformedPermissions[i] = gin.H{
				"id":          perm.ID,
				"name":        permName,                    // Frontend expects: permission name
				"description": perm.Aliasname,              // Frontend expects: description (Chinese)
				"module":      perm.Namespace,              // Frontend expects: module name
				"namespace":   perm.Namespace,              // Keep original for reference
				"controller":  perm.Controller,             // Keep original for reference
				"action":      perm.Action,                 // Keep original for reference
				"aliasname":   perm.Aliasname,              // Keep original for reference
				"rbac":        perm.RBAC,                   // Keep RBAC level
				"created_at":  nil,                         // No timestamp in current schema
				"updated_at":  nil,                         // No timestamp in current schema
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    transformedPermissions,
		})
	}
}

// GetPermission returns permission details by ID
func (h *PermissionHandler) GetPermission() gin.HandlerFunc {
	return func(c *gin.Context) {
		permissionID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid permission ID",
			})
			return
		}

		permission, err := h.permissionService.GetPermissionByID(c.Request.Context(), uint(permissionID))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "Permission not found",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    permission,
		})
	}
}

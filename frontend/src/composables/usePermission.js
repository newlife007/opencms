import { computed } from 'vue'
import { useAuthStore } from '@/stores/auth'

/**
 * Permission checking composable
 * Provides helper functions for RBAC permission checks
 */
export function usePermission() {
  const authStore = useAuthStore()

  /**
   * Check if user has a specific permission
   * @param {string} permission - Permission code to check
   * @returns {boolean}
   */
  const hasPermission = (permission) => {
    return authStore.hasPermission(permission)
  }

  /**
   * Check if user has a specific role
   * @param {string} role - Role name or code to check
   * @returns {boolean}
   */
  const hasRole = (role) => {
    return authStore.hasRole(role)
  }

  /**
   * Check if user has any of the specified permissions
   * @param {Array<string>} permissions - Array of permission codes
   * @returns {boolean}
   */
  const hasAnyPermission = (permissions) => {
    return authStore.hasAnyPermission(permissions)
  }

  /**
   * Check if user has all of the specified permissions
   * @param {Array<string>} permissions - Array of permission codes
   * @returns {boolean}
   */
  const hasAllPermissions = (permissions) => {
    return authStore.hasAllPermissions(permissions)
  }

  /**
   * Check if user is admin
   * @returns {boolean}
   */
  const isAdmin = computed(() => {
    return authStore.isAdmin
  })

  /**
   * Check if user is authenticated
   * @returns {boolean}
   */
  const isAuthenticated = computed(() => {
    return authStore.isAuthenticated
  })

  /**
   * Get current user
   * @returns {Object|null}
   */
  const currentUser = computed(() => {
    return authStore.user
  })

  /**
   * Get user display name
   * @returns {string}
   */
  const userDisplayName = computed(() => {
    return authStore.userDisplayName
  })

  return {
    hasPermission,
    hasRole,
    hasAnyPermission,
    hasAllPermissions,
    isAdmin,
    isAuthenticated,
    currentUser,
    userDisplayName,
  }
}

/**
 * Permission directive for v-permission
 * Usage: v-permission="'admin.users.edit'" or v-permission="['admin.users.edit', 'admin.users.view']"
 */
export const permissionDirective = {
  mounted(el, binding) {
    const { value } = binding
    const authStore = useAuthStore()

    if (value) {
      let hasPermission = false

      if (Array.isArray(value)) {
        // Check if user has any of the permissions
        hasPermission = authStore.hasAnyPermission(value)
      } else {
        // Check single permission
        hasPermission = authStore.hasPermission(value)
      }

      if (!hasPermission) {
        el.parentNode && el.parentNode.removeChild(el)
      }
    }
  },
}

/**
 * Role directive for v-role
 * Usage: v-role="'admin'" or v-role="['admin', 'editor']"
 */
export const roleDirective = {
  mounted(el, binding) {
    const { value } = binding
    const authStore = useAuthStore()

    if (value) {
      let hasRole = false

      if (Array.isArray(value)) {
        // Check if user has any of the roles
        hasRole = value.some(role => authStore.hasRole(role))
      } else {
        // Check single role
        hasRole = authStore.hasRole(value)
      }

      if (!hasRole) {
        el.parentNode && el.parentNode.removeChild(el)
      }
    }
  },
}

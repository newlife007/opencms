import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import * as authAPI from '@/api/auth'

export const useAuthStore = defineStore('auth', () => {
  // State
  const token = ref(localStorage.getItem('token') || '')
  const user = ref(null)
  const isAuthenticated = ref(false)
  const permissions = ref([])
  const roles = ref([])

  // Getters
  const userDisplayName = computed(() => {
    return user.value?.nickname || user.value?.username || '未知用户'
  })

  const userLevel = computed(() => {
    return user.value?.level_id || 0
  })

  const userGroupId = computed(() => {
    return user.value?.group_id || null
  })

  const hasPermission = computed(() => {
    return (permission) => {
      if (!permission) return true
      if (user.value?.is_admin) return true
      return permissions.value.includes(permission)
    }
  })

  const hasRole = computed(() => {
    return (role) => {
      if (!role) return true
      if (user.value?.is_admin) return true
      return roles.value.some(r => r.name === role || r.code === role)
    }
  })

  const hasAnyPermission = computed(() => {
    return (perms) => {
      if (!perms || perms.length === 0) return true
      if (user.value?.is_admin) return true
      return perms.some(p => permissions.value.includes(p))
    }
  })

  const hasAllPermissions = computed(() => {
    return (perms) => {
      if (!perms || perms.length === 0) return true
      if (user.value?.is_admin) return true
      return perms.every(p => permissions.value.includes(p))
    }
  })

  const isAdmin = computed(() => {
    return user.value?.is_admin || false
  })

  // Actions
  const login = async (credentials) => {
    try {
      const response = await authAPI.login(credentials)
      
      if (response.success) {
        // Store token
        token.value = response.data.token || ''
        if (token.value) {
          localStorage.setItem('token', token.value)
        }

        // Store user data
        user.value = response.data.user
        isAuthenticated.value = true

        // Fetch full user info with permissions
        await fetchCurrentUser()

        return true
      }
      return false
    } catch (error) {
      console.error('Login failed:', error)
      return false
    }
  }

  const logout = async () => {
    try {
      await authAPI.logout()
    } catch (error) {
      console.error('Logout API call failed:', error)
    } finally {
      // Clear local state regardless of API result
      clearAuth()
    }
  }

  const clearAuth = () => {
    token.value = ''
    user.value = null
    isAuthenticated.value = false
    permissions.value = []
    roles.value = []
    localStorage.removeItem('token')
  }

  const fetchCurrentUser = async () => {
    try {
      const response = await authAPI.getCurrentUser()
      if (response.success) {
        user.value = response.data.user
        permissions.value = response.data.permissions || []
        roles.value = response.data.roles || []
        isAuthenticated.value = true
        return true
      }
      return false
    } catch (error) {
      console.error('Fetch current user failed:', error)
      clearAuth()
      return false
    }
  }

  const refreshToken = async () => {
    try {
      const response = await authAPI.refreshToken()
      if (response.success && response.data.token) {
        token.value = response.data.token
        localStorage.setItem('token', token.value)
        return true
      }
      return false
    } catch (error) {
      console.error('Token refresh failed:', error)
      return false
    }
  }

  const updateProfile = async (profileData) => {
    try {
      const response = await authAPI.updateProfile(profileData)
      if (response.success) {
        user.value = { ...user.value, ...response.data.user }
        return true
      }
      return false
    } catch (error) {
      console.error('Update profile failed:', error)
      throw error
    }
  }

  const changePassword = async (oldPassword, newPassword) => {
    try {
      const response = await authAPI.changePassword({
        old_password: oldPassword,
        new_password: newPassword,
      })
      return response.success
    } catch (error) {
      console.error('Change password failed:', error)
      throw error
    }
  }

  const initialize = async () => {
    const storedToken = localStorage.getItem('token')
    if (!storedToken) {
      return false
    }

    token.value = storedToken
    
    // Validate token by fetching current user
    try {
      return await fetchCurrentUser()
    } catch (error) {
      clearAuth()
      return false
    }
  }

  // Auto-refresh token before expiration
  const setupTokenRefresh = () => {
    // Check token expiration and refresh if needed
    // This should be called on app initialization
    if (!token.value) return

    // Decode JWT to get expiration (simplified - in production use a JWT library)
    try {
      const payload = JSON.parse(atob(token.value.split('.')[1]))
      const exp = payload.exp * 1000 // Convert to milliseconds
      const now = Date.now()
      const timeUntilExpiry = exp - now

      // Refresh 5 minutes before expiration
      if (timeUntilExpiry > 0 && timeUntilExpiry < 5 * 60 * 1000) {
        refreshToken()
      } else if (timeUntilExpiry > 0) {
        // Set timeout to refresh before expiration
        setTimeout(() => {
          refreshToken()
        }, timeUntilExpiry - 5 * 60 * 1000)
      }
    } catch (error) {
      console.error('Error parsing token:', error)
    }
  }

  return {
    // State
    token,
    user,
    isAuthenticated,
    permissions,
    roles,

    // Getters
    userDisplayName,
    userLevel,
    userGroupId,
    hasPermission,
    hasRole,
    hasAnyPermission,
    hasAllPermissions,
    isAdmin,

    // Actions
    login,
    logout,
    clearAuth,
    fetchCurrentUser,
    refreshToken,
    updateProfile,
    changePassword,
    initialize,
    setupTokenRefresh,
  }
})

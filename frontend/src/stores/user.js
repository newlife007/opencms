import { defineStore } from 'pinia'
import { ref } from 'vue'
import authApi from '@/api/auth'

export const useUserStore = defineStore('user', () => {
  const user = ref(null)
  const token = ref(localStorage.getItem('token') || '')
  const permissions = ref([])
  const roles = ref([])

  // Login
  async function login(credentials) {
    try {
      const res = await authApi.login(credentials)
      // Backend returns data directly at root level, not nested in res.data
      if (res.success) {
        user.value = res.user
        permissions.value = res.permissions || res.user?.permissions || []
        roles.value = res.user?.roles || res.roles || []  // Fix: Get roles from res.user.roles
        
        // Store token from backend response
        token.value = res.token || ''
        
        // Store token in localStorage
        if (token.value) {
          localStorage.setItem('token', token.value)
        }
        
        console.log('Login successful:', { user: user.value, roles: roles.value, token: token.value })
        return true
      }
      // If success is false, return false but don't throw
      console.warn('Login failed: success is false')
      return false
    } catch (error) {
      console.error('Login failed:', error)
      // Re-throw the error so Login.vue can catch it and show the message
      throw error
    }
  }

  // Logout
  async function logout() {
    try {
      await authApi.logout()
    } catch (error) {
      console.error('Logout error:', error)
    } finally {
      user.value = null
      token.value = ''
      permissions.value = []
      roles.value = []
      localStorage.removeItem('token')
    }
  }

  // Get current user info
  async function getUserInfo() {
    try {
      const res = await authApi.getCurrentUser()
      // Backend returns data directly at root level
      if (res.success) {
        user.value = res.user
        permissions.value = res.permissions || res.user?.permissions || []
        roles.value = res.user?.roles || res.roles || []  // Fix: Get roles from res.user.roles
        console.log('User info loaded:', { user: user.value, roles: roles.value })
        return true
      }
      return false
    } catch (error) {
      console.error('Get user info failed:', error)
      return false
    }
  }

  // Check permission
  function hasPermission(permission) {
    if (!permission) return true
    
    // Check if user has admin role (admin has all permissions)
    const lowerRoles = roles.value.map(r => r.toLowerCase())
    if (lowerRoles.includes('admin') || lowerRoles.includes('system') || 
        lowerRoles.includes('超级管理员') || roles.value.includes('超级管理员')) {
      return true
    }
    
    // Check permissions array
    // Backend returns permissions as strings: "files.browse.list"
    // Or as objects: {namespace: "files", controller: "browse", action: "list"}
    return permissions.value.some(p => {
      if (typeof p === 'string') {
        // String format: "files.browse.list"
        return p === permission
      } else if (typeof p === 'object' && p.namespace) {
        // Object format: {namespace, controller, action}
        return `${p.namespace}.${p.controller}.${p.action}` === permission
      }
      return false
    })
  }

  // Check role (case-insensitive)
  function hasRole(role) {
    const lowerRoles = roles.value.map(r => r.toLowerCase())
    return lowerRoles.includes(role.toLowerCase())
  }

  // Check if user is admin (case-insensitive check)
  function isAdmin() {
    console.log('Checking isAdmin, roles:', roles.value)
    // Check if user has admin roles (support both English and Chinese role names)
    const lowerRoles = roles.value.map(r => r.toLowerCase())
    const isAdminUser = lowerRoles.includes('admin') || 
           lowerRoles.includes('system') ||
           lowerRoles.includes('超级管理员') ||
           lowerRoles.includes('administrator') ||
           roles.value.includes('超级管理员') // Check original case-sensitive name
    console.log('isAdmin result:', isAdminUser)
    return isAdminUser
  }

  return {
    user,
    token,
    permissions,
    roles,
    login,
    logout,
    getUserInfo,
    hasPermission,
    hasRole,
    isAdmin,
  }
})

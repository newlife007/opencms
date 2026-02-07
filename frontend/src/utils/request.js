import axios from 'axios'
import { ElMessage } from 'element-plus'

// Create axios instance
const request = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api',
  timeout: 0, // No timeout for file uploads and long-running requests
  withCredentials: true, // Send cookies for session-based auth
})

// Request interceptor
request.interceptors.request.use(
  (config) => {
    // Add auth token if exists
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    console.error('Request error:', error)
    return Promise.reject(error)
  }
)

// Response interceptor
request.interceptors.response.use(
  (response) => {
    const res = response.data
    
    // If the custom code is not 200, it is judged as an error
    if (res.success === false) {
      // Don't show error message here, let the caller handle it
      // ElMessage.error(res.message || 'Error occurred')
      
      // 401: Unauthorized - only redirect if not on login page
      if (response.status === 401 && !window.location.pathname.includes('/login')) {
        ElMessage.error('Please login again')
        localStorage.removeItem('token')
        // Use window.location to avoid circular dependency
        window.location.href = '/login'
      }
      
      return Promise.reject(new Error(res.message || 'Error'))
    }
    
    return res
  },
  (error) => {
    console.error('Response error:', error)
    
    if (error.response) {
      const { status, data } = error.response
      const isLoginPage = window.location.pathname.includes('/login')
      
      switch (status) {
        case 401:
          // Only redirect to login if not already on login page
          if (!isLoginPage) {
            ElMessage.error('Authentication failed. Please login again')
            localStorage.removeItem('token')
            // Use window.location to avoid circular dependency
            window.location.href = '/login'
          } else {
            // On login page, show the error message but don't redirect
            ElMessage.error(data?.message || 'Invalid username or password')
          }
          break
        case 403:
          ElMessage.error('Access denied. Insufficient permissions')
          break
        case 404:
          ElMessage.error('Resource not found')
          break
        case 500:
          ElMessage.error(data?.message || 'Server error')
          break
        default:
          ElMessage.error(data?.message || 'Request failed')
      }
    } else if (error.request) {
      ElMessage.error('Network error. Please check your connection')
    } else {
      ElMessage.error('Request failed: ' + error.message)
    }
    
    return Promise.reject(error)
  }
)

export default request

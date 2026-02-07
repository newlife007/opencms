import request from '@/utils/request'

/**
 * 用户登录
 * @param {Object} data - { username, password }
 * @returns {Promise}
 */
export function login(data) {
  return request({
    url: '/auth/login',
    method: 'post',
    data,
  })
}

/**
 * 用户登出
 * @returns {Promise}
 */
export function logout() {
  return request({
    url: '/auth/logout',
    method: 'post',
  })
}

/**
 * 获取当前用户信息
 * @returns {Promise}
 */
export function getCurrentUser() {
  return request({
    url: '/auth/me',
    method: 'get',
  })
}

/**
 * 刷新 Token
 * @returns {Promise}
 */
export function refreshToken() {
  return request({
    url: '/auth/refresh',
    method: 'post',
  })
}

/**
 * 更新用户资料
 * @param {Object} data - { nickname, email, ... }
 * @returns {Promise}
 */
export function updateProfile(data) {
  return request({
    url: '/auth/profile',
    method: 'put',
    data,
  })
}

/**
 * 修改密码
 * @param {Object} data - { old_password, new_password }
 * @returns {Promise}
 */
export function changePassword(data) {
  return request({
    url: '/auth/change-password',
    method: 'post',
    data,
  })
}

/**
 * 忘记密码 - 发送重置邮件
 * @param {Object} data - { email }
 * @returns {Promise}
 */
export function forgotPassword(data) {
  return request({
    url: '/auth/forgot-password',
    method: 'post',
    data,
  })
}

/**
 * 重置密码
 * @param {Object} data - { token, password }
 * @returns {Promise}
 */
export function resetPassword(data) {
  return request({
    url: '/auth/reset-password',
    method: 'post',
    data,
  })
}

export default {
  login,
  logout,
  getCurrentUser,
  refreshToken,
  updateProfile,
  changePassword,
  forgotPassword,
  resetPassword,
}

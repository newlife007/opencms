import request from '@/utils/request'

/**
 * 获取用户列表
 * @param {Object} params - { page, pageSize, username, group_id, level_id, status }
 * @returns {Promise}
 */
export function getUserList(params) {
  return request({
    url: '/admin/users',
    method: 'get',
    params,
  })
}

/**
 * 获取用户详情
 * @param {number} id - 用户ID
 * @returns {Promise}
 */
export function getUserDetail(id) {
  return request({
    url: `/admin/users/${id}`,
    method: 'get',
  })
}

/**
 * 创建用户
 * @param {Object} data - { username, password, email, nickname, group_id, level_id, status }
 * @returns {Promise}
 */
export function createUser(data) {
  return request({
    url: '/admin/users',
    method: 'post',
    data,
  })
}

/**
 * 更新用户
 * @param {number} id - 用户ID
 * @param {Object} data - { email, nickname, group_id, level_id, status }
 * @returns {Promise}
 */
export function updateUser(id, data) {
  return request({
    url: `/admin/users/${id}`,
    method: 'put',
    data,
  })
}

/**
 * 删除用户
 * @param {number} id - 用户ID
 * @returns {Promise}
 */
export function deleteUser(id) {
  return request({
    url: `/admin/users/${id}`,
    method: 'delete',
  })
}

/**
 * 重置用户密码
 * @param {number} id - 用户ID
 * @param {Object} data - { new_password }
 * @returns {Promise}
 */
export function resetUserPassword(id, data) {
  return request({
    url: `/admin/users/${id}/reset-password`,
    method: 'post',
    data,
  })
}

/**
 * 批量删除用户
 * @param {Array<number>} ids - 用户ID数组
 * @returns {Promise}
 */
export function batchDeleteUsers(ids) {
  return request({
    url: '/admin/users/batch-delete',
    method: 'post',
    data: { ids },
  })
}

/**
 * 启用/禁用用户
 * @param {number} id - 用户ID
 * @param {number} status - 状态 (0=禁用, 1=启用)
 * @returns {Promise}
 */
export function updateUserStatus(id, status) {
  return request({
    url: `/admin/users/${id}/status`,
    method: 'put',
    data: { status },
  })
}

/**
 * 获取用户权限
 * @param {number} id - 用户ID
 * @returns {Promise}
 */
export function getUserPermissions(id) {
  return request({
    url: `/admin/users/${id}/permissions`,
    method: 'get',
  })
}

export default {
  getUserList,
  getUserDetail,
  createUser,
  updateUser,
  deleteUser,
  resetUserPassword,
  batchDeleteUsers,
  updateUserStatus,
  getUserPermissions,
}

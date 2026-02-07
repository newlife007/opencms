import request from '@/utils/request'

export default {
  // Get role list
  getList(params) {
    return request({
      url: '/admin/roles',
      method: 'get',
      params,
    })
  },

  // Get role detail
  getDetail(id) {
    return request({
      url: `/admin/roles/${id}`,
      method: 'get',
    })
  },

  // Create role
  create(data) {
    return request({
      url: '/admin/roles',
      method: 'post',
      data,
    })
  },

  // Update role
  update(id, data) {
    return request({
      url: `/admin/roles/${id}`,
      method: 'put',
      data,
    })
  },

  // Delete role
  delete(id) {
    return request({
      url: `/admin/roles/${id}`,
      method: 'delete',
    })
  },

  // Assign permissions to role
  assignPermissions(id, permissionIds) {
    return request({
      url: `/admin/roles/${id}/permissions`,
      method: 'post',
      data: { permission_ids: permissionIds },
    })
  },

  // Get all permissions
  getAllPermissions() {
    return request({
      url: '/admin/permissions',
      method: 'get',
    })
  },
}

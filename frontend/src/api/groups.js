import request from '@/utils/request'

export default {
  // Get group list
  getList(params) {
    return request({
      url: '/admin/groups',
      method: 'get',
      params,
    })
  },

  // Get group detail
  getDetail(id) {
    return request({
      url: `/admin/groups/${id}`,
      method: 'get',
    })
  },

  // Create group
  create(data) {
    return request({
      url: '/admin/groups',
      method: 'post',
      data,
    })
  },

  // Update group
  update(id, data) {
    return request({
      url: `/admin/groups/${id}`,
      method: 'put',
      data,
    })
  },

  // Delete group
  delete(id) {
    return request({
      url: `/admin/groups/${id}`,
      method: 'delete',
    })
  },

  // Assign categories to group
  assignCategories(id, categoryIds) {
    return request({
      url: `/admin/groups/${id}/categories`,
      method: 'post',
      data: { category_ids: categoryIds },
    })
  },

  // Assign roles to group
  assignRoles(id, roleIds) {
    return request({
      url: `/admin/groups/${id}/roles`,
      method: 'post',
      data: { role_ids: roleIds },
    })
  },
}

import request from '@/utils/request'

const permissionsApi = {
  // Get permissions list
  getList(params) {
    return request({
      url: '/admin/permissions',
      method: 'get',
      params
    })
  },

  // Get permission by ID
  getById(id) {
    return request({
      url: `/admin/permissions/${id}`,
      method: 'get'
    })
  }
}

export default permissionsApi

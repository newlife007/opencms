import request from '@/utils/request'

const levelsApi = {
  // Get levels list
  getList(params) {
    return request({
      url: '/admin/levels',
      method: 'get',
      params
    })
  },

  // Get level by ID
  getById(id) {
    return request({
      url: `/admin/levels/${id}`,
      method: 'get'
    })
  },

  // Create level
  create(data) {
    return request({
      url: '/admin/levels',
      method: 'post',
      data
    })
  },

  // Update level
  update(id, data) {
    return request({
      url: `/admin/levels/${id}`,
      method: 'put',
      data
    })
  },

  // Delete level
  delete(id) {
    return request({
      url: `/admin/levels/${id}`,
      method: 'delete'
    })
  }
}

export default levelsApi

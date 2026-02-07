import request from '@/utils/request'

export default {
  // Get category tree
  getTree() {
    return request({
      url: '/categories/tree',
      method: 'get',
    })
  },

  // Get category list
  getList(params) {
    return request({
      url: '/categories',
      method: 'get',
      params,
    })
  },

  // Get category detail
  getDetail(id) {
    return request({
      url: `/categories/${id}`,
      method: 'get',
    })
  },

  // Create category
  create(data) {
    return request({
      url: '/categories',
      method: 'post',
      data,
    })
  },

  // Update category
  update(id, data) {
    return request({
      url: `/categories/${id}`,
      method: 'put',
      data,
    })
  },

  // Delete category
  delete(id) {
    return request({
      url: `/categories/${id}`,
      method: 'delete',
    })
  },
}

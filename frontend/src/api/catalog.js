import request from '@/utils/request'

export default {
  // Get catalog list
  getList(params) {
    return request({
      url: '/catalog',
      method: 'get',
      params,
    })
  },

  // Get catalog tree by file type
  getTreeByType(fileType) {
    return request({
      url: `/catalog/tree`,
      method: 'get',
      params: { type: fileType },
    })
  },

  // Get catalog detail
  getDetail(id) {
    return request({
      url: `/catalog/${id}`,
      method: 'get',
    })
  },

  // Create catalog
  create(data) {
    return request({
      url: '/catalog',
      method: 'post',
      data,
    })
  },

  // Update catalog
  update(id, data) {
    return request({
      url: `/catalog/${id}`,
      method: 'put',
      data,
    })
  },

  // Delete catalog
  delete(id) {
    return request({
      url: `/catalog/${id}`,
      method: 'delete',
    })
  },
}

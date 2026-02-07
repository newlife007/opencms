import request from '@/utils/request'

export default {
  // Get file list
  getList(params) {
    return request({
      url: '/files',
      method: 'get',
      params,
    })
  },

  // Get file detail
  getDetail(id) {
    return request({
      url: `/files/${id}`,
      method: 'get',
    })
  },

  // Upload file
  upload(formData, onProgress) {
    return request({
      url: '/files',
      method: 'post',
      data: formData,
      headers: {
        'Content-Type': 'multipart/form-data',
      },
      onUploadProgress: onProgress,
    })
  },

  // Update file
  update(id, data) {
    return request({
      url: `/files/${id}`,
      method: 'put',
      data,
    })
  },

  // Delete file
  delete(id) {
    return request({
      url: `/files/${id}`,
      method: 'delete',
    })
  },

  // Submit for review
  submit(id) {
    return request({
      url: `/files/${id}/submit`,
      method: 'post',
    })
  },

  // Publish file
  publish(id) {
    return request({
      url: `/files/${id}/publish`,
      method: 'post',
    })
  },

  // Reject file
  reject(id, data) {
    return request({
      url: `/files/${id}/reject`,
      method: 'post',
      data,
    })
  },

  // Download file
  download(id) {
    return request({
      url: `/files/${id}/download`,
      method: 'get',
      responseType: 'blob',
    })
  },

  // Get preview URL
  getPreviewUrl(id) {
    return `/api/v1/files/${id}/preview`
  },

  // Get download URL
  getDownloadUrl(id) {
    return `/api/v1/files/${id}/download`
  },

  // Get file statistics
  getStats() {
    return request({
      url: '/files/stats',
      method: 'get',
    })
  },

  // Get recent files
  getRecent(limit = 10) {
    return request({
      url: '/files/recent',
      method: 'get',
      params: { limit },
    })
  },
}

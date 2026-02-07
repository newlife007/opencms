import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import * as filesAPI from '@/api/files'

export const useFilesStore = defineStore('files', () => {
  // State
  const fileList = ref([])
  const totalFiles = ref(0)
  const currentFile = ref(null)
  const uploadQueue = ref([])
  const loading = ref(false)
  
  // Filters
  const filters = ref({
    type: [],          // File types: 1=video, 2=audio, 3=image, 4=rich media
    status: [],        // Status: 0=new, 1=pending, 2=published, 3=rejected, 4=deleted
    categoryId: null,
    dateRange: [],
    searchQuery: '',
    level: null,
    groupId: null,
  })
  
  // Pagination
  const pagination = ref({
    page: 1,
    pageSize: 20,
  })
  
  // Sort
  const sortConfig = ref({
    sortBy: 'upload_at',
    sortOrder: 'desc',
  })

  // Getters
  const filteredFiles = computed(() => {
    return fileList.value
  })

  const uploadingFiles = computed(() => {
    return uploadQueue.value.filter(f => f.status === 'uploading')
  })

  const completedUploads = computed(() => {
    return uploadQueue.value.filter(f => f.status === 'completed')
  })

  const failedUploads = computed(() => {
    return uploadQueue.value.filter(f => f.status === 'failed')
  })

  // Actions
  const fetchFiles = async (params = {}) => {
    loading.value = true
    try {
      const queryParams = {
        page: pagination.value.page,
        page_size: pagination.value.pageSize,
        sort_by: sortConfig.value.sortBy,
        sort_order: sortConfig.value.sortOrder,
        ...filters.value,
        ...params,
      }

      // Clean up null/empty values
      Object.keys(queryParams).forEach(key => {
        if (queryParams[key] === null || queryParams[key] === '' || 
            (Array.isArray(queryParams[key]) && queryParams[key].length === 0)) {
          delete queryParams[key]
        }
      })

      const response = await filesAPI.getFileList(queryParams)
      if (response.success) {
        fileList.value = response.data.files || []
        totalFiles.value = response.data.total || 0
        return true
      }
      return false
    } catch (error) {
      console.error('Fetch files failed:', error)
      return false
    } finally {
      loading.value = false
    }
  }

  const fetchFileById = async (id) => {
    try {
      const response = await filesAPI.getFileDetail(id)
      if (response.success) {
        currentFile.value = response.data
        return response.data
      }
      return null
    } catch (error) {
      console.error('Fetch file by ID failed:', error)
      return null
    }
  }

  const uploadFile = async (file, metadata, onProgress) => {
    const uploadItem = {
      id: Date.now() + Math.random(),
      file,
      metadata,
      status: 'uploading',
      progress: 0,
      error: null,
      result: null,
    }
    
    uploadQueue.value.push(uploadItem)

    try {
      const formData = new FormData()
      formData.append('file', file)
      
      // Add metadata
      Object.keys(metadata).forEach(key => {
        if (metadata[key] !== null && metadata[key] !== undefined) {
          formData.append(key, metadata[key])
        }
      })

      const response = await filesAPI.uploadFile(formData, (progressEvent) => {
        const progress = Math.round((progressEvent.loaded * 100) / progressEvent.total)
        uploadItem.progress = progress
        if (onProgress) onProgress(progress)
      })

      if (response.success) {
        uploadItem.status = 'completed'
        uploadItem.result = response.data
        return response.data
      } else {
        uploadItem.status = 'failed'
        uploadItem.error = response.message || 'Upload failed'
        throw new Error(uploadItem.error)
      }
    } catch (error) {
      uploadItem.status = 'failed'
      uploadItem.error = error.message || 'Upload failed'
      throw error
    }
  }

  const updateFileCatalog = async (fileId, catalogData) => {
    try {
      const response = await filesAPI.updateFileCatalog(fileId, catalogData)
      if (response.success) {
        // Update current file if it matches
        if (currentFile.value && currentFile.value.id === fileId) {
          currentFile.value = { ...currentFile.value, ...response.data }
        }
        // Update in list
        const index = fileList.value.findIndex(f => f.id === fileId)
        if (index !== -1) {
          fileList.value[index] = { ...fileList.value[index], ...response.data }
        }
        return true
      }
      return false
    } catch (error) {
      console.error('Update file catalog failed:', error)
      throw error
    }
  }

  const deleteFile = async (id) => {
    try {
      const response = await filesAPI.deleteFile(id)
      if (response.success) {
        // Remove from list
        fileList.value = fileList.value.filter(f => f.id !== id)
        totalFiles.value -= 1
        return true
      }
      return false
    } catch (error) {
      console.error('Delete file failed:', error)
      throw error
    }
  }

  const updateFileStatus = async (id, status) => {
    try {
      const response = await filesAPI.updateFileStatus(id, status)
      if (response.success) {
        // Update in list
        const index = fileList.value.findIndex(f => f.id === id)
        if (index !== -1) {
          fileList.value[index].status = status
        }
        // Update current file if it matches
        if (currentFile.value && currentFile.value.id === id) {
          currentFile.value.status = status
        }
        return true
      }
      return false
    } catch (error) {
      console.error('Update file status failed:', error)
      throw error
    }
  }

  const downloadFile = async (id, filename) => {
    try {
      const response = await filesAPI.downloadFile(id)
      // Create blob link to download
      const url = window.URL.createObjectURL(new Blob([response]))
      const link = document.createElement('a')
      link.href = url
      link.setAttribute('download', filename)
      document.body.appendChild(link)
      link.click()
      link.remove()
      window.URL.revokeObjectURL(url)
      return true
    } catch (error) {
      console.error('Download file failed:', error)
      throw error
    }
  }

  const clearFilters = () => {
    filters.value = {
      type: [],
      status: [],
      categoryId: null,
      dateRange: [],
      searchQuery: '',
      level: null,
      groupId: null,
    }
  }

  const setFilter = (key, value) => {
    filters.value[key] = value
  }

  const setPagination = (page, pageSize) => {
    pagination.value.page = page
    if (pageSize) {
      pagination.value.pageSize = pageSize
    }
  }

  const setSort = (sortBy, sortOrder) => {
    sortConfig.value.sortBy = sortBy
    sortConfig.value.sortOrder = sortOrder
  }

  const clearUploadQueue = () => {
    uploadQueue.value = []
  }

  const removeFromUploadQueue = (id) => {
    uploadQueue.value = uploadQueue.value.filter(item => item.id !== id)
  }

  return {
    // State
    fileList,
    totalFiles,
    currentFile,
    uploadQueue,
    loading,
    filters,
    pagination,
    sortConfig,

    // Getters
    filteredFiles,
    uploadingFiles,
    completedUploads,
    failedUploads,

    // Actions
    fetchFiles,
    fetchFileById,
    uploadFile,
    updateFileCatalog,
    deleteFile,
    updateFileStatus,
    downloadFile,
    clearFilters,
    setFilter,
    setPagination,
    setSort,
    clearUploadQueue,
    removeFromUploadQueue,
  }
})

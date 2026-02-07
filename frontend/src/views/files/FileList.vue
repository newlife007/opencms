<template>
  <div class="file-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ t("fileList.title") }}</span>
          <el-button type="primary" icon="Upload" @click="$router.push('/files/upload')">
            上传文件
          </el-button>
        </div>
      </template>

      <!-- Filters -->
      <el-form :inline="true" :model="filters" class="filter-form">
        <el-form-item :label="t('files.fileType')">
          <el-select v-model="filters.type" placeholder="全部" clearable style="width: 150px;">
            <el-option label="视频" :value="1" />
            <el-option label="音频" :value="2" />
            <el-option label="图片" :value="3" />
            <el-option label="富媒体" :value="4" />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('files.status')">
          <el-select v-model="filters.status" placeholder="全部" clearable style="width: 150px;">
            <el-option label="新上传" :value="0" />
            <el-option label="待审核" :value="1" />
            <el-option label="已发布" :value="2" />
            <el-option label="已拒绝" :value="3" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" icon="Search" @click="loadFiles">查询</el-button>
          <el-button icon="Refresh" @click="resetFilters">{{ t("common.reset") }}</el-button>
        </el-form-item>
      </el-form>

      <!-- File Table -->
      <el-table :data="files" v-loading="loading" style="width: 100%">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="title" label="标题" min-width="200" />
        <el-table-column prop="type" label="类型" width="100">
          <template #default="{ row }">
            <el-tag>{{ getFileTypeName(row.type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" :label="t('files.status')" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">
              {{ getStatusName(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="size" label="大小" width="120">
          <template #default="{ row }">
            {{ formatFileSize(row.size) }}
          </template>
        </el-table-column>
        <el-table-column prop="upload_username" :label="t('files.uploader')" width="120" />
        <el-table-column prop="upload_at" :label="t('files.uploadTime')" width="180">
          <template #default="{ row }">
            {{ formatDate(row.upload_at) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('common.actions')" width="280" fixed="right" align="center">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button 
                size="small" 
                type="primary" 
                @click="viewDetail(row.id)" 
                title="查看详情"
              >
                <el-icon><View /></el-icon>
                详情
              </el-button>
              <el-button 
                size="small" 
                type="success" 
                @click="handleSubmit(row)" 
                v-if="row.status === 0"
                title="提交审核"
              >
                <el-icon><Upload /></el-icon>{{ t("common.submit") }}</el-button>
              <el-button 
                size="small" 
                type="warning" 
                @click="handleDownload(row.id)" 
                title="下载文件"
              >
                <el-icon><Download /></el-icon>{{ t("common.download") }}</el-button>
              <el-button 
                size="small" 
                type="danger" 
                @click="handleDelete(row.id)" 
                title="删除文件"
              >
                <el-icon><Delete /></el-icon>{{ t("common.delete") }}</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <!-- Pagination -->
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="loadFiles"
        @current-change="loadFiles"
        style="margin-top: 20px; justify-content: flex-end"
      />
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { View, Download, Delete, Upload } from '@element-plus/icons-vue'
import filesApi from '@/api/files'


const { t } = useI18n()

const router = useRouter()
const loading = ref(false)
const files = ref([])

const filters = reactive({
  type: null,
  status: null,
})

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0,
})

const getFileTypeName = (type) => {
  const types = { 1: '视频', 2: '音频', 3: '图片', 4: '富媒体' }
  return types[type] || '未知'
}

const getStatusName = (status) => {
  const statuses = {
    0: '新上传',
    1: '待审核',
    2: '已发布',
    3: '已拒绝',
    4: '已删除',
  }
  return statuses[status] || '未知'
}

const getStatusType = (status) => {
  const types = {
    0: 'info',
    1: 'warning',
    2: 'success',
    3: 'danger',
    4: 'info',
  }
  return types[status] || 'info'
}

const formatFileSize = (bytes) => {
  if (!bytes) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i]
}

const formatDate = (timestamp) => {
  if (!timestamp) return '-'
  return new Date(timestamp * 1000).toLocaleString('zh-CN')
}

const loadFiles = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      page_size: pagination.pageSize,
      ...filters,
    }
    
    const res = await filesApi.getList(params)
    if (res.success) {
      // Backend returns: { success: true, data: [...], pagination: {...} }
      files.value = res.data || []
      pagination.total = res.pagination?.total || 0
    }
  } catch (error) {
    console.error('Failed to load files:', error)
    ElMessage.error('加载文件列表失败')
  } finally {
    loading.value = false
  }
}

const resetFilters = () => {
  filters.type = null
  filters.status = null
  pagination.page = 1
  loadFiles()
}

const viewDetail = (id) => {
  router.push(`/files/${id}`)
}

const handleSubmit = async (file) => {
  try {
    await ElMessageBox.confirm('确定要提交此文件进行审核吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    })
    
    await filesApi.submit(file.id)
    ElMessage.success('已提交审核')
    loadFiles()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('提交失败')
    }
  }
}

const handleDownload = async (id) => {
  try {
    // Use direct link download to preserve Content-Disposition and Content-Type from backend
    const downloadUrl = `/api/v1/files/${id}/download`
    const link = document.createElement('a')
    link.href = downloadUrl
    link.style.display = 'none'
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    ElMessage.success('下载成功')
  } catch (error) {
    ElMessage.error('下载失败')
  }
}

const handleDelete = async (id) => {
  try {
    await ElMessageBox.confirm('确定要删除此文件吗？', '警告', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    })
    
    await filesApi.delete(id)
    ElMessage.success(t('message.deleteSuccess'))
    loadFiles()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(t('message.deleteFailed'))
    }
  }
}

onMounted(() => {
  loadFiles()
})
</script>

<style scoped>
.file-list {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.filter-form {
  margin-bottom: 20px;
}

/* 操作按钮组样式优化 */
.action-buttons {
  display: flex;
  flex-wrap: nowrap;
  gap: 4px;
}

.action-buttons .el-button {
  margin: 0 !important;
  padding: 5px 10px;
  font-size: 12px;
  white-space: nowrap;
}

/* 确保按钮在一行显示 */
:deep(.el-table__cell) {
  overflow: visible !important;
}

/* 按钮悬停效果 */
.action-buttons .el-button:hover {
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
  transition: all 0.2s ease;
}
</style>

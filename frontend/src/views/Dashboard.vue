<template>
  <div class="dashboard">
    <!-- 无权限提示 -->
    <el-alert
      v-if="hasNoPermissions"
      :title="t('common.info')"
      type="info"
      :closable="false"
      center
      show-icon
      style="margin-bottom: 20px;"
    >
      <template #default>
        <div class="no-permission-message">
          <p style="font-size: 18px; margin: 20px 0; font-weight: 500;">
            {{ t('dashboard.contactAdmin') }}
          </p>
        </div>
      </template>
    </el-alert>

    <!-- 有权限时显示正常内容 -->
    <template v-else>
      <el-row :gutter="20">
        <el-col :span="6">
          <el-card class="stat-card">
            <div class="stat-content">
              <el-icon class="stat-icon" color="#409eff"><Document /></el-icon>
              <div class="stat-info">
                <div class="stat-value">{{ stats.totalFiles }}</div>
                <div class="stat-label">{{ t('dashboard.totalFiles') }}</div>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="stat-card">
            <div class="stat-content">
              <el-icon class="stat-icon" color="#67c23a"><VideoPlay /></el-icon>
              <div class="stat-info">
                <div class="stat-value">{{ stats.videoFiles }}</div>
                <div class="stat-label">{{ t('dashboard.videoFiles') }}</div>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="stat-card">
            <div class="stat-content">
              <el-icon class="stat-icon" color="#e6a23c"><Microphone /></el-icon>
              <div class="stat-info">
                <div class="stat-value">{{ stats.audioFiles }}</div>
                <div class="stat-label">{{ t('dashboard.audioFiles') }}</div>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="stat-card">
            <div class="stat-content">
              <el-icon class="stat-icon" color="#f56c6c"><Picture /></el-icon>
              <div class="stat-info">
                <div class="stat-value">{{ stats.imageFiles }}</div>
                <div class="stat-label">{{ t('dashboard.imageFiles') }}</div>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>

      <el-row :gutter="20" style="margin-top: 20px">
        <el-col :span="16">
          <el-card>
            <template #header>
              <span>{{ t('dashboard.recentUploads') }}</span>
            </template>
            <el-table :data="recentFiles" style="width: 100%">
              <el-table-column prop="title" :label="t('files.fileName')" />
              <el-table-column prop="type" :label="t('files.fileType')" width="100">
                <template #default="{ row }">
                  <el-tag>{{ getFileTypeName(row.type) }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="size" :label="t('files.fileSize')" width="120">
                <template #default="{ row }">
                  {{ formatFileSize(row.size) }}
                </template>
              </el-table-column>
              <el-table-column prop="upload_at" :label="t('files.uploadTime')" width="180">
                <template #default="{ row }">
                  {{ formatDate(row.upload_at) }}
                </template>
              </el-table-column>
            </el-table>
          </el-card>
        </el-col>
        <el-col :span="8">
          <el-card>
            <template #header>
              <span>{{ t('dashboard.quickLinks') }}</span>
            </template>
            <div class="quick-links">
              <el-button 
                v-if="hasUploadPermission"
                type="primary" 
                icon="Upload" 
                @click="$router.push('/files/upload')"
              >
                {{ t('dashboard.uploadFile') }}
              </el-button>
              <el-button 
                v-if="hasFilesPermission"
                type="success" 
                icon="Document" 
                @click="$router.push('/files')"
              >
                {{ t('dashboard.fileManagement') }}
              </el-button>
              <el-button 
                v-if="hasSearchPermission"
                type="warning" 
                icon="Search" 
                @click="$router.push('/search')"
              >
                {{ t('dashboard.searchFiles') }}
              </el-button>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </template>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { useUserStore } from '@/stores/user'
import filesApi from '@/api/files'

const { t } = useI18n()
const userStore = useUserStore()

const stats = ref({
  totalFiles: 0,
  videoFiles: 0,
  audioFiles: 0,
  imageFiles: 0,
})

const recentFiles = ref([])

// 检查用户是否有任何权限
const hasNoPermissions = computed(() => {
  // 管理员始终有权限（优先检查）
  if (userStore.isAdmin()) {
    console.log('User is admin, has all permissions')
    return false
  }
  
  // 检查permissions数组（从store中获取）
  const perms = userStore.permissions || []
  console.log('Dashboard checking permissions:', perms, 'Length:', perms.length)
  
  return perms.length === 0
})

// 检查特定功能权限
const hasFilesPermission = computed(() => 
  userStore.hasPermission('files.browse.list')
)

const hasUploadPermission = computed(() => 
  userStore.hasPermission('files.upload.create')
)

const hasSearchPermission = computed(() => 
  userStore.hasPermission('files.browse.search')
)

const getFileTypeName = (type) => {
  const types = { 
    1: t('files.type.video'), 
    2: t('files.type.audio'), 
    3: t('files.type.image'), 
    4: t('files.type.document') 
  }
  return types[type] || t('common.unknown')
}

const formatFileSize = (bytes) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i]
}

const formatDate = (timestamp) => {
  if (!timestamp) return '-'
  if (typeof timestamp === 'string') {
    return new Date(timestamp).toLocaleString('zh-CN')
  }
  return new Date(timestamp * 1000).toLocaleString('zh-CN')
}

const loadDashboardData = async () => {
  // 如果没有权限，不加载数据
  if (hasNoPermissions.value) {
    return
  }
  
  try {
    console.log('Loading dashboard statistics...')
    const statsRes = await filesApi.getStats()
    console.log('Stats response:', statsRes)
    
    if (statsRes.success) {
      stats.value = {
        totalFiles: statsRes.data.total || 0,
        videoFiles: statsRes.data.video || 0,
        audioFiles: statsRes.data.audio || 0,
        imageFiles: statsRes.data.image || 0,
      }
    } else {
      // 如果是权限问题，不显示错误
      if (statsRes.message !== 'Permission denied') {
        ElMessage.error(t('message.loadFailed'))
      }
    }
    
    console.log('Loading recent files...')
    const recentRes = await filesApi.getRecent(5)
    console.log('Recent files response:', recentRes)
    
    if (recentRes.success) {
      recentFiles.value = recentRes.data || []
    } else {
      // 如果是权限问题，不显示错误
      if (recentRes.message !== 'Permission denied') {
        ElMessage.error(t('message.loadFailed'))
      }
    }
  } catch (error) {
    console.error('Load dashboard error:', error)
    // 如果是权限问题，不显示错误
    if (error.response?.status !== 403) {
      ElMessage.error(t('message.loadFailed') + ': ' + (error.message || t('message.networkError')))
    }
  }
}

onMounted(() => {
  loadDashboardData()
})
</script>

<style scoped>
.dashboard {
  padding: 20px;
}

.no-permission-message {
  text-align: center;
  padding: 20px;
}

.stat-card {
  cursor: pointer;
  transition: transform 0.3s;
}

.stat-card:hover {
  transform: translateY(-5px);
}

.stat-content {
  display: flex;
  align-items: center;
  gap: 20px;
}

.stat-icon {
  font-size: 48px;
}

.stat-info {
  flex: 1;
}

.stat-value {
  font-size: 32px;
  font-weight: bold;
  color: #333;
}

.stat-label {
  font-size: 14px;
  color: #666;
  margin-top: 5px;
}

.quick-links {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.quick-links .el-button {
  width: 100%;
}
</style>

<template>
  <div class="search-page">
    <el-card>
      <template #header>
        <span>{{ t('search.title') }}</span>
      </template>

      <!-- Search Form -->
      <el-form :model="searchForm" class="search-form">
        <el-row :gutter="20">
          <el-col :span="16">
            <el-input
              v-model="searchForm.keyword"
              :placeholder="t('search.placeholder')"
              size="large"
              clearable
              @keyup.enter="handleSearch"
            >
              <template #prefix>
                <el-icon><Search /></el-icon>
              </template>
              <template #append>
                <el-button icon="Search" @click="handleSearch">{{ t('search.searchButton') }}</el-button>
              </template>
            </el-input>
          </el-col>
          <el-col :span="8">
            <el-button @click="showAdvanced = !showAdvanced">
              <el-icon><Filter /></el-icon>
              {{ t('search.advancedSearch') }}
            </el-button>
            <el-button @click="resetSearch">
              <el-icon><Refresh /></el-icon>
              {{ t('common.reset') }}
            </el-button>
          </el-col>
        </el-row>

        <!-- Advanced Filters -->
        <el-collapse-transition>
          <div v-show="showAdvanced" class="advanced-filters">
            <el-divider />
            <el-row :gutter="20">
              <el-col :span="8">
                <el-form-item :label="t('search.fileType')">
                  <el-checkbox-group v-model="searchForm.types">
                    <el-checkbox :label="1">{{ t('files.type.video') }}</el-checkbox>
                    <el-checkbox :label="2">{{ t('files.type.audio') }}</el-checkbox>
                    <el-checkbox :label="3">{{ t('files.type.image') }}</el-checkbox>
                    <el-checkbox :label="4">{{ t('files.type.document') }}</el-checkbox>
                  </el-checkbox-group>
                </el-form-item>
              </el-col>
              <el-col :span="8">
                <el-form-item :label="t('search.uploadTime')">
                  <el-date-picker
                    v-model="searchForm.dateRange"
                    type="daterange"
                    :range-separator="t('search.rangeSeparator')"
                    :start-placeholder="t('search.startDate')"
                    :end-placeholder="t('search.endDate')"
                    format="YYYY-MM-DD"
                    value-format="YYYY-MM-DD"
                  />
                </el-form-item>
              </el-col>
              <el-col :span="8">
                <el-form-item :label="t('search.uploader')">
                  <el-input v-model="searchForm.uploader" :placeholder="t('search.uploaderPlaceholder')" />
                </el-form-item>
              </el-col>
            </el-row>
            <el-row :gutter="20">
              <el-col :span="12">
                <el-form-item :label="t('search.sortBy')">
                  <el-select v-model="searchForm.sortBy">
                    <el-option :label="t('search.relevance')" value="relevance" />
                    <el-option :label="t('search.uploadTimeDesc')" value="upload_time_desc" />
                    <el-option :label="t('search.uploadTimeAsc')" value="upload_time_asc" />
                    <el-option :label="t('search.sizeDesc')" value="size_desc" />
                    <el-option :label="t('search.sizeAsc')" value="size_asc" />
                  </el-select>
                </el-form-item>
              </el-col>
            </el-row>
          </div>
        </el-collapse-transition>
      </el-form>

      <!-- Search Results -->
      <div v-if="searched" class="search-results">
        <el-divider />
        
        <div class="results-header">
          <span class="results-count">{{ t('search.resultsCount', { count: total }) }}</span>
        </div>

        <el-row :gutter="20" v-loading="loading">
          <el-col :span="24" v-for="file in results" :key="file.id">
            <el-card class="result-card" @click="viewDetail(file.id)">
              <el-row :gutter="20">
                <el-col :span="2">
                  <div class="file-icon">
                    <el-icon :size="40" :color="getFileTypeColor(file.type)">
                      <component :is="getFileTypeIcon(file.type)" />
                    </el-icon>
                  </div>
                </el-col>
                <el-col :span="18">
                  <div class="file-info">
                    <h3 class="file-title" v-html="file.snippet || highlightText(file.title)"></h3>
                    <p class="file-description" v-if="file.description" v-html="highlightText(file.description)"></p>
                    <div class="file-meta">
                      <el-tag size="small">{{ getFileTypeName(file.type) }}</el-tag>
                      <el-tag size="small" :type="getStatusType(file.status)">
                        {{ getStatusName(file.status) }}
                      </el-tag>
                      <span class="meta-item">{{ formatFileSize(file.size) }}</span>
                      <span class="meta-item">{{ t('search.uploader') }}: {{ file.upload_username }}</span>
                      <span class="meta-item">{{ formatDate(file.upload_at) }}</span>
                    </div>
                  </div>
                </el-col>
                <el-col :span="4">
                  <div class="file-actions">
                    <el-button type="primary" size="small" @click.stop="viewDetail(file.id)">
                      {{ t('fileList.viewDetail') }}
                    </el-button>
                  </div>
                </el-col>
              </el-row>
            </el-card>
          </el-col>
        </el-row>

        <!-- Pagination -->
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :total="total"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSearch"
          @current-change="handleSearch"
          style="margin-top: 20px; justify-content: center"
        />
      </div>

      <el-empty v-else :description="t('search.placeholder')" />
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { VideoPlay, Microphone, Picture, Document } from '@element-plus/icons-vue'
import * as searchApi from '@/api/search'

const { t } = useI18n()
const router = useRouter()
const loading = ref(false)
const searched = ref(false)
const showAdvanced = ref(false)
const results = ref([])
const total = ref(0)

const searchForm = reactive({
  keyword: '',
  types: [],
  dateRange: null,
  uploader: '',
  sortBy: 'relevance',
})

const pagination = reactive({
  page: 1,
  pageSize: 20,
})

const handleSearch = async () => {
  if (!searchForm.keyword.trim()) {
    ElMessage.warning(t('search.emptyKeyword'))
    return
  }

  loading.value = true
  searched.value = true

  try {
    const params = {
      q: searchForm.keyword,
      types: searchForm.types,
      date_from: searchForm.dateRange?.[0],
      date_to: searchForm.dateRange?.[1],
      uploader: searchForm.uploader,
      sort_by: searchForm.sortBy,
      page: pagination.page,
      page_size: pagination.pageSize,
    }
    
    console.log('Searching with params:', params)
    const res = await searchApi.search(params)
    console.log('Search response:', res)
    
    if (res.success) {
      results.value = res.results || []
      total.value = res.pagination?.total || 0
      
      if (results.value.length === 0) {
        ElMessage.info(t('search.noResults'))
      } else {
        ElMessage.success(t('search.resultsCount', { count: total.value }))
      }
    } else {
      ElMessage.error(res.message || t('message.loadFailed'))
      results.value = []
      total.value = 0
    }
    
    loading.value = false
  } catch (error) {
    console.error('Search error:', error)
    ElMessage.error(t('message.loadFailed') + ': ' + (error.message || t('message.networkError')))
    loading.value = false
    results.value = []
    total.value = 0
  }
}

const resetSearch = () => {
  searchForm.keyword = ''
  searchForm.types = []
  searchForm.dateRange = null
  searchForm.uploader = ''
  searchForm.sortBy = 'relevance'
  pagination.page = 1
  results.value = []
  total.value = 0
  searched.value = false
}

const viewDetail = (id) => {
  router.push(`/files/${id}`)
}

const getFileTypeName = (type) => {
  const types = { 
    1: t('files.type.video'), 
    2: t('files.type.audio'), 
    3: t('files.type.image'), 
    4: t('files.type.document') 
  }
  return types[type] || t('common.unknown')
}

const getStatusName = (status) => {
  const statuses = {
    0: t('files.status.new'),
    1: t('files.status.pending'),
    2: t('files.status.published'),
    3: t('files.status.rejected'),
    4: t('files.status.deleted'),
  }
  return statuses[status] || t('common.unknown')
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

const getFileTypeIcon = (type) => {
  const icons = {
    1: VideoPlay,
    2: Microphone,
    3: Picture,
    4: Document,
  }
  return icons[type] || Document
}

const getFileTypeColor = (type) => {
  const colors = {
    1: '#67c23a',
    2: '#e6a23c',
    3: '#f56c6c',
    4: '#409eff',
  }
  return colors[type] || '#909399'
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
  if (typeof timestamp === 'string') {
    return new Date(timestamp).toLocaleString('zh-CN')
  }
  return new Date(timestamp * 1000).toLocaleString('zh-CN')
}

const highlightText = (text) => {
  if (!text || !searchForm.keyword) return text
  const keyword = searchForm.keyword.trim()
  const escapedKeyword = keyword.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
  const regex = new RegExp(`(${escapedKeyword})`, 'gi')
  return text.replace(regex, '<mark>$1</mark>')
}
</script>

<style scoped>
.search-page {
  padding: 20px;
}

.search-form {
  margin-bottom: 20px;
}

.advanced-filters {
  margin-top: 20px;
}

.results-header {
  margin-bottom: 20px;
}

.results-count {
  font-size: 14px;
  color: #666;
}

.result-card {
  margin-bottom: 15px;
  cursor: pointer;
  transition: all 0.3s;
}

.result-card:hover {
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
}

.file-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
}

.file-info {
  flex: 1;
}

.file-title {
  font-size: 16px;
  font-weight: 500;
  margin-bottom: 8px;
  color: #303133;
}

.file-title :deep(mark) {
  background-color: #ffd700;
  padding: 2px 4px;
}

.file-description {
  font-size: 14px;
  color: #606266;
  margin-bottom: 10px;
  line-height: 1.5;
}

.file-description :deep(mark) {
  background-color: #ffd700;
  padding: 2px 4px;
}

.file-meta {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 12px;
  color: #909399;
}

.meta-item {
  display: inline-flex;
  align-items: center;
}

.file-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  height: 100%;
}
</style>

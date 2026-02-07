<template>
  <div class="file-approval-container">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>文件审批</span>
          <el-radio-group v-model="activeTab" @change="fetchFiles">
            <el-radio-button label="pending">待审批</el-radio-button>
            <el-radio-button label="approved">已通过</el-radio-button>
            <el-radio-button label="rejected">已拒绝</el-radio-button>
          </el-radio-group>
        </div>
      </template>

      <!-- 筛选条件 -->
      <el-form :inline="true" :model="filters" class="filter-form">
        <el-form-item :label="t('files.fileType')">
          <el-select 
            v-model="filters.type" 
            placeholder="全部" 
            clearable 
            @change="fetchFiles"
            style="width: 180px;"
          >
            <el-option label="全部类型" :value="null" />
            <el-option label="视频" :value="1" />
            <el-option label="音频" :value="2" />
            <el-option label="图片" :value="3" />
            <el-option label="富媒体" :value="4" />
          </el-select>
        </el-form-item>

        <el-form-item :label="t('files.category')">
          <el-tree-select
            v-model="filters.category_id"
            :data="categoryTree"
            :props="{ label: 'name', value: 'id' }"
            placeholder="全部分类"
            check-strictly
            clearable
            @change="fetchFiles"
            style="width: 220px;"
          />
        </el-form-item>

        <el-form-item :label="t('files.uploader')">
          <el-input 
            v-model="filters.uploader" 
            placeholder="输入用户名" 
            clearable 
            @change="fetchFiles"
            style="width: 160px;"
          />
        </el-form-item>

        <el-form-item :label="t('files.uploadTime')">
          <el-date-picker
            v-model="filters.dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            @change="fetchFiles"
            style="width: 340px;"
          />
        </el-form-item>

        <el-form-item>
          <el-button type="primary" @click="fetchFiles">
            <el-icon><Search /></el-icon>
            查询
          </el-button>
          <el-button @click="resetFilters">
            <el-icon><Refresh /></el-icon>{{ t("common.reset") }}</el-button>
        </el-form-item>
      </el-form>

      <!-- 统计信息 -->
      <el-row :gutter="20" class="stats-row">
        <el-col :span="6">
          <el-statistic title="待审批" :value="stats.pending">
            <template #suffix>
              <span class="stat-suffix">个</span>
            </template>
          </el-statistic>
        </el-col>
        <el-col :span="6">
          <el-statistic title="今日已审批" :value="stats.today_approved">
            <template #suffix>
              <span class="stat-suffix">个</span>
            </template>
          </el-statistic>
        </el-col>
        <el-col :span="6">
          <el-statistic title="本周已审批" :value="stats.week_approved">
            <template #suffix>
              <span class="stat-suffix">个</span>
            </template>
          </el-statistic>
        </el-col>
        <el-col :span="6">
          <el-statistic title="通过率" :value="stats.approval_rate" precision="1">
            <template #suffix>
              <span class="stat-suffix">%</span>
            </template>
          </el-statistic>
        </el-col>
      </el-row>

      <!-- 文件列表 -->
      <el-table
        v-loading="loading"
        :data="fileList"
        stripe
        style="width: 100%"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="55" />
        
        <el-table-column label="缩略图" width="100">
          <template #default="{ row }">
            <el-image
              v-if="row.type === 3"
              :src="row.path"
              fit="cover"
              style="width: 60px; height: 60px; border-radius: 4px;"
            >
              <template #error>
                <el-icon :size="40"><Picture /></el-icon>
              </template>
            </el-image>
            <div v-else class="file-icon">
              <el-icon :size="40">
                <VideoCamera v-if="row.type === 1" />
                <Headset v-else-if="row.type === 2" />
                <Document v-else />
              </el-icon>
            </div>
          </template>
        </el-table-column>

        <el-table-column prop="title" label="标题" min-width="200" show-overflow-tooltip />
        
        <el-table-column prop="type" label="类型" width="80">
          <template #default="{ row }">
            <el-tag :type="getFileTypeColor(row.type)">
              {{ getFileTypeName(row.type) }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="category_name" :label="t('files.category')" width="120" show-overflow-tooltip />

        <el-table-column prop="size" label="大小" width="100">
          <template #default="{ row }">
            {{ formatFileSize(row.size) }}
          </template>
        </el-table-column>

        <el-table-column prop="upload_username" :label="t('files.uploader')" width="100" />

        <el-table-column prop="upload_at" :label="t('files.uploadTime')" width="160">
          <template #default="{ row }">
            {{ formatDate(row.upload_at) }}
          </template>
        </el-table-column>

        <el-table-column v-if="activeTab === 'pending'" label="待审时长" width="100">
          <template #default="{ row }">
            <el-tag type="warning">{{ getPendingDuration(row.upload_at) }}</el-tag>
          </template>
        </el-table-column>

        <el-table-column v-if="activeTab !== 'pending'" prop="catalog_at" label="审批时间" width="160">
          <template #default="{ row }">
            {{ formatDate(row.catalog_at) }}
          </template>
        </el-table-column>

        <el-table-column v-if="activeTab !== 'pending'" prop="catalog_username" label="审批人" width="100" />

        <el-table-column :label="t('common.actions')" width="240" fixed="right">
          <template #default="{ row }">
            <el-button-group>
              <el-button size="small" @click="viewFile(row.id)">
                <el-icon><View /></el-icon>{{ t("common.view") }}</el-button>
              
              <el-button 
                v-if="activeTab === 'pending'" 
                type="success" 
                size="small" 
                @click="approveFile(row.id)"
              >
                <el-icon><Check /></el-icon>{{ t("fileApproval.approve") }}</el-button>
              
              <el-button 
                v-if="activeTab === 'pending'" 
                type="danger" 
                size="small" 
                @click="rejectFile(row.id)"
              >
                <el-icon><Close /></el-icon>{{ t("fileApproval.reject") }}</el-button>

              <el-button 
                v-if="activeTab === 'approved'" 
                type="warning" 
                size="small" 
                @click="unpublishFile(row.id)"
              >
                <el-icon><Warning /></el-icon>
                撤回
              </el-button>
            </el-button-group>
          </template>
        </el-table-column>

        <!-- 空状态 -->
        <template #empty>
          <el-empty :description="getEmptyDescription()">
            <el-button v-if="hasFilters()" type="primary" @click="resetFilters">
              清除筛选条件
            </el-button>
          </el-empty>
        </template>
      </el-table>

      <!-- 批量操作 -->
      <div v-if="selectedFiles.length > 0 && activeTab === 'pending'" class="batch-actions">
        <el-alert
          :title="`已选择 ${selectedFiles.length} 个文件`"
          type="info"
          :closable="false"
        >
          <template #default>
            <el-button type="success" size="small" @click="batchApprove">
              <el-icon><Check /></el-icon>{{ t("fileApproval.batchApprove") }}</el-button>
            <el-button type="danger" size="small" @click="batchReject">
              <el-icon><Close /></el-icon>{{ t("fileApproval.batchReject") }}</el-button>
          </template>
        </el-alert>
      </div>

      <!-- 分页 -->
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :page-sizes="[10, 20, 50, 100]"
        :total="pagination.total"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="fetchFiles"
        @current-change="fetchFiles"
        class="pagination"
      />
    </el-card>

    <!-- 审批详情对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="800px"
    >
      <el-form :model="approvalForm" label-width="100px">
        <el-form-item label="审批意见">
          <el-input
            v-model="approvalForm.comment"
            type="textarea"
            :rows="4"
            :placeholder="approvalForm.action === 'approve' ? '请输入审批通过意见（可选）' : '请输入拒绝原因（必填）'"
          />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">{{ t("common.cancel") }}</el-button>
        <el-button 
          :type="approvalForm.action === 'approve' ? 'success' : 'danger'" 
          @click="submitApproval"
          :loading="submitting"
        >
          确认{{ approvalForm.action === 'approve' ? '通过' : '拒绝' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Search,
  Refresh,
  VideoCamera,
  Headset,
  Picture,
  Document,
  View,
  Check,
  Close,
  Warning
} from '@element-plus/icons-vue'
import axios from '@/utils/request'


const { t } = useI18n()

const router = useRouter()

// 数据
const activeTab = ref('pending')
const loading = ref(false)
const submitting = ref(false)
const fileList = ref([])
const selectedFiles = ref([])
const categoryTree = ref([])
const dialogVisible = ref(false)
const dialogTitle = ref('')

// 筛选条件
const filters = reactive({
  type: null,
  category_id: null,
  uploader: '',
  dateRange: null
})

// 分页
const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
})

// 统计信息
const stats = reactive({
  pending: 0,
  today_approved: 0,
  week_approved: 0,
  approval_rate: 0
})

// 审批表单
const approvalForm = reactive({
  action: 'approve',
  fileId: null,
  comment: ''
})

// 获取文件列表
const fetchFiles = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      page_size: pagination.pageSize,  // 修改为page_size
      status: activeTab.value === 'pending' ? 1 : activeTab.value === 'approved' ? 2 : 3
    }

    // 只添加有值的筛选条件
    if (filters.type) params.type = filters.type
    if (filters.category_id) params.category_id = filters.category_id
    if (filters.uploader) params.uploader = filters.uploader

    if (filters.dateRange && filters.dateRange.length === 2) {
      params.start_date = Math.floor(filters.dateRange[0].getTime() / 1000)
      params.end_date = Math.floor(filters.dateRange[1].getTime() / 1000)
    }

    const response = await axios.get('/files', { params })
    
    // 处理不同的响应格式
    if (response.data) {
      // 如果是 { data: [], total: 0 } 格式
      if (Array.isArray(response.data.data)) {
        fileList.value = response.data.data
        pagination.total = response.data.total || 0
      }
      // 如果是 { items: [], total: 0 } 格式
      else if (Array.isArray(response.data.items)) {
        fileList.value = response.data.items
        pagination.total = response.data.total || 0
      }
      // 如果直接是数组
      else if (Array.isArray(response.data)) {
        fileList.value = response.data
        pagination.total = response.data.length
      }
      // 否则为空
      else {
        fileList.value = []
        pagination.total = 0
        console.warn('未知的响应格式:', response.data)
      }
    } else {
      fileList.value = []
      pagination.total = 0
    }

    // 如果没有数据，显示提示
    if (fileList.value.length === 0 && !filters.type && !filters.category_id && !filters.uploader) {
      console.log('当前标签页暂无数据:', activeTab.value)
    }

  } catch (error) {
    console.error('获取文件列表失败:', error)
    ElMessage.error('获取文件列表失败：' + (error.response?.data?.message || error.message))
    fileList.value = []
    pagination.total = 0
  } finally {
    loading.value = false
  }
}

// 获取统计信息
const fetchStats = async () => {
  try {
    const response = await axios.get('/admin/workflow/stats')
    Object.assign(stats, response.data)
  } catch (error) {
    console.error('获取统计信息失败：', error)
  }
}

// 获取分类树
const fetchCategories = async () => {
  try {
    const response = await axios.get('/categories')
    categoryTree.value = buildTree(response.data)
  } catch (error) {
    console.error('获取分类失败：', error)
  }
}

// 构建树结构
const buildTree = (items, parentId = null) => {
  return items
    .filter(item => item.parent_id === parentId)
    .map(item => ({
      ...item,
      children: buildTree(items, item.id)
    }))
}

// 重置筛选
const resetFilters = () => {
  filters.type = null
  filters.category_id = null
  filters.uploader = ''
  filters.dateRange = null
  pagination.page = 1
  fetchFiles()
}

// 查看文件
const viewFile = (fileId) => {
  router.push(`/files/${fileId}`)
}

// 通过文件
const approveFile = (fileId) => {
  approvalForm.action = 'approve'
  approvalForm.fileId = fileId
  approvalForm.comment = ''
  dialogTitle.value = '审批通过'
  dialogVisible.value = true
}

// 拒绝文件
const rejectFile = (fileId) => {
  approvalForm.action = 'reject'
  approvalForm.fileId = fileId
  approvalForm.comment = ''
  dialogTitle.value = '审批拒绝'
  dialogVisible.value = true
}

// 提交审批
const submitApproval = async () => {
  if (approvalForm.action === 'reject' && !approvalForm.comment) {
    ElMessage.warning('请输入拒绝原因')
    return
  }

  submitting.value = true
  try {
    const status = approvalForm.action === 'approve' ? 2 : 3
    
    await axios.put(`/files/${approvalForm.fileId}`, {
      status,
      comment: approvalForm.comment
    })

    ElMessage.success(approvalForm.action === 'approve' ? '审批通过成功' : '已拒绝该文件')
    dialogVisible.value = false
    
    // 刷新列表和统计
    await Promise.all([fetchFiles(), fetchStats()])

  } catch (error) {
    ElMessage.error('操作失败：' + (error.response?.data?.message || error.message))
  } finally {
    submitting.value = false
  }
}

// 撤回发布
const unpublishFile = async (fileId) => {
  try {
    await ElMessageBox.confirm('确定要撤回该文件的发布状态吗？', '确认撤回', {
      type: 'warning'
    })

    await axios.put(`/files/${fileId}`, { status: 1 })
    ElMessage.success('已撤回发布')
    await Promise.all([fetchFiles(), fetchStats()])

  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('操作失败：' + (error.response?.data?.message || error.message))
    }
  }
}

// 批量通过
const batchApprove = async () => {
  try {
    await ElMessageBox.confirm(`确定要批量通过 ${selectedFiles.value.length} 个文件吗？`, '批量审批', {
      type: 'success'
    })

    const promises = selectedFiles.value.map(file =>
      axios.put(`/files/${file.id}`, { status: 2 })
    )

    await Promise.all(promises)
    ElMessage.success(`已批量通过 ${selectedFiles.value.length} 个文件`)
    
    selectedFiles.value = []
    await Promise.all([fetchFiles(), fetchStats()])

  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('操作失败：' + (error.response?.data?.message || error.message))
    }
  }
}

// 批量拒绝
const batchReject = async () => {
  try {
    const { value: comment } = await ElMessageBox.prompt('请输入拒绝原因', '批量拒绝', {
      inputType: 'textarea',
      inputValidator: (value) => {
        if (!value) return '请输入拒绝原因'
        return true
      }
    })

    const promises = selectedFiles.value.map(file =>
      axios.put(`/files/${file.id}`, { status: 3, comment })
    )

    await Promise.all(promises)
    ElMessage.success(`已批量拒绝 ${selectedFiles.value.length} 个文件`)
    
    selectedFiles.value = []
    await Promise.all([fetchFiles(), fetchStats()])

  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('操作失败：' + (error.response?.data?.message || error.message))
    }
  }
}

// 选择变化
const handleSelectionChange = (selection) => {
  selectedFiles.value = selection
}

// 格式化文件大小
const formatFileSize = (bytes) => {
  if (!bytes) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return (bytes / Math.pow(k, i)).toFixed(2) + ' ' + sizes[i]
}

// 格式化日期
const formatDate = (timestamp) => {
  if (!timestamp) return '-'
  const date = new Date(timestamp * 1000)
  return date.toLocaleString('zh-CN')
}

// 获取待审时长
const getPendingDuration = (uploadTime) => {
  if (!uploadTime) return '-'
  const now = Date.now()
  const upload = uploadTime * 1000
  const diff = now - upload
  
  const hours = Math.floor(diff / (1000 * 60 * 60))
  if (hours < 24) return `${hours}小时`
  
  const days = Math.floor(hours / 24)
  return `${days}天`
}

// 获取文件类型名称
const getFileTypeName = (type) => {
  const types = { 1: '视频', 2: '音频', 3: '图片', 4: '富媒体' }
  return types[type] || '未知'
}

// 获取文件类型颜色
const getFileTypeColor = (type) => {
  const colors = { 1: 'primary', 2: 'success', 3: 'warning', 4: 'info' }
  return colors[type] || ''
}

// 获取空状态描述
const getEmptyDescription = () => {
  const tabTexts = {
    pending: '暂无待审批文件',
    approved: '暂无已通过文件',
    rejected: '暂无已拒绝文件'
  }
  return tabTexts[activeTab.value] || '暂无数据'
}

// 检查是否有筛选条件
const hasFilters = () => {
  return !!(filters.type || filters.category_id || filters.uploader || filters.dateRange)
}

// 初始化
onMounted(async () => {
  await Promise.all([
    fetchFiles(),
    fetchStats(),
    fetchCategories()
  ])
})
</script>

<style scoped>
.file-approval-container {
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

.stats-row {
  margin-bottom: 20px;
  padding: 20px;
  background: #f5f7fa;
  border-radius: 8px;
}

.stat-suffix {
  font-size: 14px;
  color: #909399;
}

.file-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 60px;
  height: 60px;
  background: #f5f7fa;
  border-radius: 4px;
  color: #909399;
}

.batch-actions {
  margin-top: 20px;
}

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

.el-button-group .el-button {
  padding: 5px 10px;
}
</style>

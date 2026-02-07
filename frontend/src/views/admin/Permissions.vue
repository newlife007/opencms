<template>
  <div class="permissions-management">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ t("admin.permissions.title") }}</span>
          <el-button type="primary" icon="Refresh" @click="loadPermissions">{{ t("common.refresh") }}</el-button>
        </div>
      </template>

      <!-- Filter Form -->
      <el-form :inline="true" :model="filters" class="filter-form">
        <el-form-item label="权限名称">
          <el-input 
            v-model="filters.name" 
            placeholder="请输入权限名称" 
            clearable 
            @keyup.enter="applyFilters" 
            style="width: 240px"
          />
        </el-form-item>
        <el-form-item label="模块">
          <el-select 
            v-model="filters.module" 
            placeholder="全部" 
            clearable
            style="width: 180px"
          >
            <el-option label="文件管理 (files)" value="files" />
            <el-option label="用户管理 (users)" value="users" />
            <el-option label="组管理 (groups)" value="groups" />
            <el-option label="角色管理 (roles)" value="roles" />
            <el-option label="权限管理 (permissions)" value="permissions" />
            <el-option label="分类管理 (categories)" value="categories" />
            <el-option label="属性配置 (catalog)" value="catalog" />
            <el-option label="级别管理 (levels)" value="levels" />
            <el-option label="搜索 (search)" value="search" />
            <el-option label="转码管理 (transcoding)" value="transcoding" />
            <el-option label="系统管理 (system)" value="system" />
            <el-option label="报表统计 (reports)" value="reports" />
            <el-option label="个人中心 (profile)" value="profile" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" icon="Search" @click="applyFilters">查询</el-button>
          <el-button icon="Refresh" @click="resetFilters">{{ t("common.reset") }}</el-button>
        </el-form-item>
      </el-form>

      <!-- Permissions Table -->
      <el-table :data="paginatedPermissions" v-loading="loading" style="width: 100%" border>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="name" label="权限名称" width="200">
          <template #default="{ row }">
            <el-tag type="primary">{{ row.name }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="description" :label="t('common.description')" min-width="250" />
        <el-table-column prop="module" label="所属模块" width="120">
          <template #default="{ row }">
            <el-tag :type="getModuleTagType(row.module)">{{ row.module }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('common.actions')" width="150" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" icon="View" @click="handleView(row)">{{ t("common.view") }}</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- Pagination -->
      <div class="pagination">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="filteredPermissions.length"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>

    <!-- View Dialog -->
    <el-dialog v-model="dialogVisible" title="权限详情" width="600px">
      <el-descriptions :column="1" border v-if="currentPermission">
        <el-descriptions-item label="ID">{{ currentPermission.id }}</el-descriptions-item>
        <el-descriptions-item label="权限名称">
          <el-tag type="primary">{{ currentPermission.name }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item :label="t('common.description')">{{ currentPermission.description }}</el-descriptions-item>
        <el-descriptions-item label="所属模块">
          <el-tag :type="getModuleTagType(currentPermission.module)">{{ currentPermission.module }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatDate(currentPermission.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="更新时间">{{ formatDate(currentPermission.updated_at) }}</el-descriptions-item>
      </el-descriptions>
      <template #footer>
        <el-button @click="dialogVisible = false">{{ t("common.close") }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import permissionsApi from '@/api/permissions'


const { t } = useI18n()

const loading = ref(false)
const permissions = ref([])
const filters = ref({
  name: '',
  module: ''
})

const currentPage = ref(1)
const pageSize = ref(20)
const dialogVisible = ref(false)
const currentPermission = ref(null)

// Filtered permissions
const filteredPermissions = computed(() => {
  let result = permissions.value
  
  if (filters.value.name) {
    result = result.filter(p => 
      p.name.toLowerCase().includes(filters.value.name.toLowerCase()) ||
      p.description.toLowerCase().includes(filters.value.name.toLowerCase())
    )
  }
  
  if (filters.value.module) {
    result = result.filter(p => p.module === filters.value.module)
  }
  
  return result
})

// Paginated permissions
const paginatedPermissions = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  const end = start + pageSize.value
  return filteredPermissions.value.slice(start, end)
})

const loadPermissions = async () => {
  loading.value = true
  try {
    const res = await permissionsApi.getList()
    permissions.value = res.data || []
    ElMessage.success('权限列表加载成功')
  } catch (error) {
    console.error('加载权限列表失败:', error)
    ElMessage.error('加载权限列表失败: ' + (error.message || '未知错误'))
  } finally {
    loading.value = false
  }
}

const applyFilters = () => {
  currentPage.value = 1
}

const resetFilters = () => {
  filters.value = {
    name: '',
    module: ''
  }
  currentPage.value = 1
}

const handleView = (row) => {
  currentPermission.value = row
  dialogVisible.value = true
}

const handleSizeChange = (val) => {
  pageSize.value = val
  currentPage.value = 1
}

const handleCurrentChange = (val) => {
  currentPage.value = val
}

const getModuleTagType = (module) => {
  const types = {
    'files': 'success',
    'users': 'primary',
    'groups': 'warning',
    'roles': 'danger',
    'permissions': 'info',
    'categories': 'warning',
    'catalog': 'info',
    'levels': '',
    'search': 'success',
    'transcoding': 'primary',
    'system': 'danger',
    'reports': 'warning',
    'profile': ''
  }
  return types[module] || ''
}

const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

onMounted(() => {
  loadPermissions()
})
</script>

<style scoped>
.permissions-management {
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

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}
</style>

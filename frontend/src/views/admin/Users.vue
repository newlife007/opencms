<template>
  <div class="users-management">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ t("admin.users.title") }}</span>
          <el-button type="primary" icon="Plus" @click="handleAdd">添加用户</el-button>
        </div>
      </template>

      <!-- Filter Form -->
      <el-form :inline="true" :model="filters" class="filter-form">
        <el-form-item label="用户名">
          <el-input v-model="filters.username" :placeholder="t('admin.users.usernamePlaceholder')" clearable />
        </el-form-item>
        <el-form-item :label="t('files.status')">
          <el-select v-model="filters.enabled" placeholder="全部" clearable>
            <el-option label="启用" :value="true" />
            <el-option label="禁用" :value="false" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" icon="Search" @click="loadUsers">查询</el-button>
          <el-button icon="Refresh" @click="resetFilters">{{ t("common.reset") }}</el-button>
        </el-form-item>
      </el-form>

      <!-- Users Table -->
      <el-table :data="users" v-loading="loading" style="width: 100%">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="username" label="用户名" width="120" />
        <el-table-column prop="nickname" label="昵称" width="150" />
        <el-table-column prop="email" label="邮箱" width="200" />
        <el-table-column prop="group_id" label="所属组" width="150">
          <template #default="{ row }">
            <el-tag>{{ getGroupName(row.group_id) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="level_id" label="等级" width="100">
          <template #default="{ row }">
            <el-tag type="success">{{ getLevelName(row.level_id) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="enabled" :label="t('files.status')" width="100">
          <template #default="{ row }">
            <el-switch
              v-model="row.enabled"
              :active-value="true"
              :inactive-value="false"
              @change="handleStatusChange(row)"
            />
          </template>
        </el-table-column>
        <el-table-column prop="last_login_at" label="最后登录" width="180">
          <template #default="{ row }">
            {{ formatDate(row.last_login_at) }}
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('common.actions')" width="250" fixed="right">
          <template #default="{ row }">
            <el-button size="small" type="primary" @click="handleEdit(row)">{{ t("common.edit") }}</el-button>
            <el-button size="small" type="warning" @click="handleResetPassword(row)">{{ t("admin.users.resetPassword") }}</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row.id)">{{ t("common.delete") }}</el-button>
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
        @size-change="loadUsers"
        @current-change="loadUsers"
        style="margin-top: 20px; justify-content: flex-end"
      />
    </el-card>

    <!-- Add/Edit User Dialog -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="600px"
      @close="handleDialogClose"
    >
      <el-form
        ref="userFormRef"
        :model="userForm"
        :rules="userRules"
        label-width="100px"
      >
        <el-form-item label="用户名" prop="username">
          <el-input v-model="userForm.username" :disabled="isEdit" :placeholder="t('admin.users.usernamePlaceholder')" />
        </el-form-item>
        <el-form-item label="密码" prop="password" v-if="!isEdit">
          <el-input v-model="userForm.password" type="password" show-password placeholder="请输入密码（至少6位）" />
        </el-form-item>
        <el-form-item label="昵称" prop="nickname">
          <el-input v-model="userForm.nickname" placeholder="请输入昵称（用于显示）" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="userForm.email" :placeholder="t('admin.users.emailPlaceholder')" />
        </el-form-item>
        <el-form-item label="所属组" prop="group_id">
          <el-select v-model="userForm.group_id" placeholder="请选择组">
            <el-option
              v-for="group in groups"
              :key="group.id"
              :label="group.name"
              :value="group.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="等级" prop="level_id">
          <el-select v-model="userForm.level_id" placeholder="请选择等级">
            <el-option
              v-for="level in levels"
              :key="level.id"
              :label="level.name"
              :value="level.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('files.status')">
          <el-switch v-model="userForm.enabled" :active-value="true" :inactive-value="false" />
          <span style="margin-left: 10px; color: #999; font-size: 12px;">
            {{ userForm.enabled ? '启用' : '禁用' }}
          </span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">{{ t("common.cancel") }}</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">{{ t("common.confirm") }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import * as usersApi from '@/api/users'
import groupsApi from '@/api/groups'
import levelsApi from '@/api/levels'


const { t } = useI18n()

const loading = ref(false)
const submitting = ref(false)
const dialogVisible = ref(false)
const isEdit = ref(false)
const users = ref([])
const groups = ref([])
const levels = ref([])
const userFormRef = ref(null)

const filters = reactive({
  username: '',
  enabled: null,
})

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0,
})

const userForm = reactive({
  id: null,
  username: '',
  password: '',
  nickname: '',
  email: '',
  group_id: null,
  level_id: null,
  enabled: true,
})

const userRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 50, message: '用户名长度在 3 到 50 个字符', trigger: 'blur' },
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码长度至少 6 个字符', trigger: 'blur' },
  ],
  nickname: [
    { required: true, message: '请输入昵称', trigger: 'blur' },
    { max: 64, message: '昵称长度不能超过 64 个字符', trigger: 'blur' },
  ],
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入正确的邮箱地址', trigger: 'blur' },
  ],
  group_id: [{ required: true, message: '请选择所属组', trigger: 'change' }],
  level_id: [{ required: true, message: '请选择等级', trigger: 'change' }],
}

const dialogTitle = computed(() => (isEdit.value ? '编辑用户' : '添加用户'))

const getGroupName = (groupId) => {
  const group = groups.value.find(g => g.id === groupId)
  return group ? group.name : '未知'
}

const getLevelName = (levelId) => {
  const level = levels.value.find(l => l.id === levelId)
  return level ? level.name : `Level ${levelId}`
}

const formatDate = (timestamp) => {
  if (!timestamp) return '-'
  return new Date(timestamp * 1000).toLocaleString('zh-CN')
}

const loadUsers = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      pageSize: pagination.pageSize,
      username: filters.username || undefined,
      enabled: filters.enabled !== null ? filters.enabled : undefined,
    }
    
    console.log('Loading users with params:', params)
    const res = await usersApi.getUserList(params)
    console.log('Users response:', res)
    
    if (res.success) {
      users.value = res.data || []
      pagination.total = res.pagination?.total || 0
      ElMessage.success(`加载成功，共 ${pagination.total} 个用户`)
    } else {
      ElMessage.error(res.message || '加载用户列表失败')
      users.value = []
      pagination.total = 0
    }
  } catch (error) {
    console.error('Load users error:', error)
    ElMessage.error('加载用户列表失败: ' + (error.message || '网络错误'))
    users.value = []
    pagination.total = 0
  } finally {
    loading.value = false
  }
}

const loadGroups = async () => {
  try {
    console.log('Loading groups...')
    const res = await groupsApi.getList({ page: 1, pageSize: 100 })
    console.log('Groups response:', res)
    
    if (res.success) {
      groups.value = res.data || []
    } else {
      ElMessage.error('加载组列表失败')
      groups.value = []
    }
  } catch (error) {
    console.error('Load groups error:', error)
    ElMessage.error('加载组列表失败: ' + (error.message || '网络错误'))
    groups.value = []
  }
}

const loadLevels = async () => {
  try {
    console.log('Loading levels...')
    const res = await levelsApi.getList({ page: 1, pageSize: 100 })
    console.log('Levels response:', res)
    
    if (res.success) {
      levels.value = res.data || []
    } else {
      ElMessage.error('加载等级列表失败')
      levels.value = []
    }
  } catch (error) {
    console.error('Load levels error:', error)
    ElMessage.error('加载等级列表失败: ' + (error.message || '网络错误'))
    levels.value = []
  }
}

const resetFilters = () => {
  filters.username = ''
  filters.enabled = null
  pagination.page = 1
  loadUsers()
}

const handleAdd = () => {
  isEdit.value = false
  Object.assign(userForm, {
    id: null,
    username: '',
    password: '',
    nickname: '',
    email: '',
    group_id: null,
    level_id: null,
    enabled: true,
  })
  loadGroups()
  loadLevels()
  dialogVisible.value = true
}

const handleEdit = (row) => {
  isEdit.value = true
  Object.assign(userForm, { ...row })
  loadGroups()
  loadLevels()
  dialogVisible.value = true
}

const handleDialogClose = () => {
  userFormRef.value?.resetFields()
}

const handleSubmit = async () => {
  if (!userFormRef.value) return

  await userFormRef.value.validate(async (valid) => {
    if (!valid) return

    submitting.value = true
    try {
      const submitData = {
        username: userForm.username,
        nickname: userForm.nickname,
        email: userForm.email,
        group_id: userForm.group_id,
        level_id: userForm.level_id,
        enabled: userForm.enabled,
      }
      
      // Add password for new users
      if (!isEdit.value) {
        submitData.password = userForm.password
      }
      
      console.log('Submitting user:', submitData)
      
      let res
      if (isEdit.value) {
        res = await usersApi.updateUser(userForm.id, submitData)
      } else {
        res = await usersApi.createUser(submitData)
      }
      
      console.log('Submit response:', res)
      
      if (res.success) {
        ElMessage.success(isEdit.value ? '更新成功' : '添加成功')
        dialogVisible.value = false
        loadUsers()
      } else {
        ElMessage.error(res.message || (isEdit.value ? '更新失败' : '添加失败'))
      }
    } catch (error) {
      console.error('Submit error:', error)
      ElMessage.error((isEdit.value ? '更新失败: ' : '添加失败: ') + (error.message || '网络错误'))
    } finally {
      submitting.value = false
    }
  })
}

const handleStatusChange = async (row) => {
  try {
    console.log('Updating enabled for user:', row.id, 'to', row.enabled)
    const res = await usersApi.updateUserStatus(row.id, row.enabled)
    console.log('Status update response:', res)
    
    if (res.success) {
      ElMessage.success('状态更新成功')
    } else {
      ElMessage.error(res.message || '状态更新失败')
      // Revert enabled
      row.enabled = !row.enabled
    }
  } catch (error) {
    console.error('Status change error:', error)
    ElMessage.error('状态更新失败: ' + (error.message || '网络错误'))
    // Revert enabled
    row.enabled = !row.enabled
  }
}

const handleResetPassword = async (row) => {
  try {
    const result = await ElMessageBox.prompt('请输入新密码', '重置密码', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      inputType: 'password',
      inputValidator: (value) => {
        if (!value || value.length < 6) {
          return '密码长度至少 6 个字符'
        }
        return true
      },
    })
    
    console.log('Resetting password for user:', row.id)
    const res = await usersApi.resetUserPassword(row.id, { new_password: result.value })
    console.log('Reset password response:', res)
    
    if (res.success) {
      ElMessage.success('密码重置成功')
    } else {
      ElMessage.error(res.message || '密码重置失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Reset password error:', error)
      ElMessage.error('密码重置失败: ' + (error.message || '网络错误'))
    }
  }
}

const handleDelete = async (id) => {
  try {
    await ElMessageBox.confirm('确定要删除该用户吗？删除后无法恢复！', '警告', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'error',
    })
    
    console.log('Deleting user:', id)
    const res = await usersApi.deleteUser(id)
    console.log('Delete response:', res)
    
    if (res.success) {
      ElMessage.success(t('message.deleteSuccess'))
      loadUsers()
    } else {
      ElMessage.error(res.message || '删除失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Delete error:', error)
      ElMessage.error('删除失败: ' + (error.message || '网络错误'))
    }
  }
}

// Initialize on mount
onMounted(() => {
  loadUsers()
})
</script>

<style scoped>
.users-management {
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
</style>

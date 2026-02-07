<template>
  <div class="groups-management">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ t("admin.groups.title") }}</span>
          <el-button type="primary" icon="Plus" @click="handleAdd">添加组</el-button>
        </div>
      </template>

      <!-- Groups Table -->
      <el-table :data="groups" v-loading="loading" style="width: 100%">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="name" label="组名称" width="200" />
        <el-table-column prop="description" :label="t('common.description')" />
        <el-table-column prop="user_count" label="用户数" width="100" />
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('common.actions')" width="300" fixed="right">
          <template #default="{ row }">
            <el-button size="small" type="primary" @click="handleEdit(row)">{{ t("common.edit") }}</el-button>
            <el-button size="small" type="success" @click="handleAssignCategories(row)">
              分配分类
            </el-button>
            <el-button size="small" type="warning" @click="handleAssignRoles(row)">{{ t("admin.groups.assignRole") }}</el-button>
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
        @size-change="loadGroups"
        @current-change="loadGroups"
        style="margin-top: 20px; justify-content: flex-end"
      />
    </el-card>

    <!-- Add/Edit Group Dialog -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="600px"
      @close="handleDialogClose"
    >
      <el-form
        ref="groupFormRef"
        :model="groupForm"
        :rules="groupRules"
        label-width="100px"
      >
        <el-form-item label="组名称" prop="name">
          <el-input v-model="groupForm.name" placeholder="请输入组名称" />
        </el-form-item>
        <el-form-item :label="t('common.description')" prop="description">
          <el-input
            v-model="groupForm.description"
            type="textarea"
            :rows="4"
            placeholder="请输入组描述"
          />
        </el-form-item>
        <el-form-item v-if="!isEdit" label="关联角色" prop="role_ids">
          <el-select
            v-model="groupForm.role_ids"
            multiple
            placeholder="请选择角色"
            style="width: 100%"
          >
            <el-option
              v-for="role in allRoles"
              :key="role.id"
              :label="role.name"
              :value="role.id"
            >
              <span>{{ role.name }}</span>
              <span style="color: #8492a6; font-size: 13px; margin-left: 10px">
                {{ role.description }}
              </span>
            </el-option>
          </el-select>
          <div style="color: #909399; font-size: 12px; margin-top: 4px">
            默认关联"查看者"角色，新建组的成员将拥有文件查看、搜索和下载权限
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">{{ t("common.cancel") }}</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">{{ t("common.confirm") }}</el-button>
      </template>
    </el-dialog>

    <!-- Assign Categories Dialog -->
    <el-dialog v-model="categoriesDialogVisible" title="分配分类" width="600px">
      <el-tree
        ref="categoriesTreeRef"
        :data="categoryTree"
        show-checkbox
        node-key="id"
        :props="{ label: 'name', children: 'children' }"
        :default-checked-keys="selectedCategories"
      />
      <template #footer>
        <el-button @click="categoriesDialogVisible = false">{{ t("common.cancel") }}</el-button>
        <el-button type="primary" @click="handleSubmitCategories" :loading="submitting">{{ t("common.confirm") }}</el-button>
      </template>
    </el-dialog>

    <!-- Assign Roles Dialog -->
    <el-dialog v-model="rolesDialogVisible" title="分配角色" width="600px">
      <el-checkbox-group v-model="selectedRoles">
        <el-checkbox
          v-for="role in allRoles"
          :key="role.id"
          :label="role.id"
          style="display: block; margin-bottom: 10px"
        >
          {{ role.name }} - {{ role.description }}
        </el-checkbox>
      </el-checkbox-group>
      <template #footer>
        <el-button @click="rolesDialogVisible = false">{{ t("common.cancel") }}</el-button>
        <el-button type="primary" @click="handleSubmitRoles" :loading="submitting">{{ t("common.confirm") }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import groupsApi from '@/api/groups'
import categoryApi from '@/api/category'
import rolesApi from '@/api/roles'


const { t } = useI18n()

const VIEWER_ROLE_ID = 5  // 查看者角色ID

const loading = ref(false)
const submitting = ref(false)
const dialogVisible = ref(false)
const categoriesDialogVisible = ref(false)
const rolesDialogVisible = ref(false)
const isEdit = ref(false)
const groups = ref([])
const categoryTree = ref([])
const allRoles = ref([])
const selectedCategories = ref([])
const selectedRoles = ref([])
const currentGroupId = ref(null)

const groupFormRef = ref(null)
const categoriesTreeRef = ref(null)

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0,
})

const groupForm = reactive({
  id: null,
  name: '',
  description: '',
  role_ids: [VIEWER_ROLE_ID],  // 默认选中查看者角色
})

const groupRules = {
  name: [
    { required: true, message: '请输入组名称', trigger: 'blur' },
    { min: 2, max: 50, message: '组名称长度在 2 到 50 个字符', trigger: 'blur' },
  ],
}

const dialogTitle = computed(() => (isEdit.value ? '编辑组' : '添加组'))

const formatDate = (timestamp) => {
  if (!timestamp) return '-'
  return new Date(timestamp * 1000).toLocaleString('zh-CN')
}

const loadGroups = async () => {
  loading.value = true
  try {
    const res = await groupsApi.getList({
      page: pagination.page,
      page_size: pagination.pageSize,
    })
    if (res.success) {
      // Backend returns: { success: true, data: [...], pagination: {...} }
      groups.value = res.data || []
      pagination.total = res.pagination?.total || 0
    }
  } catch (error) {
    console.error('Failed to load groups:', error)
    ElMessage.error('加载组列表失败')
  } finally {
    loading.value = false
  }
}

const loadCategoryTree = async () => {
  try {
    const res = await categoryApi.getTree()
    if (res.success) {
      categoryTree.value = res.data || []
    }
  } catch (error) {
    console.error('Load category tree failed:', error)
  }
}

const loadAllRoles = async () => {
  try {
    const res = await rolesApi.getList({ page_size: 1000 })
    if (res.success) {
      // Backend returns: { success: true, data: [...], pagination: {...} }
      allRoles.value = res.data || []
    }
  } catch (error) {
    console.error('Load roles failed:', error)
  }
}

const handleAdd = async () => {
  isEdit.value = false
  
  // 加载角色列表
  await loadAllRoles()
  
  Object.assign(groupForm, {
    id: null,
    name: '',
    description: '',
    role_ids: [VIEWER_ROLE_ID],  // 默认查看者角色
  })
  dialogVisible.value = true
}

const handleEdit = (row) => {
  isEdit.value = true
  Object.assign(groupForm, { ...row })
  dialogVisible.value = true
}

const handleDialogClose = () => {
  groupFormRef.value?.resetFields()
}

const handleSubmit = async () => {
  if (!groupFormRef.value) return

  await groupFormRef.value.validate(async (valid) => {
    if (!valid) return

    submitting.value = true
    try {
      if (isEdit.value) {
        await groupsApi.update(groupForm.id, groupForm)
        ElMessage.success('更新成功')
      } else {
        // 创建组并关联角色
        const res = await groupsApi.create({
          name: groupForm.name,
          description: groupForm.description,
          enabled: true
        })
        
        const groupId = res.data.id
        
        // 自动关联选择的角色
        if (groupForm.role_ids && groupForm.role_ids.length > 0) {
          await groupsApi.assignRoles(groupId, groupForm.role_ids)
        }
        
        ElMessage.success('添加成功，已关联角色')
      }
      dialogVisible.value = false
      loadGroups()
    } catch (error) {
      console.error('Submit error:', error)
      ElMessage.error(isEdit.value ? '更新失败' : '添加失败')
    } finally {
      submitting.value = false
    }
  })
}

const handleAssignCategories = async (row) => {
  currentGroupId.value = row.id
  
  // Load assigned categories
  try {
    const res = await groupsApi.getDetail(row.id)
    // Backend returns categories at top level, not in res.data
    if (res.success && res.categories) {
      selectedCategories.value = res.categories.map(c => c.id)
    } else {
      selectedCategories.value = []
    }
  } catch (error) {
    console.error('Failed to load assigned categories:', error)
    selectedCategories.value = []
  }
  
  await loadCategoryTree()
  categoriesDialogVisible.value = true
}

const handleSubmitCategories = async () => {
  submitting.value = true
  try {
    const checkedKeys = categoriesTreeRef.value.getCheckedKeys()
    await groupsApi.assignCategories(currentGroupId.value, checkedKeys)
    ElMessage.success('分配分类成功')
    categoriesDialogVisible.value = false
  } catch (error) {
    ElMessage.error('分配分类失败')
  } finally {
    submitting.value = false
  }
}

const handleAssignRoles = async (row) => {
  currentGroupId.value = row.id
  
  // Load assigned roles
  try {
    const res = await groupsApi.getDetail(row.id)
    // Backend returns roles at top level, not in res.data
    if (res.success && res.roles) {
      selectedRoles.value = res.roles.map(r => r.id)
    } else {
      selectedRoles.value = []
    }
  } catch (error) {
    console.error('Failed to load assigned roles:', error)
    selectedRoles.value = []
  }
  
  await loadAllRoles()
  rolesDialogVisible.value = true
}

const handleSubmitRoles = async () => {
  submitting.value = true
  try {
    await groupsApi.assignRoles(currentGroupId.value, selectedRoles.value)
    ElMessage.success('分配角色成功')
    rolesDialogVisible.value = false
  } catch (error) {
    ElMessage.error('分配角色失败')
  } finally {
    submitting.value = false
  }
}

const handleDelete = async (id) => {
  try {
    await ElMessageBox.confirm('确定要删除该组吗？删除后无法恢复！', '警告', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'error',
    })
    
    await groupsApi.delete(id)
    ElMessage.success(t('message.deleteSuccess'))
    loadGroups()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(t('message.deleteFailed'))
    }
  }
}

onMounted(() => {
  loadGroups()
})
</script>

<style scoped>
.groups-management {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>

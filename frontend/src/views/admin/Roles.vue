<template>
  <div class="roles-management">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ t("admin.roles.title") }}</span>
          <el-button type="primary" icon="Plus" @click="handleAdd">添加角色</el-button>
        </div>
      </template>

      <!-- Roles Table -->
      <el-table :data="roles" v-loading="loading" style="width: 100%">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="name" label="角色名称" width="200" />
        <el-table-column prop="description" :label="t('common.description')" />
        <el-table-column label="类型" width="120">
          <template #default="{ row }">
            <el-tag v-if="row.is_system" type="warning" size="small">
              系统角色
            </el-tag>
            <el-tag v-else type="info" size="small">
              自定义
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="permission_count" label="权限数" width="100" />
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column :label="t('common.actions')" width="250" fixed="right">
          <template #default="{ row }">
            <el-button size="small" type="primary" @click="handleEdit(row)">{{ t("common.edit") }}</el-button>
            <el-button size="small" type="success" @click="handleAssignPermissions(row)">
              分配权限
            </el-button>
            <!-- 只有自定义角色才显示删除按钮 -->
            <el-button
              v-if="!row.is_system"
              size="small"
              type="danger"
              @click="handleDelete(row)"
            >{{ t("common.delete") }}</el-button>
            <!-- 系统角色显示禁用的删除按钮 -->
            <el-tooltip
              v-else
              content="系统角色不可删除"
              placement="top"
            >
              <el-button
                size="small"
                type="danger"
                disabled
              >{{ t("common.delete") }}</el-button>
            </el-tooltip>
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
        @size-change="loadRoles"
        @current-change="loadRoles"
        style="margin-top: 20px; justify-content: flex-end"
      />
    </el-card>

    <!-- Add/Edit Role Dialog -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="600px"
      @close="handleDialogClose"
    >
      <el-form
        ref="roleFormRef"
        :model="roleForm"
        :rules="roleRules"
        label-width="100px"
      >
        <el-form-item label="角色名称" prop="name">
          <el-input v-model="roleForm.name" placeholder="请输入角色名称" />
        </el-form-item>
        <el-form-item :label="t('common.description')" prop="description">
          <el-input
            v-model="roleForm.description"
            type="textarea"
            :rows="4"
            placeholder="请输入角色描述"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">{{ t("common.cancel") }}</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">{{ t("common.confirm") }}</el-button>
      </template>
    </el-dialog>

    <!-- Assign Permissions Dialog -->
    <el-dialog v-model="permissionsDialogVisible" title="分配权限" width="800px">
      <div style="margin-bottom: 15px;">
        <el-alert 
          title="提示：勾选模块可选择该模块下所有权限，展开模块可精细选择" 
          type="info" 
          :closable="false"
        />
      </div>
      
      <div style="margin-bottom: 15px;">
        <el-tag type="success">已选择 {{ selectedPermissions.length }} 个权限</el-tag>
        <el-button 
          size="small" 
          style="margin-left: 10px" 
          @click="expandAll"
        >
          展开全部
        </el-button>
        <el-button 
          size="small" 
          @click="collapseAll"
        >
          收起全部
        </el-button>
      </div>

      <!-- 权限树 -->
      <el-tree
        ref="permissionTreeRef"
        :data="permissionTree"
        show-checkbox
        node-key="id"
        :props="treeProps"
        :default-checked-keys="selectedPermissions"
        :default-expand-all="false"
        @check="handleTreeCheck"
        style="border: 1px solid #dcdfe6; border-radius: 4px; padding: 10px; max-height: 500px; overflow-y: auto;"
      >
        <template #default="{ node, data }">
          <span class="custom-tree-node">
            <span class="tree-node-label">
              <el-icon v-if="data.isModule" style="margin-right: 5px;">
                <Folder />
              </el-icon>
              <el-icon v-else style="margin-right: 5px; color: #409eff;">
                <DocumentChecked />
              </el-icon>
              {{ data.label }}
            </span>
            <span v-if="data.isModule" class="tree-node-count">
              ({{ data.children?.length || 0 }})
            </span>
            <el-tag v-else-if="data.rbac" :type="getRbacTagType(data.rbac)" size="small" style="margin-left: 8px;">
              {{ data.rbac }}
            </el-tag>
          </span>
        </template>
      </el-tree>
      
      <template #footer>
        <el-button @click="permissionsDialogVisible = false">{{ t("common.cancel") }}</el-button>
        <el-button type="primary" @click="handleSubmitPermissions" :loading="submitting">
          确定 (已选 {{ selectedPermissions.length }} 个)
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Folder, DocumentChecked } from '@element-plus/icons-vue'
import rolesApi from '@/api/roles'


const { t } = useI18n()

const loading = ref(false)
const submitting = ref(false)
const dialogVisible = ref(false)
const permissionsDialogVisible = ref(false)
const isEdit = ref(false)
const roles = ref([])
const allPermissions = ref([])
const selectedPermissions = ref([])
const currentRoleId = ref(null)

const roleFormRef = ref(null)
const permissionTreeRef = ref(null)

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0,
})

const roleForm = reactive({
  id: null,
  name: '',
  description: '',
})

const roleRules = {
  name: [
    { required: true, message: '请输入角色名称', trigger: 'blur' },
    { min: 2, max: 50, message: '角色名称长度在 2 到 50 个字符', trigger: 'blur' },
  ],
}

const dialogTitle = computed(() => (isEdit.value ? '编辑角色' : '添加角色'))

// Tree props for el-tree
const treeProps = {
  children: 'children',
  label: 'label'
}

// 构建权限树结构
const permissionTree = computed(() => {
  const tree = []
  const moduleMap = {}
  
  // 按模块分组权限
  allPermissions.value.forEach(permission => {
    const module = permission.module || permission.namespace || 'other'
    
    if (!moduleMap[module]) {
      moduleMap[module] = {
        id: `module_${module}`,
        label: getModuleLabel(module),
        isModule: true,
        children: []
      }
    }
    
    moduleMap[module].children.push({
      id: permission.id,
      label: permission.description || permission.name,
      name: permission.name,
      rbac: permission.rbac,
      isModule: false,
      permission: permission
    })
  })
  
  // 转换为数组并排序
  Object.keys(moduleMap).sort().forEach(module => {
    tree.push(moduleMap[module])
  })
  
  return tree
})

const formatDate = (timestamp) => {
  if (!timestamp) return '-'
  return new Date(timestamp * 1000).toLocaleString('zh-CN')
}

// 树节点勾选事件
const handleTreeCheck = () => {
  // Element Plus Tree 会自动处理父子关联
  // 模块全选 → 模块显示对号
  // 模块部分选 → 模块显示半选（蓝色填充）
  updateSelectedPermissions()
}

// 更新已选权限列表
const updateSelectedPermissions = () => {
  const tree = permissionTreeRef.value
  if (!tree) return
  
  // 获取所有选中的节点（包括半选）
  const checkedNodes = tree.getCheckedNodes()
  const halfCheckedNodes = tree.getHalfCheckedNodes()
  
  // 只获取权限节点（非模块节点）
  // 包括完全选中的和半选父节点下的权限
  selectedPermissions.value = checkedNodes
    .filter(node => !node.isModule)
    .map(node => node.id)
}

// 展开全部
const expandAll = () => {
  const tree = permissionTreeRef.value
  if (!tree) return
  
  permissionTree.value.forEach(node => {
    tree.store.nodesMap[node.id].expanded = true
  })
}

// 收起全部
const collapseAll = () => {
  const tree = permissionTreeRef.value
  if (!tree) return
  
  permissionTree.value.forEach(node => {
    tree.store.nodesMap[node.id].expanded = false
  })
}

const getModuleLabel = (module) => {
  const labels = {
    'files': '文件管理',
    'users': '用户管理',
    'groups': '组管理',
    'roles': '角色管理',
    'permissions': '权限管理',
    'categories': '分类管理',
    'catalog': '属性配置',
    'levels': '级别管理',
    'search': '搜索',
    'transcoding': '转码管理',
    'system': '系统管理',
    'reports': '报表统计',
    'profile': '个人中心'
  }
  return labels[module] || module
}

const getRbacTagType = (rbac) => {
  const types = {
    'ACL_ALL': '',
    'ACL_EDIT': 'warning',
    'ACL_CATALOG': 'success',
    'ACL_PUTOUT': 'primary',
    'ACL_ADMIN': 'danger'
  }
  return types[rbac] || ''
}

const loadRoles = async () => {
  loading.value = true
  try {
    const res = await rolesApi.getList({
      page: pagination.page,
      page_size: pagination.pageSize,
    })
    if (res.success) {
      // Backend returns: { success: true, data: [...], pagination: {...} }
      roles.value = res.data || []
      pagination.total = res.pagination?.total || 0
    }
  } catch (error) {
    console.error('Failed to load roles:', error)
    ElMessage.error('加载角色列表失败')
  } finally {
    loading.value = false
  }
}

const loadAllPermissions = async () => {
  try {
    const res = await rolesApi.getAllPermissions()
    if (res.success) {
      // Backend returns: { success: true, data: [...] }
      allPermissions.value = res.data || []
    }
  } catch (error) {
    console.error('Load permissions failed:', error)
  }
}

const handleAdd = () => {
  isEdit.value = false
  Object.assign(roleForm, {
    id: null,
    name: '',
    description: '',
  })
  dialogVisible.value = true
}

const handleEdit = (row) => {
  isEdit.value = true
  Object.assign(roleForm, { ...row })
  dialogVisible.value = true
}

const handleDialogClose = () => {
  roleFormRef.value?.resetFields()
}

const handleSubmit = async () => {
  if (!roleFormRef.value) return

  await roleFormRef.value.validate(async (valid) => {
    if (!valid) return

    submitting.value = true
    try {
      if (isEdit.value) {
        await rolesApi.update(roleForm.id, roleForm)
        ElMessage.success('更新成功')
      } else {
        await rolesApi.create(roleForm)
        ElMessage.success('添加成功')
      }
      dialogVisible.value = false
      loadRoles()
    } catch (error) {
      ElMessage.error(isEdit.value ? '更新失败' : '添加失败')
    } finally {
      submitting.value = false
    }
  })
}

const handleAssignPermissions = async (row) => {
  currentRoleId.value = row.id
  
  // Load assigned permissions
  try {
    const res = await rolesApi.getDetail(row.id)
    // Backend returns: { success: true, data: {...}, permissions: [...] }
    // Permissions are at root level, not in data
    if (res.success && res.permissions) {
      selectedPermissions.value = res.permissions.map(p => p.id)
    } else {
      selectedPermissions.value = []
    }
  } catch (error) {
    selectedPermissions.value = []
  }
  
  await loadAllPermissions()
  permissionsDialogVisible.value = true
}

const handleSubmitPermissions = async () => {
  submitting.value = true
  try {
    // 从树中获取所有勾选的权限ID（只包含权限节点，不包含模块节点）
    const tree = permissionTreeRef.value
    const checkedNodes = tree.getCheckedNodes()
    
    // 过滤出权限节点（排除模块节点）
    const permissionIds = checkedNodes
      .filter(node => !node.isModule)
      .map(node => node.id)
    
    console.log('提交权限IDs:', permissionIds)
    
    await rolesApi.assignPermissions(currentRoleId.value, permissionIds)
    ElMessage.success('分配权限成功')
    permissionsDialogVisible.value = false
    
    // 重新加载角色列表以更新权限数
    await loadRoles()
  } catch (error) {
    console.error('分配权限失败:', error)
    ElMessage.error('分配权限失败')
  } finally {
    submitting.value = false
  }
}

const handleDelete = async (row) => {
  // 双重确认（虽然后端有保护，但前端也做检查）
  if (row.is_system) {
    ElMessage.warning('系统角色不能删除')
    return
  }
  
  try {
    await ElMessageBox.confirm(
      `确定要删除角色"${row.name}"吗？删除后无法恢复！`, 
      '警告', 
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'error',
      }
    )
    
    await rolesApi.delete(row.id)
    ElMessage.success(t('message.deleteSuccess'))
    loadRoles()
  } catch (error) {
    if (error !== 'cancel') {
      const errorMsg = error.response?.data?.message || error.message || '删除失败'
      ElMessage.error(errorMsg)
    }
  }
}

onMounted(() => {
  loadRoles()
})
</script>

<style scoped>
.roles-management {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.custom-tree-node {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding-right: 10px;
}

.tree-node-label {
  display: flex;
  align-items: center;
  font-size: 14px;
}

.tree-node-count {
  color: #909399;
  font-size: 12px;
  margin-left: 8px;
}

:deep(.el-tree-node__content) {
  height: 36px;
  padding: 0 8px;
}

:deep(.el-tree-node__label) {
  font-size: 14px;
}

/* 模块节点样式 */
:deep(.el-tree-node > .el-tree-node__content) {
  font-weight: 500;
}

/* 子节点缩进 */
:deep(.el-tree-node__children .el-tree-node__content) {
  padding-left: 24px !important;
}
</style>

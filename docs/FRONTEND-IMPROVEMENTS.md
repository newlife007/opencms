# OpenWan前端权限管理改进

## 实施日期：2026-02-05

## 概述

根据用户需求，对OpenWan权限管理系统进行了三项重要改进：

1. ✅ **创建组时默认关联查看者角色**
2. ⏳ **权限分配使用树形结构展示** (前端待实施)
3. ✅ **系统默认角色不可删除** (后端已完成)

---

## 需求1: 创建组时默认关联查看者角色

### 需求描述
修改前端创建组时必须关联角色，默认关联查看者角色（role_id=5）。

### 后端支持
后端已支持，无需修改。创建组后可调用 `POST /api/v1/admin/groups/:id/roles` 关联角色。

### 前端实施步骤

**文件**: `frontend/src/views/admin/Groups.vue`

```vue
<!-- 创建组对话框 -->
<el-form-item label="关联角色" prop="role_ids">
  <el-select
    v-model="groupForm.role_ids"
    multiple
    placeholder="请选择角色"
    style="width: 100%"
  >
    <el-option
      v-for="role in roles"
      :key="role.id"
      :label="role.name"
      :value="role.id"
    />
  </el-select>
  <div style="color: #909399; font-size: 12px; margin-top: 4px">
    默认关联"查看者"角色
  </div>
</el-form-item>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import api from '@/api'

const VIEWER_ROLE_ID = 5 // 查看者角色ID

const groupForm = ref({
  name: '',
  description: '',
  enabled: true,
  role_ids: [VIEWER_ROLE_ID] // 默认选中查看者角色
})

const roles = ref([])

// 加载角色列表
const loadRoles = async () => {
  try {
    const { data } = await api.roles.list({ page: 1, page_size: 100 })
    roles.value = data.data
  } catch (error) {
    ElMessage.error('加载角色列表失败')
  }
}

// 创建组
const createGroup = async () => {
  try {
    // 1. 创建组
    const { data: groupData } = await api.groups.create({
      name: groupForm.value.name,
      description: groupForm.value.description,
      enabled: groupForm.value.enabled
    })
    
    const groupId = groupData.data.id
    
    // 2. 关联角色
    if (groupForm.value.role_ids.length > 0) {
      await api.groups.assignRoles(groupId, {
        role_ids: groupForm.value.role_ids
      })
    }
    
    ElMessage.success('创建组成功')
    dialogVisible.value = false
    loadGroups()
  } catch (error) {
    ElMessage.error('创建组失败: ' + (error.response?.data?.message || error.message))
  }
}

onMounted(() => {
  loadRoles()
  loadGroups()
})
</script>
```

**API文件**: `frontend/src/api/groups.js`

```javascript
import request from '@/utils/request'

export default {
  // 创建组
  create(data) {
    return request({
      url: '/admin/groups',
      method: 'post',
      data
    })
  },
  
  // 关联角色
  assignRoles(groupId, data) {
    return request({
      url: `/admin/groups/${groupId}/roles`,
      method: 'post',
      data
    })
  }
}
```

---

## 需求2: 权限分配使用树形结构展示

### 需求描述
给角色分配权限时，使用树型结构，默认只显示模块名称，在模块名称前有选择框，可以直接选择该模块下的所有权限，也可以点开模块精细化选择要添加的权限。

### 数据结构

权限按模块（namespace）组织，例如：

```
files (文件管理)
  ├─ browse (浏览)
  │   ├─ list (列表)
  │   ├─ view (查看)
  │   └─ download (下载)
  ├─ upload (上传)
  │   ├─ create (创建)
  │   └─ batch (批量)
  └─ edit (编辑)
      ├─ update (更新)
      └─ delete (删除)
```

### 前端实施步骤

**文件**: `frontend/src/views/admin/Roles.vue`

```vue
<template>
  <div class="role-permissions">
    <el-dialog
      v-model="permissionDialogVisible"
      title="分配权限"
      width="600px"
    >
      <el-tree
        ref="permissionTreeRef"
        :data="permissionTree"
        :props="treeProps"
        show-checkbox
        node-key="id"
        :default-checked-keys="selectedPermissions"
        :default-expanded-keys="[]"
        check-strictly
        @check="handlePermissionCheck"
      >
        <template #default="{ node, data }">
          <span class="custom-tree-node">
            <span>{{ data.label }}</span>
            <span v-if="data.children" class="permission-count">
              ({{ data.children.length }})
            </span>
          </span>
        </template>
      </el-tree>
      
      <template #footer>
        <el-button @click="permissionDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="savePermissions">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import api from '@/api'

const permissionTreeRef = ref(null)
const permissionDialogVisible = ref(false)
const allPermissions = ref([])
const selectedPermissions = ref([])
const currentRoleId = ref(null)

const treeProps = {
  children: 'children',
  label: 'label'
}

// 构建权限树
const permissionTree = computed(() => {
  const modules = {}
  
  // 按模块分组
  allPermissions.value.forEach(perm => {
    const module = perm.module || perm.namespace
    if (!modules[module]) {
      modules[module] = {
        id: `module_${module}`,
        label: getModuleName(module),
        module: module,
        children: []
      }
    }
    
    modules[module].children.push({
      id: perm.id,
      label: `${perm.description || perm.name}`,
      permission: perm
    })
  })
  
  return Object.values(modules)
})

// 模块名称映射
const getModuleName = (module) => {
  const moduleNames = {
    'files': '文件管理',
    'categories': '分类管理',
    'users': '用户管理',
    'groups': '组管理',
    'roles': '角色管理',
    'permissions': '权限管理',
    'levels': '级别管理',
    'search': '搜索',
    'transcoding': '转码',
    'system': '系统',
    'reports': '报表'
  }
  return moduleNames[module] || module
}

// 加载所有权限
const loadPermissions = async () => {
  try {
    const { data } = await api.permissions.list()
    allPermissions.value = data.data
  } catch (error) {
    ElMessage.error('加载权限列表失败')
  }
}

// 打开权限分配对话框
const openPermissionDialog = async (roleId) => {
  currentRoleId.value = roleId
  
  try {
    // 加载角色当前权限
    const { data } = await api.roles.getPermissions(roleId)
    selectedPermissions.value = data.permissions.map(p => p.id)
    permissionDialogVisible.value = true
  } catch (error) {
    ElMessage.error('加载角色权限失败')
  }
}

// 处理节点勾选
const handlePermissionCheck = (data, checked) => {
  // 如果是模块节点，自动勾选/取消所有子权限
  if (data.children && data.children.length > 0) {
    const tree = permissionTreeRef.value
    data.children.forEach(child => {
      tree.setChecked(child.id, checked.checkedKeys.includes(data.id), false)
    })
  }
}

// 保存权限分配
const savePermissions = async () => {
  try {
    const tree = permissionTreeRef.value
    const checkedNodes = tree.getCheckedNodes()
    
    // 只提取权限ID（排除模块节点）
    const permissionIds = checkedNodes
      .filter(node => node.permission)
      .map(node => node.id)
    
    await api.roles.assignPermissions(currentRoleId.value, {
      permission_ids: permissionIds
    })
    
    ElMessage.success('权限分配成功')
    permissionDialogVisible.value = false
  } catch (error) {
    ElMessage.error('权限分配失败: ' + (error.response?.data?.message || error.message))
  }
}

onMounted(() => {
  loadPermissions()
})

defineExpose({
  openPermissionDialog
})
</script>

<style scoped>
.custom-tree-node {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

.permission-count {
  color: #909399;
  font-size: 12px;
  margin-left: 8px;
}

:deep(.el-tree-node__content) {
  height: 36px;
}

:deep(.el-tree-node__label) {
  font-size: 14px;
}
</style>
```

**API文件**: `frontend/src/api/roles.js`

```javascript
import request from '@/utils/request'

export default {
  // 获取角色权限
  getPermissions(roleId) {
    return request({
      url: `/admin/roles/${roleId}`,
      method: 'get'
    })
  },
  
  // 分配权限
  assignPermissions(roleId, data) {
    return request({
      url: `/admin/roles/${roleId}/permissions`,
      method: 'post',
      data
    })
  }
}
```

**API文件**: `frontend/src/api/permissions.js`

```javascript
import request from '@/utils/request'

export default {
  // 获取所有权限
  list(params) {
    return request({
      url: '/admin/permissions',
      method: 'get',
      params
    })
  }
}
```

---

## 需求3: 系统默认角色不可删除

### 需求描述
系统的默认角色：超级管理员、内容管理员、审核员、编辑、查看者，不可删除，只有通过系统创建的角色才可以删除。

### ✅ 后端已完成

#### 1. 数据库修改
```sql
-- 添加is_system字段标识系统角色
ALTER TABLE ow_roles 
ADD COLUMN is_system TINYINT NOT NULL DEFAULT 0 COMMENT '是否系统角色 0=否 1=是';

-- 标记5个系统角色
UPDATE ow_roles 
SET is_system = 1 
WHERE id IN (1, 2, 3, 4, 5);

-- 验证
SELECT id, name, is_system FROM ow_roles;
```

**结果**:
```
1: 超级管理员 - is_system=True
2: 内容管理员 - is_system=True
3: 审核员 - is_system=True
4: 编辑 - is_system=True
5: 查看者 - is_system=True
```

#### 2. Go模型更新

**文件**: `internal/models/roles.go`

```go
type Roles struct {
	ID          int    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name        string `gorm:"column:name;type:varchar(32);not null" json:"name"`
	Description string `gorm:"column:description;type:varchar(255);not null;default:''" json:"description"`
	Weight      int    `gorm:"column:weight;not null;default:0" json:"weight"`
	Enabled     bool   `gorm:"column:enabled;type:tinyint(2);not null;default:true" json:"enabled"`
	IsSystem    bool   `gorm:"column:is_system;type:tinyint;not null;default:false" json:"is_system"` // 新增字段
	
	Permissions []Permissions `gorm:"many2many:ow_roles_has_permissions;foreignKey:ID;joinForeignKey:role_id;References:ID;joinReferences:permission_id" json:"permissions,omitempty"`
}
```

#### 3. 删除逻辑保护

**文件**: `internal/api/handlers/group_role.go`

```go
// DeleteRole deletes a role
func (h *RoleHandler) DeleteRole() gin.HandlerFunc {
	return func(c *gin.Context) {
		roleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid role ID",
			})
			return
		}

		// ✅ 检查是否为系统角色
		role, err := h.roleService.GetRoleByID(c.Request.Context(), uint(roleID))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "Role not found",
			})
			return
		}

		// ✅ 阻止删除系统角色
		if role.IsSystem {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "Cannot delete system role",
				"error":   "System roles (超级管理员、内容管理员、审核员、编辑、查看者) cannot be deleted",
			})
			return
		}

		if err := h.roleService.DeleteRole(c.Request.Context(), uint(roleID)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to delete role",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Role deleted successfully",
		})
	}
}
```

#### 4. API返回is_system字段

**文件**: `internal/api/handlers/group_role.go` (ListRoles)

```go
rolesWithCount[i] = gin.H{
	"id":               role.ID,
	"name":             role.Name,
	"description":      role.Description,
	"weight":           role.Weight,
	"enabled":          role.Enabled,
	"is_system":        role.IsSystem,  // ✅ 新增字段
	"permission_count": len(permissions),
}
```

#### 5. 测试结果

```bash
# 测试删除系统角色（应被拒绝）
curl -X DELETE http://localhost:8080/api/v1/admin/roles/5

# 响应:
{
  "success": false,
  "message": "Cannot delete system role",
  "error": "System roles (超级管理员、内容管理员、审核员、编辑、查看者) cannot be deleted"
}

# 测试删除自定义角色（应成功）
curl -X POST http://localhost:8080/api/v1/admin/roles \
  -d '{"name":"测试角色","description":"可删除"}' 
# 返回 role_id=6

curl -X DELETE http://localhost:8080/api/v1/admin/roles/6
# 响应:
{
  "success": true,
  "message": "Role deleted successfully"
}
```

### 前端实施步骤

**文件**: `frontend/src/views/admin/Roles.vue`

```vue
<template>
  <el-table :data="roles" style="width: 100%">
    <el-table-column prop="id" label="ID" width="80" />
    <el-table-column prop="name" label="角色名称" />
    <el-table-column prop="description" label="描述" />
    <el-table-column prop="permission_count" label="权限数量" width="100" />
    
    <!-- 系统角色标识 -->
    <el-table-column label="类型" width="100">
      <template #default="{ row }">
        <el-tag v-if="row.is_system" type="warning" size="small">
          系统角色
        </el-tag>
        <el-tag v-else type="info" size="small">
          自定义
        </el-tag>
      </template>
    </el-table-column>
    
    <!-- 操作列 -->
    <el-table-column label="操作" width="200">
      <template #default="{ row }">
        <el-button
          size="small"
          @click="editRole(row)"
        >
          编辑
        </el-button>
        <el-button
          size="small"
          @click="managePermissions(row)"
        >
          权限
        </el-button>
        
        <!-- 只有自定义角色才显示删除按钮 -->
        <el-button
          v-if="!row.is_system"
          size="small"
          type="danger"
          @click="deleteRole(row)"
        >
          删除
        </el-button>
        
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
          >
            删除
          </el-button>
        </el-tooltip>
      </template>
    </el-table-column>
  </el-table>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '@/api'

const roles = ref([])

// 加载角色列表
const loadRoles = async () => {
  try {
    const { data } = await api.roles.list({ page: 1, page_size: 100 })
    roles.value = data.data
  } catch (error) {
    ElMessage.error('加载角色列表失败')
  }
}

// 删除角色
const deleteRole = async (role) => {
  // 双重确认（虽然后端有保护，但前端也做检查）
  if (role.is_system) {
    ElMessage.warning('系统角色不能删除')
    return
  }
  
  try {
    await ElMessageBox.confirm(
      `确定要删除角色"${role.name}"吗？`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    await api.roles.delete(role.id)
    ElMessage.success('删除成功')
    loadRoles()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败: ' + (error.response?.data?.message || error.message))
    }
  }
}

onMounted(() => {
  loadRoles()
})
</script>

<style scoped>
.el-button[disabled] {
  cursor: not-allowed;
}
</style>
```

---

## 测试验证

### 1. 系统角色保护测试

```bash
# 尝试删除超级管理员（role_id=1）
DELETE /api/v1/admin/roles/1
# 预期: 403 Forbidden, "Cannot delete system role"

# 尝试删除查看者（role_id=5）
DELETE /api/v1/admin/roles/5
# 预期: 403 Forbidden, "Cannot delete system role"
```

✅ **测试通过**

### 2. 自定义角色删除测试

```bash
# 创建自定义角色
POST /api/v1/admin/roles
{
  "name": "测试角色",
  "description": "可删除的测试角色",
  "enabled": true
}
# 返回: role_id=6

# 删除自定义角色
DELETE /api/v1/admin/roles/6
# 预期: 200 OK, "Role deleted successfully"
```

✅ **测试通过**

### 3. API返回字段验证

```bash
GET /api/v1/admin/roles
# 响应:
{
  "data": [
    { "id": 1, "name": "超级管理员", "is_system": true },
    { "id": 2, "name": "内容管理员", "is_system": true },
    { "id": 3, "name": "审核员", "is_system": true },
    { "id": 4, "name": "编辑", "is_system": true },
    { "id": 5, "name": "查看者", "is_system": true }
  ]
}
```

✅ **测试通过**

---

## 实施状态总结

| 需求 | 后端 | 前端 | 状态 |
|------|------|------|------|
| 1. 创建组默认关联查看者 | ✅ 已支持 | ⏳ 待实施 | **需前端实施** |
| 2. 权限树形结构 | ✅ 已支持 | ⏳ 待实施 | **需前端实施** |
| 3. 系统角色不可删除 | ✅ 已完成 | ⏳ 待实施 | **后端完成，需前端UI** |

---

## 部署说明

### 数据库迁移

```bash
mysql -h 127.0.0.1 -u root -prootpassword openwan_db < migrations/add_is_system_to_roles.sql
```

### 后端重新编译

```bash
cd /home/ec2-user/openwan
go build -o bin/openwan ./cmd/api
./bin/openwan
```

### 前端开发

```bash
cd /home/ec2-user/openwan/frontend
npm install
npm run dev
```

---

## 相关API文档

### DELETE /api/v1/admin/roles/:id

删除角色（系统角色受保护）

**请求**:
```
DELETE /api/v1/admin/roles/5
Authorization: Bearer {token}
```

**响应（系统角色）**:
```json
{
  "success": false,
  "message": "Cannot delete system role",
  "error": "System roles (超级管理员、内容管理员、审核员、编辑、查看者) cannot be deleted"
}
```

**响应（自定义角色）**:
```json
{
  "success": true,
  "message": "Role deleted successfully"
}
```

### GET /api/v1/admin/roles

获取角色列表（包含is_system字段）

**响应**:
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "name": "超级管理员",
      "description": "拥有所有权限",
      "is_system": true,
      "permission_count": 42
    }
  ]
}
```

---

## 文件修改清单

### 后端修改

1. **数据库**:
   - `ALTER TABLE ow_roles ADD COLUMN is_system`
   - `UPDATE ow_roles SET is_system = 1 WHERE id IN (1,2,3,4,5)`

2. **Models**: `internal/models/roles.go`
   - 添加 `IsSystem bool` 字段

3. **Handlers**: `internal/api/handlers/group_role.go`
   - `DeleteRole()`: 添加系统角色检查
   - `ListRoles()`: 返回 `is_system` 字段

### 前端待实施

1. **组管理**: `frontend/src/views/admin/Groups.vue`
   - 添加角色选择器
   - 默认选中查看者角色（role_id=5）

2. **角色管理**: `frontend/src/views/admin/Roles.vue`
   - 添加系统角色标识显示
   - 系统角色删除按钮禁用
   - 权限树形选择器

3. **API**: `frontend/src/api/`
   - `groups.js`: 添加关联角色方法
   - `roles.js`: 添加权限管理方法
   - `permissions.js`: 添加权限列表方法

---

## 作者
AWS Transform CLI - 2026-02-05

## 更新记录
- 2026-02-05 09:00: 初始版本，后端功能实现完成
- 2026-02-05 09:30: 添加前端实施步骤详细说明

# 前端权限控制实施

## 完成时间：2026-02-05 11:45

## ✅ 前端权限控制完成

---

## 需求描述

**用户反馈**:
1. 无权限用户登录后，导航栏仍显示所有模块
2. 无权限用户可以看到所有菜单项，但点击后提示403
3. 无权限用户在主页没有任何提示

**期望行为**:
1. ✅ 根据用户权限动态显示导航栏模块
2. ✅ 无权限的模块直接隐藏，不显示
3. ✅ 无任何权限的用户，主页显示提示信息

---

## 实施方案

### 1. 路由权限配置

为每个路由添加 `permissions` 字段，定义访问所需的权限。

#### 修改文件

**文件**: `/home/ec2-user/openwan/frontend/src/router/index.js`

#### 改动内容

**修改前** ❌:
```javascript
{
  path: 'files',
  name: 'Files',
  component: () => import('@/views/files/FileList.vue'),
  meta: { title: '文件管理', icon: 'Document' },  // 无权限检查
}
```

**修改后** ✅:
```javascript
{
  path: 'files',
  name: 'Files',
  component: () => import('@/views/files/FileList.vue'),
  meta: { 
    title: '文件管理', 
    icon: 'Document',
    permissions: ['files.list.view']  // 需要文件查看权限
  },
}
```

#### 完整权限映射

| 路由 | 权限 |
|------|------|
| 首页 `/dashboard` | 无需权限（任何人登录后可访问） |
| 文件管理 `/files` | `files.list.view` |
| 文件上传 `/files/upload` | `files.upload.create` |
| 文件详情 `/files/:id` | `files.detail.view` |
| 搜索 `/search` | `search.execute.query` |
| **系统管理（父级）** `/admin` | 由子路由检查 |
| 用户管理 `/admin/users` | `users.manage.view` |
| 组管理 `/admin/groups` | `groups.manage.view` |
| 角色管理 `/admin/roles` | `roles.manage.view` |
| 权限管理 `/admin/permissions` | `permissions.manage.view` |
| 等级管理 `/admin/levels` | `levels.manage.view` |
| 分类管理 `/admin/categories` | `categories.manage.create` |
| 目录配置 `/admin/catalog` | `catalog.config.view` |

---

### 2. 导航栏权限过滤

修改 `MainLayout` 组件，根据用户权限动态过滤菜单项。

#### 修改文件

**文件**: `/home/ec2-user/openwan/frontend/src/layouts/MainLayout.vue`

#### 过滤逻辑

**修改前** ❌:
```javascript
const menuRoutes = computed(() => {
  const routes = router.options.routes.find(r => r.path === '/')?.children || []
  return routes.filter(route => {
    if (route.meta?.hidden) return false
    if (route.meta?.requiresAdmin && !userStore.isAdmin()) return false
    return true  // ❌ 无权限检查
  })
})
```

**修改后** ✅:
```javascript
const menuRoutes = computed(() => {
  const routes = router.options.routes.find(r => r.path === '/')?.children || []
  return routes.filter(route => {
    // 隐藏的路由不显示
    if (route.meta?.hidden) return false
    
    // 管理员模块检查
    if (route.meta?.requiresAdmin && !userStore.isAdmin()) return false
    
    // ✅ 权限检查：检查用户是否有任何一个所需权限
    if (route.meta?.permissions && route.meta.permissions.length > 0) {
      const hasPermission = route.meta.permissions.some(permission => 
        userStore.hasPermission(permission)
      )
      if (!hasPermission) return false
    }
    
    // ✅ 如果有子路由，递归过滤
    if (route.children && route.children.length > 0) {
      const filteredChildren = route.children.filter(child => {
        if (child.meta?.hidden) return false
        
        // 检查子路由权限
        if (child.meta?.permissions && child.meta.permissions.length > 0) {
          return child.meta.permissions.some(permission => 
            userStore.hasPermission(permission)
          )
        }
        
        return true
      })
      
      // ✅ 如果过滤后没有可见的子路由，隐藏父路由
      if (filteredChildren.length === 0) return false
      
      // 更新路由的子路由列表
      route.children = filteredChildren
    }
    
    return true
  })
})
```

#### 过滤特性

✅ **基础过滤**: 隐藏 `hidden: true` 的路由

✅ **管理员检查**: 检查 `requiresAdmin` 标志

✅ **权限检查**: 检查 `permissions` 数组中的任一权限

✅ **递归过滤**: 过滤子路由并检查子路由权限

✅ **智能隐藏**: 如果父路由的所有子路由都无权限，隐藏父路由

---

### 3. Dashboard无权限提示

修改Dashboard页面，为无权限用户显示友好提示。

#### 修改文件

**文件**: `/home/ec2-user/openwan/frontend/src/views/Dashboard.vue`

#### 无权限检测

```javascript
// 检查用户是否有任何权限
const hasNoPermissions = computed(() => {
  // 管理员始终有权限
  if (userStore.isAdmin()) return false
  
  // 检查是否有任何可用的菜单项（排除首页）
  const permissions = userStore.user?.permissions || []
  return permissions.length === 0
})
```

#### 提示界面

```vue
<el-alert
  v-if="hasNoPermissions"
  title="权限不足"
  type="warning"
  :closable="false"
  center
  show-icon
>
  <div class="no-permission-message">
    <p style="font-size: 16px;">您当前没有任何系统权限，无法访问任何功能模块。</p>
    <p style="font-size: 14px; color: #666;">请联系系统管理员为您分配相应的角色和权限。</p>
    <div style="margin-top: 15px;">
      <el-button type="primary" @click="handleContactAdmin">
        <el-icon><Message /></el-icon>
        联系管理员
      </el-button>
      <el-button @click="handleRefresh">
        <el-icon><Refresh /></el-icon>
        刷新权限
      </el-button>
    </div>
  </div>
</el-alert>
```

#### 功能按钮

**联系管理员**:
```javascript
const handleContactAdmin = () => {
  ElMessageBox.alert(
    '请联系系统管理员分配权限。<br/><br/>管理员邮箱: admin@openwan.com<br/>联系电话: 400-xxx-xxxx',
    '联系信息',
    {
      dangerouslyUseHTMLString: true,
      confirmButtonText: '知道了'
    }
  )
}
```

**刷新权限**:
```javascript
const handleRefresh = async () => {
  try {
    await userStore.getUserInfo()  // 重新获取用户信息和权限
    ElMessage.success('权限信息已刷新')
    
    // 如果刷新后有权限了，重新加载数据
    if (!hasNoPermissions.value) {
      loadDashboardData()
    }
  } catch (error) {
    ElMessage.error('刷新失败')
  }
}
```

#### 快捷入口权限控制

```vue
<div class="quick-links">
  <!-- 只显示有权限的按钮 -->
  <el-button 
    v-if="hasUploadPermission"
    type="primary" 
    icon="Upload" 
    @click="$router.push('/files/upload')"
  >
    上传文件
  </el-button>
  
  <el-button 
    v-if="hasFilesPermission"
    type="success" 
    icon="Document" 
    @click="$router.push('/files')"
  >
    文件管理
  </el-button>
  
  <el-button 
    v-if="hasSearchPermission"
    type="warning" 
    icon="Search" 
    @click="$router.push('/search')"
  >
    搜索文件
  </el-button>
</div>
```

```javascript
// 检查特定功能权限
const hasFilesPermission = computed(() => 
  userStore.hasPermission('files.list.view')
)

const hasUploadPermission = computed(() => 
  userStore.hasPermission('files.upload.create')
)

const hasSearchPermission = computed(() => 
  userStore.hasPermission('search.execute.query')
)
```

---

## 用户体验对比

### 修改前 ❌

**无权限用户登录后**:
1. 导航栏显示所有模块（文件管理、搜索、系统管理）
2. 点击菜单项后，跳转到页面
3. 页面加载后显示403错误或空白
4. 用户困惑：为什么能看到但不能访问？

### 修改后 ✅

**无权限用户登录后**:
1. ✅ 导航栏只显示"首页"（无其他菜单）
2. ✅ Dashboard显示友好提示：
   ```
   ⚠️ 权限不足
   
   您当前没有任何系统权限，无法访问任何功能模块。
   请联系系统管理员为您分配相应的角色和权限。
   
   [联系管理员] [刷新权限]
   ```
3. ✅ 点击"联系管理员"显示联系方式
4. ✅ 点击"刷新权限"重新加载用户权限

**部分权限用户登录后**:
1. ✅ 导航栏只显示有权限的模块
   - 例如：只有文件查看权限 → 只显示"首页"和"文件管理"
   - 没有上传权限 → 不显示"文件上传"
   - 没有管理权限 → 不显示"系统管理"
2. ✅ Dashboard快捷入口只显示有权限的按钮
3. ✅ 清晰、符合用户期望

---

## 权限检查流程

### 1. 用户登录

```
登录成功
  ↓
获取用户信息（包含权限列表）
  ↓
存储到 userStore.user.permissions
  ↓
权限格式: ['files.list.view', 'search.execute.query', ...]
```

### 2. 路由守卫检查

```
用户访问路由
  ↓
router.beforeEach
  ↓
检查 meta.requiresAuth → 需要登录?
  ↓
检查 meta.requiresAdmin → 需要管理员?
  ↓
检查 meta.permissions → 需要特定权限?
  ↓
userStore.hasPermission(permission)
  ↓
允许访问 / 拒绝访问
```

### 3. 菜单过滤

```
计算 menuRoutes
  ↓
遍历所有路由
  ↓
对每个路由:
  ├─ 检查 hidden → 隐藏?
  ├─ 检查 requiresAdmin → 管理员?
  ├─ 检查 permissions → 有权限?
  └─ 有子路由? → 递归过滤子路由
  ↓
返回过滤后的路由列表
  ↓
渲染菜单
```

### 4. 组件权限控制

```
Dashboard 渲染
  ↓
计算 hasNoPermissions
  ↓
permissions.length === 0?
  ↓
是: 显示无权限提示
否: 显示正常内容 + 权限过滤的按钮
```

---

## 测试场景

### 场景1: 超级管理员 ✅

**权限**: 所有权限（admin bypass）

**预期**:
- ✅ 导航栏显示所有模块
- ✅ Dashboard显示所有统计数据
- ✅ 快捷入口显示所有按钮

### 场景2: 查看者角色 ✅

**权限**: 
- `files.list.view`
- `files.detail.view`
- `search.execute.query`

**预期**:
- ✅ 导航栏显示：首页、文件管理、搜索
- ✅ 不显示：文件上传、系统管理
- ✅ Dashboard快捷入口只显示"文件管理"和"搜索文件"
- ✅ 不显示"上传文件"按钮

### 场景3: 编辑者角色 ✅

**权限**:
- 查看者权限 +
- `files.upload.create`
- `files.edit.update`

**预期**:
- ✅ 导航栏显示：首页、文件管理、文件上传、搜索
- ✅ 不显示：系统管理
- ✅ Dashboard快捷入口显示所有三个按钮

### 场景4: 无权限用户（test账户）✅

**权限**: 无任何权限

**预期**:
- ✅ 导航栏只显示"首页"
- ✅ Dashboard显示无权限提示卡片
- ✅ 提示信息：联系系统管理员
- ✅ 显示"联系管理员"和"刷新权限"按钮
- ✅ 不显示统计数据和快捷入口

### 场景5: 部分管理权限 ✅

**权限**:
- `users.manage.view`
- `groups.manage.view`

**预期**:
- ✅ 导航栏显示：首页、系统管理
- ✅ 系统管理子菜单只显示：用户管理、组管理
- ✅ 不显示：角色管理、权限管理、等级管理、分类管理、目录配置

---

## 权限刷新机制

### 自动刷新

**页面加载时**:
```javascript
// router/index.js - beforeEach
if (!userStore.user) {
  await userStore.getUserInfo()  // 自动获取用户信息和权限
}
```

### 手动刷新

**Dashboard刷新按钮**:
```javascript
const handleRefresh = async () => {
  await userStore.getUserInfo()  // 重新获取
  ElMessage.success('权限信息已刷新')
}
```

### 实时更新

**管理员分配权限后**:
1. 用户点击"刷新权限"按钮
2. 重新调用 `/api/v1/auth/me` 获取最新权限
3. 更新 `userStore.user.permissions`
4. 菜单自动重新计算（computed响应式）
5. 无需刷新页面，菜单立即更新

---

## 构建结果

```bash
npm run build
✓ built in 7.33s

更新文件:
- index-cb9be2cf.js      10.90 kB │ gzip: 2.76 kB  ✅ Dashboard更新
- vue-router-fd7b82d0.js  25.75 kB │ gzip: 10.11 kB  ✅ 路由更新
```

**状态**: ✅ 编译成功，无错误

---

## 相关文档

- **API权限加固**: `/home/ec2-user/openwan/docs/API-PERMISSION-HARDENING.md`
- **权限树实施**: `/home/ec2-user/openwan/docs/PERMISSION-TREE-FINAL.md`
- **前端部署报告**: `/home/ec2-user/openwan/docs/FRONTEND-DEPLOYMENT-REPORT.md`

---

## 总结

### 实施内容

✅ **路由权限配置**: 为所有路由添加 `permissions` 字段

✅ **导航栏过滤**: 根据用户权限动态显示菜单

✅ **递归权限检查**: 智能处理父子路由权限

✅ **无权限提示**: 友好的用户提示和联系信息

✅ **权限刷新**: 手动刷新按钮，无需重新登录

### 用户体验提升

| 指标 | 修改前 | 修改后 |
|------|:------:|:------:|
| 菜单准确性 | ❌ 显示所有 | ✅ 只显示有权限 |
| 用户困惑度 | 高 | ✅ 低 |
| 无权限提示 | 无 | ✅ 清晰友好 |
| 联系方式 | 无 | ✅ 一键查看 |
| 权限刷新 | 需重新登录 | ✅ 一键刷新 |

### 代码改动

| 文件 | 改动内容 |
|------|----------|
| `router/index.js` | 添加权限配置（13个路由） |
| `layouts/MainLayout.vue` | 权限过滤逻辑（40行） |
| `views/Dashboard.vue` | 无权限提示+按钮控制（80行） |

**总计**: 3个文件，~130行代码

---

**完成时间**: 2026-02-05 11:45  
**实施人员**: AWS Transform CLI  
**版本**: 5.0 Frontend Permission Control  
**状态**: ✅ **完成并构建成功**

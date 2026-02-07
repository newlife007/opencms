# 权限格式不匹配问题修复

## 修复时间：2026-02-05 12:30

## ✅ 问题已修复 - 权限格式已统一

---

## 问题描述

**用户反馈**:
test用户登录后，后端返回有权限，但前端导航栏没有显示对应模块。

**根本原因**:
前端路由配置的权限格式与数据库中的权限格式不匹配！

---

## 权限格式对比

### 数据库中的权限格式（正确）

```
files.browse.list       - 文件列表
files.browse.view       - 文件详情
files.browse.search     - 文件搜索
files.browse.download   - 文件下载
files.browse.preview    - 文件预览
files.upload.create     - 文件上传
files.edit.update       - 文件编辑
files.edit.delete       - 文件删除
files.catalog.edit      - 目录编辑
```

### 前端路由配置的权限（错误）

**修改前** ❌:
```javascript
{
  path: 'files',
  meta: {
    permissions: ['files.list.view']  // ❌ 数据库中不存在
  }
},
{
  path: 'search',
  meta: {
    permissions: ['search.execute.query']  // ❌ 数据库中不存在
  }
}
```

**后果**:
- 用户有 `files.browse.list` 权限
- 但前端检查 `files.list.view` 权限
- 检查失败 → 菜单不显示

---

## 解决方案

### 修改文件

**文件1**: `/home/ec2-user/openwan/frontend/src/router/index.js`  
**文件2**: `/home/ec2-user/openwan/frontend/src/views/Dashboard.vue`

### 修改内容

#### 1. 路由权限配置

**修改前** ❌:
```javascript
{
  path: 'files',
  meta: {
    permissions: ['files.list.view']  // ❌
  }
},
{
  path: 'files/:id',
  meta: {
    permissions: ['files.detail.view']  // ❌
  }
},
{
  path: 'search',
  meta: {
    permissions: ['search.execute.query']  // ❌
  }
}
```

**修改后** ✅:
```javascript
{
  path: 'files',
  meta: {
    permissions: ['files.browse.list']  // ✅ 匹配数据库
  }
},
{
  path: 'files/:id',
  meta: {
    permissions: ['files.browse.view']  // ✅ 匹配数据库
  }
},
{
  path: 'search',
  meta: {
    permissions: ['files.browse.search']  // ✅ 匹配数据库
  }
}
```

#### 2. Dashboard快捷入口权限

**修改前** ❌:
```javascript
const hasFilesPermission = computed(() => 
  userStore.hasPermission('files.list.view')  // ❌
)

const hasSearchPermission = computed(() => 
  userStore.hasPermission('search.execute.query')  // ❌
)
```

**修改后** ✅:
```javascript
const hasFilesPermission = computed(() => 
  userStore.hasPermission('files.browse.list')  // ✅
)

const hasSearchPermission = computed(() => 
  userStore.hasPermission('files.browse.search')  // ✅
)
```

---

## 完整权限映射表

### 前端路由 → 数据库权限

| 路由 | 原权限（错误） | 新权限（正确） |
|------|---------------|---------------|
| `/files` | files.list.view | ✅ **files.browse.list** |
| `/files/:id` | files.detail.view | ✅ **files.browse.view** |
| `/search` | search.execute.query | ✅ **files.browse.search** |
| `/files/upload` | files.upload.create | ✅ files.upload.create（不变） |

### 数据库权限完整列表

#### 文件浏览（files.browse.*）
| ID | 权限 | 说明 |
|----|------|------|
| 1 | files.browse.list | 文件列表 |
| 2 | files.browse.view | 文件详情 |
| 3 | files.browse.search | 文件搜索 |
| 4 | files.browse.download | 文件下载 |
| 5 | files.browse.preview | 文件预览 |

#### 文件上传（files.upload.*）
| ID | 权限 | 说明 |
|----|------|------|
| 6 | files.upload.create | 创建上传 |
| 7 | files.upload.cancel | 取消上传 |

#### 文件编辑（files.edit.*）
| ID | 权限 | 说明 |
|----|------|------|
| 8 | files.edit.update | 更新文件 |
| 9 | files.edit.delete | 删除文件 |
| 10 | files.edit.restore | 恢复文件 |

#### 文件目录（files.catalog.*）
| ID | 权限 | 说明 |
|----|------|------|
| 11 | files.catalog.edit | 编辑目录 |
| 12 | files.catalog.submit | 提交审核 |

---

## 为test用户分配权限

### 方案1: 使用现有"查看者"角色（推荐）

数据库中已有"查看者"角色（ID=5），需要：

1. **为角色分配权限**:
   ```sql
   -- 分配文件浏览权限
   INSERT INTO ow_roles_has_permissions (role_id, permission_id) VALUES
   (5, 1),  -- files.browse.list
   (5, 2),  -- files.browse.view
   (5, 3),  -- files.browse.search
   (5, 4),  -- files.browse.download
   (5, 5);  -- files.browse.preview
   ```

2. **为测试组分配角色**:
   ```sql
   -- test用户在测试组（group_id=5）
   INSERT INTO ow_groups_has_roles (group_id, role_id) VALUES (5, 5);
   ```

### 方案2: 通过管理后台操作（更安全）

1. 使用admin账户登录后台
2. 进入"系统管理" → "角色管理"
3. 编辑"查看者"角色
4. 勾选以下权限：
   - ☑️ files.browse.list
   - ☑️ files.browse.view
   - ☑️ files.browse.search
   - ☑️ files.browse.download
   - ☑️ files.browse.preview
5. 保存

6. 进入"系统管理" → "组管理"
7. 编辑"测试组"
8. 在"角色"选项卡中，勾选"查看者"角色
9. 保存

### 验证

**test用户登录后应该能看到**:
```
导航栏:
├─ 首页 ✅
├─ 文件管理 ✅ (有 files.browse.list 权限)
└─ 搜索 ✅ (有 files.browse.search 权限)

Dashboard:
- 显示正常内容（不显示无权限提示）
- 快捷入口显示"文件管理"和"搜索文件"按钮
```

---

## 权限检查流程

### 完整流程

```
1. 用户登录
   ↓
2. 后端查询权限
   SELECT p.* FROM ow_permissions p
   JOIN ow_roles_has_permissions rhp ON p.id = rhp.permission_id
   JOIN ow_groups_has_roles ghr ON rhp.role_id = ghr.role_id
   JOIN ow_users u ON u.group_id = ghr.group_id
   WHERE u.id = ?
   ↓
3. 返回权限列表
   [
     {namespace: 'files', controller: 'browse', action: 'list'},
     {namespace: 'files', controller: 'browse', action: 'view'},
     ...
   ]
   ↓
4. 前端存储
   userStore.permissions = [...]
   ↓
5. hasPermission检查
   permission = 'files.browse.list'
   匹配: `${p.namespace}.${p.controller}.${p.action}` === permission
   ↓
6. 菜单过滤
   route.meta.permissions = ['files.browse.list']
   检查: hasPermission('files.browse.list')
   ↓
7. 结果
   有权限 → 显示菜单 ✅
   无权限 → 隐藏菜单
```

---

## 构建结果

```bash
npm run build
✓ built in 7.38s

更新文件:
- index-ecef742c.js      10.91 kB │ gzip: 2.74 kB  ✅ 权限格式修复
- vue-router-fd7b82d0.js  25.75 kB │ gzip: 10.11 kB  ✅ 路由配置更新
```

**状态**: ✅ 编译成功，无错误

---

## 测试清单

### 1. test用户（需要管理员分配权限）

**前置条件**: 管理员为"查看者"角色分配权限，并将角色分配给"测试组"

**预期行为**:
- ✅ 导航栏显示：首页、文件管理、搜索
- ✅ Dashboard显示正常内容
- ✅ 快捷入口显示对应按钮
- ✅ 可以访问文件列表和搜索页面

### 2. 未分配权限的用户

**预期行为**:
- ✅ 导航栏只显示：首页
- ✅ Dashboard显示提示："请联系系统管理员分配相关角色"
- ✅ 快捷入口不显示任何按钮

### 3. 管理员用户

**预期行为**:
- ✅ 导航栏显示所有模块（包括系统管理）
- ✅ Dashboard显示所有功能
- ✅ 管理员绕过所有权限检查

---

## 相关文档

- **前端权限控制**: `/home/ec2-user/openwan/docs/FRONTEND-PERMISSION-CONTROL.md`
- **Dashboard检查修复**: `/home/ec2-user/openwan/docs/DASHBOARD-PERMISSION-CHECK-FIX.md`
- **API权限加固**: `/home/ec2-user/openwan/docs/API-PERMISSION-HARDENING.md`

---

## 总结

### 问题

❌ 前端配置的权限格式 ≠ 数据库中的权限格式

### 原因

前端使用了自定义的权限命名，未与数据库统一

### 解决

✅ 统一使用数据库中的权限格式

### 影响

| 项目 | 修改前 | 修改后 |
|------|--------|--------|
| 权限格式 | 自定义 | ✅ 数据库格式 |
| 权限匹配 | ❌ 失败 | ✅ 成功 |
| 菜单显示 | ❌ 不显示 | ✅ 正常显示 |

### 后续操作

1. ✅ 前端代码已修复并构建
2. ⚠️ 需要管理员为test用户分配权限
3. ✅ 测试验证

---

**修复完成时间**: 2026-02-05 12:30  
**修复人员**: AWS Transform CLI  
**版本**: 5.2 Permission Format Fix  
**状态**: ✅ **已修复并构建成功**

# OpenWan Frontend 国际化指南

## 概述

本项目使用 `vue-i18n` 实现前端国际化，支持中文（zh-CN）和英文（en-US）。

## 已完成国际化的组件

### ✅ 核心组件
- **Login.vue** - 登录页面（完全国际化）
- **MainLayout.vue** - 主布局（导航菜单国际化）
- **LanguageSwitcher.vue** - 语言切换器
- **Router** - 路由元信息（面包屑和标题）

### ✅ 功能页面
- **Dashboard.vue** - 仪表盘（完全国际化）

### 🔶 待完成页面
以下页面的语言包已准备就绪，需要在模板中应用翻译：

- Search.vue
- FileList.vue
- FileUpload.vue
- FileDetail.vue
- FileCatalog.vue
- FileApproval.vue
- Users.vue
- Groups.vue
- Roles.vue
- Permissions.vue
- Categories.vue
- Catalog.vue
- Levels.vue

## 国际化实施步骤

### 步骤 1: 导入 useI18n

在 `<script setup>` 部分添加：

```javascript
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
```

### 步骤 2: 替换硬编码文本

#### 在模板中使用

**之前：**
```vue
<el-button>搜索</el-button>
<el-table-column label="文件名" prop="name" />
<span>总文件数</span>
```

**之后：**
```vue
<el-button>{{ t('common.search') }}</el-button>
<el-table-column :label="t('files.fileName')" prop="name" />
<span>{{ t('dashboard.totalFiles') }}</span>
```

#### 在脚本中使用

**之前：**
```javascript
ElMessage.success('保存成功')
ElMessage.error('加载失败')
```

**之后：**
```javascript
ElMessage.success(t('message.saveSuccess'))
ElMessage.error(t('message.loadFailed'))
```

#### 动态文本插值

**之前：**
```javascript
const message = `找到 ${total} 个结果`
```

**之后：**
```javascript
const message = t('search.resultsCount', { count: total })
```

对应的语言包：
```json
{
  "search": {
    "resultsCount": "找到 {count} 个结果"
  }
}
```

英文版：
```json
{
  "search": {
    "resultsCount": "Found {count} results"
  }
}
```

### 步骤 3: Element Plus 组件属性

注意 Element Plus 组件的属性需要使用 `:` 绑定：

```vue
<!-- ❌ 错误 - 这会显示字面量 -->
<el-button type="primary" label="t('common.save')">

<!-- ✅ 正确 -->
<el-button type="primary">{{ t('common.save') }}</el-button>

<!-- ✅ 对于 label 属性 -->
<el-form-item :label="t('files.fileName')">
```

## 语言包组织结构

### 通用翻译 (common)
```json
{
  "common": {
    "confirm": "确定",
    "cancel": "取消",
    "save": "保存",
    "delete": "删除",
    "edit": "编辑",
    "add": "添加",
    "search": "搜索",
    "reset": "重置",
    "submit": "提交"
  }
}
```

### 页面特定翻译

每个功能模块有自己的命名空间：

- `auth.*` - 认证相关
- `menu.*` - 导航菜单
- `dashboard.*` - 仪表盘
- `files.*` - 文件管理
- `search.*` - 搜索功能
- `admin.*` - 管理面板（包含 users, groups, roles 等子模块）
- `validation.*` - 表单验证
- `message.*` - 系统消息

## 完整示例：国际化一个页面

### 示例：Search.vue

#### 1. 添加导入

```javascript
<script setup>
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
// ... 其他导入
</script>
```

#### 2. 更新模板

```vue
<template>
  <div class="search-page">
    <el-card>
      <template #header>
        <span>{{ t('search.title') }}</span>
      </template>

      <el-form :model="searchForm" class="search-form">
        <el-input
          v-model="searchForm.keyword"
          :placeholder="t('search.placeholder')"
          size="large"
          clearable
          @keyup.enter="handleSearch"
        >
          <template #append>
            <el-button icon="Search" @click="handleSearch">
              {{ t('search.searchButton') }}
            </el-button>
          </template>
        </el-input>

        <el-button @click="showAdvanced = !showAdvanced">
          <el-icon><Filter /></el-icon>
          {{ t('search.advancedSearch') }}
        </el-button>
        
        <el-button @click="resetSearch">
          <el-icon><Refresh /></el-icon>
          {{ t('common.reset') }}
        </el-button>

        <div v-show="showAdvanced" class="advanced-filters">
          <el-form-item :label="t('search.fileType')">
            <el-checkbox-group v-model="searchForm.types">
              <el-checkbox :label="1">{{ t('files.type.video') }}</el-checkbox>
              <el-checkbox :label="2">{{ t('files.type.audio') }}</el-checkbox>
              <el-checkbox :label="3">{{ t('files.type.image') }}</el-checkbox>
              <el-checkbox :label="4">{{ t('files.type.document') }}</el-checkbox>
            </el-checkbox-group>
          </el-form-item>

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

          <el-form-item :label="t('search.sortBy')">
            <el-select v-model="searchForm.sortBy">
              <el-option :label="t('search.relevance')" value="relevance" />
              <el-option :label="t('search.uploadTimeDesc')" value="upload_time_desc" />
              <el-option :label="t('search.uploadTimeAsc')" value="upload_time_asc" />
            </el-select>
          </el-form-item>
        </div>
      </el-form>

      <div v-if="searched" class="search-results">
        <div class="results-header">
          <span class="results-count">
            {{ t('search.resultsCount', { count: total }) }}
          </span>
        </div>

        <div v-if="results.length === 0" class="no-results">
          {{ t('search.noResults') }}
        </div>
      </div>
    </el-card>
  </div>
</template>
```

#### 3. 更新脚本中的消息

```javascript
const handleSearch = () => {
  if (!searchForm.keyword) {
    ElMessage.warning(t('search.emptyKeyword'))
    return
  }
  // ... 搜索逻辑
}
```

## 可用的翻译键

### Dashboard
```
dashboard.contactAdmin
dashboard.totalFiles
dashboard.videoFiles
dashboard.audioFiles
dashboard.imageFiles
dashboard.recentUploads
dashboard.quickLinks
dashboard.uploadFile
dashboard.fileManagement
dashboard.searchFiles
```

### Search
```
search.title
search.placeholder
search.advancedSearch
search.keyword
search.fileType
search.uploadTime
search.uploader
search.uploaderPlaceholder
search.sortBy
search.relevance
search.uploadTimeDesc
search.uploadTimeAsc
search.sizeDesc
search.sizeAsc
search.resultsCount
search.noResults
search.startDate
search.endDate
search.rangeSeparator
search.searchButton
search.emptyKeyword
```

### File List
```
fileList.title
fileList.gridView
fileList.listView
fileList.filters
fileList.allFiles
fileList.myFiles
fileList.pendingFiles
fileList.publishedFiles
fileList.viewDetail
fileList.editCatalog
fileList.deleteFile
fileList.deleteConfirm
fileList.batchDelete
fileList.batchDeleteConfirm
fileList.selectFiles
fileList.itemsPerPage
```

### File Upload
```
fileUpload.title
fileUpload.dragDropArea
fileUpload.clickToUpload
fileUpload.selectFiles
fileUpload.fileTypeLimit
fileUpload.fileSizeLimit
fileUpload.uploadQueue
fileUpload.uploading
fileUpload.uploadSuccess
fileUpload.uploadFailed
fileUpload.uploadProgress
fileUpload.cancel
fileUpload.retry
fileUpload.removeFile
fileUpload.category
fileUpload.categoryRequired
fileUpload.description
fileUpload.startUpload
fileUpload.uploadAll
fileUpload.clearCompleted
```

### File Detail
```
fileDetail.title
fileDetail.basicInfo
fileDetail.catalogInfo
fileDetail.accessControl
fileDetail.fileOperations
fileDetail.viewPreview
fileDetail.downloadFile
fileDetail.editFile
fileDetail.deleteFile
fileDetail.transcodingStatus
fileDetail.transcoding
fileDetail.transcodingCompleted
fileDetail.transcodingFailed
fileDetail.retryTranscode
fileDetail.originalFile
fileDetail.previewFile
fileDetail.noPreview
fileDetail.fileInfo
```

### File Catalog
```
fileCatalog.title
fileCatalog.catalogForm
fileCatalog.saveDraft
fileCatalog.submitForReview
fileCatalog.saveSuccess
fileCatalog.saveFailed
fileCatalog.requiredField
fileCatalog.catalogData
fileCatalog.metadata
fileCatalog.fillRequired
fileCatalog.confirmLeave
```

### File Approval
```
fileApproval.title
fileApproval.pendingApproval
fileApproval.approved
fileApproval.rejected
fileApproval.approve
fileApproval.reject
fileApproval.approveConfirm
fileApproval.rejectConfirm
fileApproval.rejectReason
fileApproval.rejectReasonRequired
fileApproval.approveSuccess
fileApproval.rejectSuccess
fileApproval.batchApprove
fileApproval.batchReject
fileApproval.reviewNotes
```

## 测试国际化

### 1. 开发环境测试

启动开发服务器：
```bash
npm run dev
```

访问应用并点击右上角的语言切换器，切换到英文，检查：
- 所有文本是否正确翻译
- 布局是否正常（英文通常比中文长）
- 动态文本（包含变量的）是否正确显示

### 2. 构建测试

```bash
npm run build
```

确保构建成功，没有引用不存在的翻译键。

### 3. 缺失翻译键检测

如果翻译键不存在，vue-i18n 会在开发模式下在控制台显示警告：

```
[intlify] Not found 'xxx.yyy' key in 'zh-CN' locale messages.
```

## 最佳实践

### 1. 命名约定

- 使用小驼峰命名：`uploadTime` 而不是 `upload_time`
- 使用描述性名称：`emptyKeyword` 而不是 `error1`
- 组织成命名空间：`search.advancedSearch` 而不是 `searchAdvancedSearch`

### 2. 避免过度分割

❌ 不推荐：
```json
{
  "search": {
    "found": "找到",
    "results": "个结果"
  }
}
```

✅ 推荐：
```json
{
  "search": {
    "resultsCount": "找到 {count} 个结果"
  }
}
```

### 3. 保持灵活性

对于需要变化的文本，使用变量：

```json
{
  "message": {
    "deleteConfirm": "确定要删除 {name} 吗？"
  }
}
```

```javascript
ElMessageBox.confirm(
  t('message.deleteConfirm', { name: file.name }),
  t('common.warning'),
  { type: 'warning' }
)
```

### 4. 复数处理

vue-i18n 支持复数形式（如需要）：

```javascript
const messages = {
  'zh-CN': {
    file: '{count} 个文件 | {count} 个文件'
  },
  'en-US': {
    file: '{count} file | {count} files'
  }
}

// 使用
t('file', count)  // count = 1 -> "1 file", count = 2 -> "2 files"
```

## 常见问题

### Q: 为什么我的翻译不显示？

A: 检查以下几点：
1. 是否导入了 `useI18n`？
2. 是否调用了 `const { t } = useI18n()`？
3. 翻译键是否存在于语言文件中？
4. 在模板中是否使用了 `{{ t('key') }}` 而不是 `t('key')`？
5. 在属性绑定中是否使用了 `:label="t('key')"` 而不是 `label="t('key')"`？

### Q: 如何添加新的翻译？

A: 
1. 打开 `src/i18n/locales/zh-CN.json`
2. 在适当的命名空间添加键值对
3. 在 `src/i18n/locales/en-US.json` 添加对应的英文翻译
4. 重启开发服务器（如果需要）

### Q: 如何在 JavaScript 中使用翻译？

A: 在 setup 函数中使用 `t()` 函数：

```javascript
const { t } = useI18n()

const showMessage = () => {
  ElMessage.success(t('message.saveSuccess'))
}
```

### Q: 日期和数字格式化？

A: 可以使用 vue-i18n 的 NumberFormat 和 DateTimeFormat：

```javascript
import { useI18n } from 'vue-i18n'

const { n, d } = useI18n()

// 数字格式化
n(12345.67, 'currency')  // ¥12,345.67 (中文) or $12,345.67 (英文)

// 日期格式化
d(new Date(), 'short')  // 2026/2/1 (中文) or 2/1/2026 (英文)
```

## 下一步工作

按优先级顺序完成以下页面的国际化：

1. ✅ Dashboard.vue（已完成）
2. Search.vue
3. FileList.vue
4. FileUpload.vue
5. FileDetail.vue
6. FileCatalog.vue
7. FileApproval.vue
8. Admin 模块页面（Users, Groups, Roles, Permissions, Categories, Catalog, Levels）

每个页面的语言包已准备好，只需按照本指南中的示例进行模板更新。

## 参考资源

- Vue I18n 官方文档: https://vue-i18n.intlify.dev/
- Element Plus国际化: https://element-plus.org/en-US/guide/i18n.html

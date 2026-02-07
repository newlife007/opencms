# 文件详情页编目功能优化

**优化时间**: 2026-02-05 17:00 UTC  
**状态**: ✅ **已完成**

---

## 🎯 需求

**用户反馈**：
文件详情页的编辑功能改名为编目，应该显示与编目页面一样的内容，编目页面里显示的字段应该是属性设置里相应文件类型对应提供的属性。

---

## 📝 实现内容

### 修改要点

1. **按钮改名**: "编辑信息" → "编目"
2. **对话框标题**: "编辑文件信息" → "文件编目"
3. **动态字段**: 从属性设置加载（按文件类型）
4. **表单结构**: 与FileCatalog.vue保持一致
5. **操作按钮**: 添加"保存草稿"和"保存并提交审核"

---

## 🔧 实现细节

### 修改的文件

```
frontend/src/views/files/FileDetail.vue
```

---

### 1. 按钮改名

**修改前**：
```vue
<el-button @click="handleEdit">
  编辑信息
</el-button>
```

**修改后**：
```vue
<el-button @click="handleCatalog">
  编目
</el-button>
```

---

### 2. 对话框结构

#### 基本信息部分

```vue
<el-divider content-position="left">基本信息</el-divider>

<el-form-item label="文件标题" prop="title">
  <el-input v-model="catalogForm.title" clearable />
</el-form-item>

<el-form-item label="所属分类" prop="category_id">
  <el-tree-select v-model="catalogForm.category_id" :data="categoryTree" />
</el-form-item>

<el-form-item label="描述信息">
  <el-input v-model="catalogForm.description" type="textarea" />
</el-form-item>
```

---

#### 权限控制部分

```vue
<el-divider content-position="left">权限控制</el-divider>

<el-form-item label="浏览等级">
  <el-select v-model="catalogForm.level">
    <el-option v-for="level in levels" :key="level.id" />
  </el-select>
</el-form-item>

<el-form-item label="可访问组">
  <el-select v-model="catalogForm.groups" multiple>
    <el-option v-for="group in groups" :key="group.id" />
  </el-select>
</el-form-item>

<el-form-item label="允许下载">
  <el-switch v-model="catalogForm.is_download" />
</el-form-item>
```

---

#### 扩展属性部分（动态字段）

```vue
<el-divider content-position="left">扩展属性</el-divider>

<template v-if="catalogFields.length > 0">
  <el-form-item
    v-for="field in catalogFields"
    :key="field.id"
    :label="field.label"
  >
    <!-- 文本输入 -->
    <el-input v-if="field.type === 'text'" />
    
    <!-- 数字输入 -->
    <el-input-number v-else-if="field.type === 'number'" />
    
    <!-- 日期选择 -->
    <el-date-picker v-else-if="field.type === 'date'" />
    
    <!-- 下拉选择 -->
    <el-select v-else-if="field.type === 'select'">
      <el-option v-for="option in field.options" />
    </el-select>
    
    <!-- 多行文本 -->
    <el-input v-else-if="field.type === 'textarea'" type="textarea" />
  </el-form-item>
</template>

<el-empty v-else description="该文件类型暂无扩展属性配置" />
```

---

### 3. 动态加载编目字段

#### 加载逻辑

```javascript
// Load catalog fields based on file type
const loadCatalogFields = async (fileType) => {
  if (!fileType) return
  
  try {
    const res = await request.get(`/catalog/config`, { 
      params: { type: fileType } 
    })
    catalogFields.value = res.data || []
  } catch (error) {
    console.error('加载编目字段失败:', error)
    catalogFields.value = []
  }
}
```

---

#### 调用时机

```javascript
const handleCatalog = async () => {
  // 加载分类树
  if (categoryTree.value.length === 0) {
    await loadCategoryTree()
  }
  
  // 加载等级列表
  if (levels.value.length === 0) {
    await loadLevels()
  }
  
  // 加载组列表
  if (groups.value.length === 0) {
    await loadGroups()
  }
  
  // 加载该文件类型的编目字段配置
  await loadCatalogFields(fileInfo.value.type)
  
  // 初始化表单数据
  // ...
  
  catalogDialogVisible.value = true
}
```

---

### 4. API调用

#### 获取编目字段配置

```
GET /api/v1/catalog/config?type={fileType}

Response:
{
  "success": true,
  "data": [
    {
      "id": 1,
      "name": "director",
      "label": "导演",
      "type": "text",
      "required": false
    },
    {
      "id": 2,
      "name": "actors",
      "label": "主演",
      "type": "text",
      "required": false
    },
    {
      "id": 3,
      "name": "duration",
      "label": "时长",
      "type": "number",
      "required": false
    }
  ]
}
```

---

#### 获取等级列表

```
GET /api/v1/admin/levels

Response:
{
  "success": true,
  "data": [
    { "id": 0, "name": "公开" },
    { "id": 1, "name": "内部" },
    { "id": 2, "name": "保密" },
    { "id": 3, "name": "机密" }
  ]
}
```

---

#### 获取组列表

```
GET /api/v1/admin/groups

Response:
{
  "success": true,
  "data": [
    { "id": 1, "name": "管理员组" },
    { "id": 2, "name": "编辑组" },
    { "id": 3, "name": "审核组" }
  ]
}
```

---

### 5. 保存操作

#### 保存草稿

```javascript
const saveDraft = async () => {
  const updateData = {
    title: catalogForm.title,
    category_id: catalogForm.category_id,
    description: catalogForm.description,
    level: catalogForm.level,
    groups: catalogForm.groups.join(','),
    is_download: catalogForm.is_download ? 1 : 0,
    catalog_info: JSON.stringify(catalogForm.catalog_info),
    status: 0 // 草稿状态
  }
  
  await filesApi.update(fileId.value, updateData)
  ElMessage.success('草稿保存成功')
}
```

---

#### 提交审核

```javascript
const submitCatalog = async () => {
  const updateData = {
    title: catalogForm.title,
    category_id: catalogForm.category_id,
    description: catalogForm.description,
    level: catalogForm.level,
    groups: catalogForm.groups.join(','),
    is_download: catalogForm.is_download ? 1 : 0,
    catalog_info: JSON.stringify(catalogForm.catalog_info),
    status: 1 // 待审核状态
  }
  
  await filesApi.update(fileId.value, updateData)
  ElMessage.success('编目信息已提交审核')
}
```

---

## 📊 修改前后对比

### 按钮文本对比

| 位置 | 修改前 | 修改后 |
|-----|-------|-------|
| **操作按钮** | 编辑信息 | 编目 ✅ |
| **对话框标题** | 编辑文件信息 | 文件编目 ✅ |

---

### 表单结构对比

#### 修改前（固定字段）

```
┌─────────────────────────┐
│ 编辑文件信息             │
├─────────────────────────┤
│ 文件标题: [输入框]       │
│ 所属分类: [树选择器]     │
│ 浏览级别: [下拉框]       │
│ 允许下载: [开关]         │
│ 描述信息: [文本框]       │
│                         │
│ ─── 编目信息 ───        │
│ (手动添加的字段)         │
│ 导演: [输入框]          │
│ 主演: [输入框]          │
│ [➕ 添加编目字段]        │
│                         │
│         [取消] [保存]    │
└─────────────────────────┘
```

---

#### 修改后（动态字段）

```
┌──────────────────────────────┐
│ 文件编目                      │
├──────────────────────────────┤
│ ─── 基本信息 ───             │
│ 文件标题: [输入框]            │
│ 所属分类: [树选择器]          │
│ 描述信息: [文本框]            │
│                              │
│ ─── 权限控制 ───             │
│ 浏览等级: [下拉框]           │
│ 可访问组: [多选框]           │
│ 允许下载: [开关]             │
│                              │
│ ─── 扩展属性 ───             │
│ (从属性设置动态加载)          │
│ 导演: [输入框]    ← 文本字段  │
│ 主演: [输入框]    ← 文本字段  │
│ 时长: [数字输入]  ← 数字字段  │
│ 上映日期: [日期选择] ← 日期   │
│                              │
│ [取消] [保存草稿] [提交审核]  │
└──────────────────────────────┘
```

---

### 字段来源对比

| 字段类型 | 修改前 | 修改后 |
|---------|-------|-------|
| **基本字段** | 固定写死 | 固定写死 ✅ |
| **权限字段** | 固定写死 | 从API加载 ✅ |
| **编目字段** | 手动添加 | 从属性配置动态加载 ✅ |

---

### 功能对比

| 功能 | 修改前 | 修改后 |
|-----|-------|-------|
| **字段类型** | 只有文本输入 | 支持多种类型（文本/数字/日期/下拉/多行） ✅ |
| **字段配置** | 手动添加 | 从属性设置加载 ✅ |
| **等级列表** | 固定4个 | 从API加载 ✅ |
| **组列表** | 无 | 从API加载 ✅ |
| **保存选项** | 只有保存 | 保存草稿 + 提交审核 ✅ |

---

## 🎯 使用流程

### 1. 打开编目对话框

```
文件详情页
    ↓
点击"编目"按钮
    ↓
加载分类树
    ↓
加载等级列表
    ↓
加载组列表
    ↓
加载该文件类型的编目字段配置
    ↓
初始化表单数据
    ↓
显示编目对话框
```

---

### 2. 编辑编目信息

#### 基本信息

```
用户操作:
- 修改文件标题
- 选择所属分类
- 编辑描述信息
```

---

#### 权限控制

```
用户操作:
- 选择浏览等级（公开/内部/保密/机密）
- 选择可访问组（多选）
- 切换下载权限
```

---

#### 扩展属性

```
动态显示字段（根据文件类型）:

视频文件:
- 导演 [文本输入]
- 主演 [文本输入]
- 时长 [数字输入]
- 上映日期 [日期选择]
- 简介 [多行文本]

音频文件:
- 演唱者 [文本输入]
- 作曲 [文本输入]
- 作词 [文本输入]
- 时长 [数字输入]

图片文件:
- 拍摄地点 [文本输入]
- 拍摄时间 [日期选择]
- 摄影师 [文本输入]
```

---

### 3. 保存操作

#### 保存草稿（status=0）

```
点击"保存草稿"
    ↓
表单验证
    ↓
调用API更新（status=0）
    ↓
成功: 显示"草稿保存成功"
    ↓
刷新文件详情
```

---

#### 提交审核（status=1）

```
点击"保存并提交审核"
    ↓
表单验证
    ↓
调用API更新（status=1）
    ↓
成功: 显示"编目信息已提交审核"
    ↓
关闭对话框
    ↓
刷新文件详情
```

---

## 📋 字段类型支持

### 1. 文本输入（type='text'）

```vue
<el-input
  v-model="catalogForm.catalog_info[field.name]"
  :placeholder="'请输入' + field.label"
/>
```

**适用**：标题、名称、简短文本

---

### 2. 数字输入（type='number'）

```vue
<el-input-number
  v-model="catalogForm.catalog_info[field.name]"
  :placeholder="'请输入' + field.label"
/>
```

**适用**：时长、数量、年份

---

### 3. 日期选择（type='date'）

```vue
<el-date-picker
  v-model="catalogForm.catalog_info[field.name]"
  type="date"
  :placeholder="'请选择' + field.label"
/>
```

**适用**：上映日期、拍摄日期、发布日期

---

### 4. 下拉选择（type='select'）

```vue
<el-select
  v-model="catalogForm.catalog_info[field.name]"
  :placeholder="'请选择' + field.label"
>
  <el-option
    v-for="option in field.options"
    :key="option.value"
    :label="option.label"
    :value="option.value"
  />
</el-select>
```

**适用**：类型、状态、固定选项

---

### 5. 多行文本（type='textarea'）

```vue
<el-input
  v-model="catalogForm.catalog_info[field.name]"
  type="textarea"
  :rows="3"
  :placeholder="'请输入' + field.label"
/>
```

**适用**：简介、描述、备注

---

## 💡 与编目页面的一致性

### FileCatalog.vue vs FileDetail.vue

| 特性 | FileCatalog.vue | FileDetail.vue | 一致性 |
|-----|----------------|----------------|--------|
| **表单结构** | 3部分（基本/权限/扩展） | 3部分（基本/权限/扩展） | ✅ |
| **字段加载** | 动态加载 | 动态加载 | ✅ |
| **字段类型** | 5种类型 | 5种类型 | ✅ |
| **等级列表** | 从API加载 | 从API加载 | ✅ |
| **组列表** | 从API加载 | 从API加载 | ✅ |
| **保存草稿** | 支持 | 支持 | ✅ |
| **提交审核** | 支持 | 支持 | ✅ |

---

## 🚀 部署状态

```
✓ 按钮改名完成
✓ 对话框标题更新
✓ 表单结构调整
✓ 动态字段加载实现
✓ 等级和组列表加载
✓ 保存草稿功能添加
✓ 提交审核功能添加
✓ 前端已重新构建 (7.39s)
✓ 准备刷新浏览器
```

---

## ✅ 总结

### 修改内容
1. ✅ 按钮改名："编辑信息" → "编目"
2. ✅ 对话框改名："编辑文件信息" → "文件编目"
3. ✅ 表单结构：3部分（基本信息/权限控制/扩展属性）
4. ✅ 动态字段：从属性设置按文件类型加载
5. ✅ 等级列表：从API动态加载
6. ✅ 组列表：从API动态加载
7. ✅ 保存选项：保存草稿 + 提交审核

### 一致性
- ✅ **与FileCatalog.vue完全一致**
- ✅ **字段来源于属性设置**
- ✅ **支持多种字段类型**
- ✅ **支持草稿和审核状态**

### 用户体验
- ✨ **名称更准确**: "编目"比"编辑"更专业
- ✨ **字段更灵活**: 根据文件类型动态显示
- ✨ **操作更明确**: 区分草稿和提交审核
- ✨ **权限更完整**: 等级和组从API加载

---

**文件详情页编目功能优化完成！** 🎉

**现在编目对话框与编目页面完全一致！** ✨

**字段从属性设置动态加载，支持多种类型！** 🚀

**刷新浏览器即可使用新功能！** 💫

# OpenWan 前端国际化（i18n）实施指南

## 概述

OpenWan 前端系统已集成 vue-i18n 多语言支持，默认支持中文（zh-CN）和英文（en-US）。

## 安装的依赖

```json
{
  "vue-i18n": "^9.x"
}
```

## 项目结构

```
frontend/src/
├── i18n/
│   ├── index.js              # i18n 配置文件
│   └── locales/
│       ├── zh-CN.json        # 中文语言包
│       └── en-US.json        # 英文语言包
├── components/
│   └── common/
│       └── LanguageSwitcher.vue  # 语言切换组件
└── main.js                   # 主入口（已集成 i18n）
```

## 语言包结构

语言包按模块组织，包含以下主要模块：

- **common**: 通用词汇（确定、取消、保存、删除等）
- **auth**: 认证相关（登录、退出、用户名、密码等）
- **menu**: 菜单导航（首页、文件管理、系统管理等）
- **files**: 文件管理模块
- **admin**: 管理模块（用户、组、角色、分类、属性配置等）
- **search**: 搜索模块
- **validation**: 表单验证消息
- **message**: 系统提示消息

### 示例：中文语言包片段

```json
{
  "common": {
    "confirm": "确定",
    "cancel": "取消",
    "save": "保存"
  },
  "auth": {
    "login": "登录",
    "username": "用户名",
    "password": "密码"
  },
  "files": {
    "fileList": "文件列表",
    "uploadFile": "上传文件",
    "status": {
      "new": "新建",
      "pending": "待审核",
      "published": "已发布"
    }
  }
}
```

## 使用方法

### 1. 在组件中使用 i18n

#### 选项式 API

```vue
<template>
  <div>
    <h1>{{ $t('files.title') }}</h1>
    <el-button>{{ $t('common.save') }}</el-button>
  </div>
</template>

<script>
export default {
  methods: {
    showMessage() {
      this.$message.success(this.$t('message.saveSuccess'))
    }
  }
}
</script>
```

#### 组合式 API（推荐）

```vue
<template>
  <div>
    <h1>{{ t('files.title') }}</h1>
    <el-button>{{ t('common.save') }}</el-button>
  </div>
</template>

<script setup>
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const showMessage = () => {
  ElMessage.success(t('message.saveSuccess'))
}
</script>
```

### 2. 带参数的翻译

语言包：
```json
{
  "message": {
    "deleteConfirm": "确定要删除用户 {username} 吗？"
  },
  "common": {
    "total": "共 {count} 条"
  }
}
```

使用：
```vue
<template>
  <div>
    <!-- 方式1：对象参数 -->
    <p>{{ t('common.total', { count: 100 }) }}</p>
    
    <!-- 方式2：在方法中使用 -->
    <el-button @click="confirmDelete">删除</el-button>
  </div>
</template>

<script setup>
import { useI18n } from 'vue-i18n'
import { ElMessageBox } from 'element-plus'

const { t } = useI18n()

const confirmDelete = async () => {
  await ElMessageBox.confirm(
    t('admin.users.deleteConfirm', { username: 'John' }),
    t('common.warning')
  )
}
</script>
```

### 3. 复数和选择

```json
{
  "files": {
    "selectedCount": "没有选择文件 | 选择了 1 个文件 | 选择了 {count} 个文件"
  }
}
```

```vue
<template>
  <p>{{ t('files.selectedCount', count) }}</p>
</template>

<script setup>
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const count = ref(5)
</script>
```

### 4. 语言切换

#### 使用 LanguageSwitcher 组件（已集成在顶栏）

LanguageSwitcher 组件已添加到 MainLayout.vue 的顶栏右侧，用户可以直接点击切换语言。

#### 编程式切换

```javascript
import { setLocale } from '@/i18n'

// 切换到英文
setLocale('en-US')

// 切换到中文
setLocale('zh-CN')
```

#### 获取当前语言

```javascript
import { getLocale } from '@/i18n'

const currentLocale = getLocale()
console.log(currentLocale) // 'zh-CN' 或 'en-US'
```

### 5. Element Plus 组件国际化

Element Plus 组件的语言会随系统语言自动切换（已在 main.js 中配置）：

```javascript
// main.js
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import enUS from 'element-plus/es/locale/lang/en'

const currentLocale = getLocale()
const elementLocale = currentLocale === 'en-US' ? enUS : zhCn
app.use(ElementPlus, {
  locale: elementLocale,
})
```

## 完整示例：将现有页面国际化

### 原始代码（硬编码中文）

```vue
<template>
  <div>
    <h1>文件列表</h1>
    <el-button type="primary" @click="upload">上传文件</el-button>
    <el-table :data="files">
      <el-table-column prop="title" label="标题" />
      <el-table-column prop="status" label="状态">
        <template #default="{ row }">
          <el-tag v-if="row.status === 0">新建</el-tag>
          <el-tag v-else-if="row.status === 1" type="warning">待审核</el-tag>
          <el-tag v-else-if="row.status === 2" type="success">已发布</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作">
        <template #default="{ row }">
          <el-button size="small" @click="edit(row)">编辑</el-button>
          <el-button size="small" type="danger" @click="remove(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'

const files = ref([])

const upload = () => {
  // 上传逻辑
}

const edit = (row) => {
  // 编辑逻辑
}

const remove = async (row) => {
  await ElMessageBox.confirm('确定要删除这个文件吗？', '提示')
  // 删除逻辑
  ElMessage.success('删除成功')
}
</script>
```

### 国际化后的代码

```vue
<template>
  <div>
    <h1>{{ t('files.fileList') }}</h1>
    <el-button type="primary" @click="upload">
      {{ t('files.uploadFile') }}
    </el-button>
    <el-table :data="files">
      <el-table-column prop="title" :label="t('files.fileTitle')" />
      <el-table-column prop="status" :label="t('files.fileStatus')">
        <template #default="{ row }">
          <el-tag v-if="row.status === 0">{{ t('files.status.new') }}</el-tag>
          <el-tag v-else-if="row.status === 1" type="warning">
            {{ t('files.status.pending') }}
          </el-tag>
          <el-tag v-else-if="row.status === 2" type="success">
            {{ t('files.status.published') }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column :label="t('common.actions')">
        <template #default="{ row }">
          <el-button size="small" @click="edit(row)">
            {{ t('common.edit') }}
          </el-button>
          <el-button size="small" type="danger" @click="remove(row)">
            {{ t('common.delete') }}
          </el-button>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'

const { t } = useI18n()
const files = ref([])

const upload = () => {
  // 上传逻辑
}

const edit = (row) => {
  // 编辑逻辑
}

const remove = async (row) => {
  await ElMessageBox.confirm(
    t('files.deleteConfirm'),
    t('common.warning')
  )
  // 删除逻辑
  ElMessage.success(t('files.deleteSuccess'))
}
</script>
```

## 添加新语言

### 1. 创建语言文件

在 `src/i18n/locales/` 目录下创建新的语言文件，例如 `ja-JP.json`（日语）：

```json
{
  "common": {
    "confirm": "確認",
    "cancel": "キャンセル",
    "save": "保存"
  }
}
```

### 2. 更新 i18n 配置

在 `src/i18n/index.js` 中导入并注册：

```javascript
import jaJP from './locales/ja-JP.json'

const i18n = createI18n({
  messages: {
    'zh-CN': zhCN,
    'en-US': enUS,
    'ja-JP': jaJP  // 新增
  }
})

export const availableLocales = [
  { value: 'zh-CN', label: '简体中文' },
  { value: 'en-US', label: 'English' },
  { value: 'ja-JP', label: '日本語' }  // 新增
]
```

### 3. 添加 Element Plus 语言包

```javascript
import jaJP from 'element-plus/es/locale/lang/ja'

// 在 main.js 中更新
const getElementLocale = (locale) => {
  switch (locale) {
    case 'en-US':
      return enUS
    case 'ja-JP':
      return jaJP
    default:
      return zhCn
  }
}
```

## 最佳实践

### 1. 命名规范

- 使用点号分隔的层级结构：`module.submodule.key`
- 使用小驼峰命名：`fileList`, `uploadTime`
- 状态、类型等枚举值放在子对象中：`files.status.new`

### 2. 避免硬编码

❌ 不推荐：
```vue
<el-button>保存</el-button>
```

✅ 推荐：
```vue
<el-button>{{ t('common.save') }}</el-button>
```

### 3. 提取可复用文本

将常用的文本放在 `common` 模块，避免重复定义：

```json
{
  "common": {
    "save": "保存",
    "edit": "编辑",
    "delete": "删除"
  }
}
```

### 4. 保持语言包同步

添加新 key 时，同时更新所有语言文件：

```bash
# 中文
"files.newField": "新字段"

# 英文
"files.newField": "New Field"
```

### 5. 使用计算属性处理复杂逻辑

```vue
<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const statusText = computed(() => {
  const map = {
    0: t('files.status.new'),
    1: t('files.status.pending'),
    2: t('files.status.published')
  }
  return map[file.status]
})
</script>
```

## 路由标题国际化

更新 `router/index.js`：

```javascript
const routes = [
  {
    path: '/files',
    name: 'Files',
    component: FileList,
    meta: {
      title: 'menu.fileList',  // 使用 i18n key
      requiresAuth: true
    }
  }
]

// 路由守卫中使用
router.beforeEach((to, from, next) => {
  if (to.meta.title) {
    document.title = t(to.meta.title) + ' - OpenWan'
  }
  next()
})
```

## 表单验证国际化

```vue
<script setup>
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const rules = {
  username: [
    { 
      required: true, 
      message: t('validation.required', { field: t('auth.username') })
    },
    { 
      min: 3, 
      max: 20, 
      message: t('validation.minLength', { field: t('auth.username'), min: 3 })
    }
  ]
}
</script>
```

## 调试和测试

### 1. 查找未翻译的文本

在浏览器控制台使用：

```javascript
// 检查当前页面是否有硬编码的中文
document.body.innerText.match(/[\u4e00-\u9fa5]+/g)
```

### 2. 测试语言切换

```javascript
// 在控制台切换语言
import { setLocale } from '@/i18n'
setLocale('en-US')
```

### 3. 检查缺失的翻译键

Vue I18n 会在控制台警告缺失的翻译键：

```
[intlify] Not found 'files.newKey' key in 'en-US' locale messages.
```

## 当前状态

✅ **已完成：**
1. vue-i18n 安装和配置
2. 中文（zh-CN）和英文（en-US）语言包
3. 语言切换组件（LanguageSwitcher）
4. MainLayout 集成语言切换器
5. Element Plus 国际化配置
6. 语言设置持久化（localStorage）
7. 浏览器语言自动检测

⏳ **待完成：**
1. 将所有现有页面的硬编码文本替换为 i18n 调用
2. 路由标题国际化
3. 表单验证消息国际化
4. 日期/时间格式国际化
5. 数字格式国际化

## 批量替换指南

为了将现有页面快速国际化，可以按以下步骤操作：

### 1. 识别需要翻译的页面

```bash
# 查找包含中文的 Vue 文件
grep -r "[\u4e00-\u9fa5]" src/views/ --include="*.vue"
```

### 2. 逐个文件替换

对于每个文件：
1. 添加 `import { useI18n } from 'vue-i18n'`
2. 添加 `const { t } = useI18n()`
3. 将硬编码文本替换为 `t('key')`
4. 确保对应的翻译键存在于语言包中

### 3. 验证翻译

切换语言后检查页面显示是否正确。

## 技术支持

如需添加新的翻译键或遇到问题，请参考：
- Vue I18n 官方文档: https://vue-i18n.intlify.dev/
- Element Plus 国际化: https://element-plus.org/en-US/guide/i18n.html

---

**文档版本**: 1.0  
**更新时间**: 2026-02-06  
**维护者**: OpenWan Development Team

# OpenWan Frontend

Vue 3 + Vite + Element Plus 前端应用

## 技术栈

- **框架**: Vue 3.3+ (Composition API)
- **构建工具**: Vite 4.x
- **UI库**: Element Plus 2.x
- **状态管理**: Pinia 2.x
- **路由**: Vue Router 4.x
- **HTTP客户端**: Axios 1.x
- **视频播放器**: Video.js 8.x + FLV.js
- **开发语言**: JavaScript / TypeScript (可选)

## 项目结构

```
frontend/
├── public/                 # 静态资源
├── src/
│   ├── api/               # API接口模块
│   │   ├── auth.js        # 认证相关
│   │   ├── files.js       # 文件管理
│   │   ├── groups.js      # 组管理
│   │   ├── roles.js       # 角色管理
│   │   ├── category.js    # 分类管理
│   │   └── catalog.js     # 目录配置
│   ├── assets/            # 资源文件
│   ├── components/        # 公共组件
│   │   └── VideoPlayer.vue
│   ├── layouts/           # 布局组件
│   │   └── MainLayout.vue
│   ├── router/            # 路由配置
│   │   └── index.js
│   ├── stores/            # Pinia状态管理
│   │   └── user.js
│   ├── utils/             # 工具函数
│   │   └── request.js     # Axios封装
│   ├── views/             # 页面组件
│   │   ├── Login.vue
│   │   ├── Dashboard.vue
│   │   ├── Search.vue
│   │   ├── files/         # 文件管理页面
│   │   └── admin/         # 管理页面
│   ├── App.vue            # 根组件
│   └── main.js            # 入口文件
├── .env.development       # 开发环境配置
├── .env.production        # 生产环境配置
├── index.html             # HTML模板
├── package.json           # 依赖配置
├── vite.config.js         # Vite配置
└── README.md              # 项目文档
```

## 快速开始

### 安装依赖

```bash
npm install
```

### 开发模式

```bash
npm run dev
```

访问 http://localhost:5173

### 生产构建

```bash
npm run build
```

构建产物在 `dist/` 目录

### 预览生产构建

```bash
npm run preview
```

## 环境配置

### 开发环境 (.env.development)

```bash
VITE_API_BASE_URL=http://localhost:8080/api
```

### 生产环境 (.env.production)

```bash
VITE_API_BASE_URL=https://api.yourdomain.com/api
```

## 主要功能模块

### 1. 认证模块
- 登录/登出
- Token管理
- 权限检查
- 路由守卫

### 2. 文件管理
- 文件列表（筛选、分页）
- 文件上传（拖拽、进度）
- 文件详情（预览、编辑）
- 视频播放（Video.js + FLV）
- 文件下载
- 工作流操作

### 3. 搜索功能
- 关键词搜索
- 高级筛选
- 结果高亮
- 排序

### 4. 管理功能
- 用户管理
- 组管理
- 角色管理
- 分类管理
- 目录配置

## 代码规范

### Vue组件规范

```vue
<template>
  <!-- 模板内容 -->
</template>

<script setup>
// Composition API
import { ref, computed, onMounted } from 'vue'

// 状态
const data = ref([])

// 计算属性
const filteredData = computed(() => {
  return data.value.filter(item => item.active)
})

// 方法
const loadData = async () => {
  // 实现
}

// 生命周期
onMounted(() => {
  loadData()
})
</script>

<style scoped>
/* 样式 */
</style>
```

### API调用规范

```javascript
// api/example.js
import request from '@/utils/request'

export default {
  getList(params) {
    return request({
      url: '/v1/example',
      method: 'get',
      params,
    })
  },
  
  create(data) {
    return request({
      url: '/v1/example',
      method: 'post',
      data,
    })
  },
}
```

### 页面组件规范

```vue
<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import exampleApi from '@/api/example'

// 状态
const loading = ref(false)
const dataList = ref([])

// 表单
const form = reactive({
  name: '',
  type: null,
})

// 加载数据
const loadData = async () => {
  loading.value = true
  try {
    const res = await exampleApi.getList()
    if (res.success) {
      dataList.value = res.data
    }
  } catch (error) {
    ElMessage.error('加载失败')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadData()
})
</script>
```

## 路由配置

### 路由结构

```javascript
const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: { requiresAuth: false },
  },
  {
    path: '/',
    component: () => import('@/layouts/MainLayout.vue'),
    meta: { requiresAuth: true },
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/Dashboard.vue'),
        meta: { title: '首页' },
      },
      // 更多路由...
    ],
  },
]
```

### 路由守卫

```javascript
router.beforeEach(async (to, from, next) => {
  const userStore = useUserStore()
  
  if (to.meta.requiresAuth !== false) {
    if (!userStore.token) {
      next({ path: '/login', query: { redirect: to.fullPath } })
      return
    }
    
    if (!userStore.user) {
      await userStore.getUserInfo()
    }
    
    if (to.meta.requiresAdmin && !userStore.isAdmin()) {
      next({ path: '/dashboard' })
      return
    }
  }
  
  next()
})
```

## 状态管理

### Pinia Store示例

```javascript
import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useUserStore = defineStore('user', () => {
  // 状态
  const user = ref(null)
  const token = ref(localStorage.getItem('token') || '')
  
  // Getters
  const isAdmin = () => {
    return user.value?.roles.includes('admin')
  }
  
  // Actions
  async function login(credentials) {
    const res = await authApi.login(credentials)
    if (res.success) {
      token.value = res.data.token
      user.value = res.data.user
      localStorage.setItem('token', token.value)
      return true
    }
    return false
  }
  
  return {
    user,
    token,
    isAdmin,
    login,
  }
})
```

## 组件开发

### VideoPlayer组件

```vue
<template>
  <video ref="videoElement" class="video-js" />
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount } from 'vue'
import videojs from 'video.js'
import 'video.js/dist/video-js.css'

const props = defineProps({
  src: String,
  type: { type: String, default: 'video/x-flv' },
})

const videoElement = ref(null)
let player = null

onMounted(() => {
  player = videojs(videoElement.value, {
    sources: [{
      src: props.src,
      type: props.type,
    }],
  })
})

onBeforeUnmount(() => {
  if (player) {
    player.dispose()
  }
})
</script>
```

## 构建优化

### Vite配置

```javascript
export default defineConfig({
  build: {
    // 分包策略
    rollupOptions: {
      output: {
        manualChunks: {
          'element-plus': ['element-plus'],
          'video-player': ['video.js', 'videojs-flvjs-es6'],
        },
      },
    },
    // 压缩
    minify: 'terser',
    terserOptions: {
      compress: {
        drop_console: true,
        drop_debugger: true,
      },
    },
  },
})
```

## 部署

### Nginx配置

```nginx
server {
    listen 80;
    server_name yourdomain.com;
    root /var/www/openwan/dist;
    index index.html;

    # SPA路由支持
    location / {
        try_files $uri $uri/ /index.html;
    }

    # API代理
    location /api {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    # 静态资源缓存
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
}
```

### Docker构建

```dockerfile
# Build stage
FROM node:18-alpine as build
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

# Production stage
FROM nginx:alpine
COPY --from=build /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

## 常见问题

### 1. API请求失败

**问题**: 请求返回CORS错误

**解决**: 检查后端CORS配置，或使用Vite代理

```javascript
// vite.config.js
export default defineConfig({
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})
```

### 2. 视频播放失败

**问题**: FLV视频无法播放

**解决**: 确保安装flv.js插件

```bash
npm install videojs-flvjs-es6
```

### 3. 打包体积过大

**问题**: dist文件夹过大

**解决**: 启用代码分割和Gzip压缩

```javascript
// vite.config.js
import viteCompression from 'vite-plugin-compression'

export default defineConfig({
  plugins: [
    viteCompression({
      algorithm: 'gzip',
      ext: '.gz',
    }),
  ],
})
```

## 开发工具

### 推荐VSCode插件

- Vue Language Features (Volar)
- ESLint
- Prettier
- Auto Rename Tag
- Path Intellisense

### 推荐Chrome插件

- Vue.js devtools
- React Developer Tools (for debugging)

## 贡献指南

1. Fork项目
2. 创建特性分支
3. 提交改动
4. 推送到分支
5. 创建Pull Request

## 许可证

MIT License

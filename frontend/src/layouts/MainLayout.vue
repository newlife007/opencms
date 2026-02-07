<template>
  <el-container class="main-layout">
    <!-- 顶栏：全屏贯穿 -->
    <el-header class="header">
      <div class="header-left">
        <div class="logo">
          <h2>OpenWan</h2>
        </div>
        <el-icon class="collapse-icon" @click="toggleCollapse">
          <Fold v-if="!isCollapse" />
          <Expand v-else />
        </el-icon>
        <el-breadcrumb separator="/">
          <el-breadcrumb-item :to="{ path: '/' }">{{ t('menu.home') }}</el-breadcrumb-item>
          <el-breadcrumb-item v-for="item in breadcrumbs" :key="item.path">
            {{ t(item.meta?.title) }}
          </el-breadcrumb-item>
        </el-breadcrumb>
      </div>
      <div class="header-right">
        <!-- 语言切换器 -->
        <LanguageSwitcher style="margin-right: 20px;" />
        
        <el-dropdown>
          <span class="user-info">
            <el-avatar :size="32" icon="UserFilled" />
            <span class="username">{{ userStore.user?.username }}</span>
          </span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item @click="handleChangePassword">
                <el-icon><EditPen /></el-icon>
                {{ t('auth.changePassword') }}
              </el-dropdown-item>
              <el-dropdown-item divided @click="handleLogout">
                <el-icon><SwitchButton /></el-icon>
                {{ t('auth.logout') }}
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </el-header>

    <!-- 下半部分容器 -->
    <el-container class="content-container">
      <!-- 侧边导航栏：白色背景，从顶栏下方到底部 -->
      <el-aside :width="isCollapse ? '64px' : '200px'" class="sidebar">
        <el-menu
          :default-active="activeMenu"
          class="sidebar-menu"
          router
          :collapse="isCollapse"
        >
          <template v-for="route in menuRoutes" :key="route.path">
            <el-sub-menu v-if="route.children && route.children.length" :index="'/' + route.path">
              <template #title>
                <el-icon v-if="route.meta?.icon"><component :is="route.meta.icon" /></el-icon>
                <span>{{ t(route.meta?.title) }}</span>
              </template>
              <el-menu-item
                v-for="child in route.children"
                :key="child.path"
                :index="'/' + route.path + '/' + child.path"
              >
                <el-icon v-if="child.meta?.icon"><component :is="child.meta.icon" /></el-icon>
                <span>{{ t(child.meta?.title) }}</span>
              </el-menu-item>
            </el-sub-menu>
            <el-menu-item v-else :index="'/' + route.path">
              <el-icon v-if="route.meta?.icon"><component :is="route.meta.icon" /></el-icon>
              <span>{{ t(route.meta?.title) }}</span>
            </el-menu-item>
          </template>
        </el-menu>
      </el-aside>

      <!-- 主内容区域 -->
      <el-main class="main-content">
        <router-view v-slot="{ Component }">
          <transition name="fade" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { ElMessageBox, ElMessage } from 'element-plus'
import LanguageSwitcher from '@/components/common/LanguageSwitcher.vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const router = useRouter()
const route = useRoute()
const userStore = useUserStore()

const isCollapse = ref(false)

const toggleCollapse = () => {
  isCollapse.value = !isCollapse.value
}

const activeMenu = computed(() => route.path)

const menuRoutes = computed(() => {
  const routes = router.options.routes.find(r => r.path === '/')?.children || []
  
  return routes.filter(route => {
    // 隐藏的路由不显示
    if (route.meta?.hidden) return false
    
    // 管理员模块检查
    if (route.meta?.requiresAdmin && !userStore.isAdmin()) return false
    
    // 权限检查：检查用户是否有任何一个所需权限
    if (route.meta?.permissions && route.meta.permissions.length > 0) {
      const hasPermission = route.meta.permissions.some(permission => 
        userStore.hasPermission(permission)
      )
      if (!hasPermission) return false
    }
    
    // 如果有子路由，递归过滤
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
      
      // 如果过滤后没有可见的子路由，隐藏父路由
      if (filteredChildren.length === 0) return false
      
      // 更新路由的子路由列表
      route.children = filteredChildren
    }
    
    return true
  })
})

const breadcrumbs = computed(() => {
  return route.matched.filter(r => r.meta?.title)
})

const handleChangePassword = () => {
  router.push('/change-password')
}

const handleLogout = async () => {
  try {
    await ElMessageBox.confirm('确定要退出登录吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    })
    
    await userStore.logout()
    ElMessage.success('已退出登录')
    router.push('/login')
  } catch (error) {
    // User cancelled
  }
}
</script>

<style scoped>
/* 主容器：垂直布局 */
.main-layout {
  height: 100vh;
  display: flex;
  flex-direction: column;
}

/* 顶栏：全屏贯穿，固定高度 */
.header {
  height: 60px !important;
  display: flex;
  align-items: center;
  justify-content: space-between;
  background-color: #fff;
  border-bottom: 1px solid #e6e6e6;
  padding: 0 20px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.08);
  z-index: 1000;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 20px;
}

/* Logo 在顶栏左侧 */
.logo {
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 10px;
}

.logo h2 {
  margin: 0;
  font-size: 20px;
  font-weight: bold;
  color: #409eff;
  background: linear-gradient(135deg, #409eff 0%, #0066cc 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.collapse-icon {
  font-size: 20px;
  cursor: pointer;
  transition: transform 0.3s;
}

.collapse-icon:hover {
  color: #409eff;
  transform: scale(1.1);
}

.header-right {
  display: flex;
  align-items: center;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
  padding: 5px 10px;
  border-radius: 4px;
  transition: background-color 0.3s;
}

.user-info:hover {
  background-color: #f5f7fa;
}

.username {
  font-size: 14px;
  color: #333;
}

/* 下半部分容器：占据剩余空间 */
.content-container {
  flex: 1;
  overflow: hidden;
}

/* 侧边导航栏：白色背景 */
.sidebar {
  background-color: #fff;
  border-right: 1px solid #e6e6e6;
  overflow-x: hidden;
  overflow-y: auto;
  transition: width 0.3s;
}

/* 滚动条样式 */
.sidebar::-webkit-scrollbar {
  width: 6px;
}

.sidebar::-webkit-scrollbar-thumb {
  background-color: #dcdfe6;
  border-radius: 3px;
}

.sidebar::-webkit-scrollbar-thumb:hover {
  background-color: #c0c4cc;
}

.sidebar-menu {
  border: none;
  background-color: #fff;
}

/* 菜单项样式：使用深色文字 */
.sidebar-menu :deep(.el-menu-item),
.sidebar-menu :deep(.el-sub-menu__title) {
  color: #303133;
  font-size: 14px;
}

.sidebar-menu :deep(.el-menu-item:hover),
.sidebar-menu :deep(.el-sub-menu__title:hover) {
  background-color: #ecf5ff !important;
  color: #409eff;
}

.sidebar-menu :deep(.el-menu-item.is-active) {
  background-color: #409eff !important;
  color: #fff;
}

.sidebar-menu :deep(.el-sub-menu .el-menu-item) {
  padding-left: 50px !important;
  min-width: auto;
}

.sidebar-menu :deep(.el-sub-menu .el-menu-item:hover) {
  background-color: #ecf5ff !important;
  color: #409eff;
}

.sidebar-menu :deep(.el-sub-menu .el-menu-item.is-active) {
  background-color: #409eff !important;
  color: #fff;
}

/* 折叠状态下的菜单 */
.sidebar-menu.el-menu--collapse {
  width: 64px;
}

.sidebar-menu.el-menu--collapse :deep(.el-menu-item),
.sidebar-menu.el-menu--collapse :deep(.el-sub-menu__title) {
  padding: 0 20px;
}

/* 主内容区域 */
.main-content {
  background-color: #f5f5f5;
  padding: 20px;
  overflow-y: auto;
}

/* 面包屑样式 */
:deep(.el-breadcrumb__inner) {
  color: #606266;
  font-weight: normal;
}

:deep(.el-breadcrumb__inner:hover) {
  color: #409eff;
}

:deep(.el-breadcrumb__inner.is-link) {
  color: #409eff;
  font-weight: 500;
}

/* 页面切换动画 */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>

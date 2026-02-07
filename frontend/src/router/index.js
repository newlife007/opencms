import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/test-video',
    name: 'TestVideoPlayer',
    component: () => import('@/views/TestVideoPlayer.vue'),
    meta: { requiresAuth: false },
  },
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
    redirect: '/dashboard',
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/Dashboard.vue'),
        meta: { title: 'menu.home', icon: 'HomeFilled' },
      },
      {
        path: 'files',
        name: 'Files',
        component: () => import('@/views/files/FileList.vue'),
        meta: { 
          title: 'menu.fileList', 
          icon: 'Document',
          permissions: ['files.browse.list']  // 修改为数据库中的权限格式
        },
      },
      {
        path: 'files/upload',
        name: 'FileUpload',
        component: () => import('@/views/files/FileUpload.vue'),
        meta: { 
          title: 'menu.fileUpload', 
          icon: 'Upload',
          permissions: ['files.upload.create']  // 文件上传权限
        },
      },
      {
        path: 'files/:id(\\d+)',
        name: 'FileDetail',
        component: () => import('@/views/files/FileDetail.vue'),
        meta: { 
          title: 'files.fileDetail', 
          hidden: true,
          permissions: ['files.browse.view']  // 修改为数据库中的权限格式
        },
      },
      {
        path: 'files/:id(\\d+)/simple',
        name: 'FileDetailSimple',
        component: () => import('@/views/files/FileDetailSimple.vue'),
        meta: { 
          title: 'files.fileDetail', 
          hidden: true,
          permissions: ['files.browse.view']  // 修改为数据库中的权限格式
        },
      },
      {
        path: 'files/:id(\\d+)/catalog',
        name: 'FileCatalog',
        component: () => import('@/views/files/FileCatalog.vue'),
        meta: { 
          title: 'files.fileCatalog', 
          icon: 'Edit',
          hidden: true,
          permissions: ['files.catalog.edit']  // 文件编目权限
        },
      },
      {
        path: 'approval',
        name: 'FileApproval',
        component: () => import('@/views/files/FileApproval.vue'),
        meta: { 
          title: 'menu.fileApproval', 
          icon: 'CircleCheck',
          permissions: ['files.putout.approve']  // 文件审批权限
        },
      },
      {
        path: 'search',
        name: 'Search',
        component: () => import('@/views/Search.vue'),
        meta: { 
          title: 'menu.search', 
          icon: 'Search',
          permissions: ['files.browse.search']  // 修改为数据库中的权限格式
        },
      },
      {
        path: 'admin',
        name: 'Admin',
        redirect: '/admin/users',
        meta: { 
          title: 'menu.admin', 
          icon: 'Setting', 
          requiresAdmin: true,
          permissions: []  // 管理员模块，由子路由检查权限
        },
        children: [
          {
            path: 'users',
            name: 'AdminUsers',
            component: () => import('@/views/admin/Users.vue'),
            meta: { 
              title: 'menu.users', 
              icon: 'User',
              permissions: ['users.manage.view']  // 需要用户管理权限
            },
          },
          {
            path: 'groups',
            name: 'AdminGroups',
            component: () => import('@/views/admin/Groups.vue'),
            meta: { 
              title: 'menu.groups', 
              icon: 'UserFilled',
              permissions: ['groups.manage.view']  // 需要组管理权限
            },
          },
          {
            path: 'roles',
            name: 'AdminRoles',
            component: () => import('@/views/admin/Roles.vue'),
            meta: { 
              title: 'menu.roles', 
              icon: 'Avatar',
              permissions: ['roles.manage.view']  // 需要角色管理权限
            },
          },
          {
            path: 'permissions',
            name: 'AdminPermissions',
            component: () => import('@/views/admin/Permissions.vue'),
            meta: { 
              title: 'menu.permissions', 
              icon: 'Lock',
              permissions: ['permissions.manage.view']  // 需要权限管理权限
            },
          },
          {
            path: 'levels',
            name: 'AdminLevels',
            component: () => import('@/views/admin/Levels.vue'),
            meta: { 
              title: 'menu.levels', 
              icon: 'TrendCharts',
              permissions: ['levels.manage.view']  // 需要等级管理权限
            },
          },
          {
            path: 'categories',
            name: 'AdminCategories',
            component: () => import('@/views/admin/Categories.vue'),
            meta: { 
              title: 'menu.categories', 
              icon: 'FolderOpened',
              permissions: ['categories.manage.create']  // 需要分类管理权限
            },
          },
          {
            path: 'catalog',
            name: 'AdminCatalog',
            component: () => import('@/views/admin/Catalog.vue'),
            meta: { 
              title: 'menu.catalog', 
              icon: 'List',
              permissions: ['catalog.config.view']  // 需要属性配置权限
            },
          },
        ],
      },
    ],
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

// Navigation guard
router.beforeEach(async (to, from, next) => {
  // Check for null/undefined from parameter (initial navigation)
  if (!from || from.name === null) {
    // Initial navigation, skip certain checks
    if (to.meta.requiresAuth === false) {
      next()
      return
    }
  }

  // Import store dynamically to avoid circular dependency
  const { useUserStore } = await import('@/stores/user')
  const userStore = useUserStore()
  const token = userStore.token

  // Check if route requires authentication
  if (to.meta.requiresAuth !== false) {
    if (!token) {
      next({ path: '/login', query: { redirect: to.fullPath } })
      return
    }

    // Get user info if not loaded
    if (!userStore.user) {
      const success = await userStore.getUserInfo()
      if (!success) {
        next({ path: '/login', query: { redirect: to.fullPath } })
        return
      }
    }

    // Check admin permission
    if (to.meta.requiresAdmin && !userStore.isAdmin()) {
      next({ path: '/dashboard' })
      return
    }
  } else if (token && to.path === '/login') {
    // Already logged in, redirect to dashboard
    next({ path: '/dashboard' })
    return
  }

  next()
})

export default router

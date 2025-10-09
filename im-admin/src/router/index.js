import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '@/stores/user'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/',
    component: () => import('@/layout/index.vue'),
    meta: { requiresAuth: true },
    children: [
      {
        path: '',
        name: 'Dashboard',
        component: () => import('@/views/Dashboard.vue'),
        meta: { title: '仪表盘' }
      },
      {
        path: '/users',
        name: 'Users',
        component: () => import('@/views/Users.vue'),
        meta: { title: '用户管理' }
      },
      {
        path: '/chats',
        name: 'Chats',
        component: () => import('@/views/Chats.vue'),
        meta: { title: '聊天管理' }
      },
      {
        path: '/messages',
        name: 'Messages',
        component: () => import('@/views/Messages.vue'),
        meta: { title: '消息管理' }
      },
      {
        path: '/system',
        name: 'System',
        component: () => import('@/views/System.vue'),
        meta: { title: '系统管理' }
      },
      {
        path: '/logs',
        name: 'Logs',
        component: () => import('@/views/Logs.vue'),
        meta: { title: '日志管理' }
      },
      {
        path: '/plugins',
        name: 'PluginManagement',
        component: () => import('@/views/PluginManagement.vue'),
        meta: { title: '插件管理' }
      },
      {
        path: '/security/2fa',
        name: 'TwoFactorSettings',
        component: () => import('@/views/TwoFactorSettings.vue'),
        meta: { title: '双因子认证' }
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由守卫
router.beforeEach((to, from, next) => {
  const userStore = useUserStore()
  
  if (to.meta.requiresAuth && !userStore.isLoggedIn) {
    next('/login')
  } else if (to.path === '/login' && userStore.isLoggedIn) {
    next('/')
  } else {
    next()
  }
})

export default router

import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    redirect: '/dashboard',
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/Login.vue'),
    meta: { title: '登录', noAuth: true },
  },
  {
    path: '/dashboard',
    name: 'Dashboard',
    component: () => import('../views/Dashboard.vue'),
    meta: { title: '仪表板', icon: 'Dashboard' },
  },
  {
    path: '/super-admin',
    name: 'SuperAdmin',
    component: () => import('../views/SuperAdmin.vue'),
    meta: { title: '超级管理后台', icon: 'Setting', role: 'super_admin' },
  },
  {
    path: '/users',
    name: 'Users',
    component: () => import('../views/Users.vue'),
    meta: { title: '用户管理', icon: 'User' },
  },
  {
    path: '/groups',
    name: 'Groups',
    component: () => import('../views/Groups.vue'),
    meta: { title: '群组管理', icon: 'ChatDotRound' },
  },
  {
    path: '/messages',
    name: 'Messages',
    component: () => import('../views/Messages.vue'),
    meta: { title: '消息管理', icon: 'ChatLineRound' },
  },
  {
    path: '/files',
    name: 'Files',
    component: () => import('../views/Files.vue'),
    meta: { title: '文件管理', icon: 'Folder' },
  },
  {
    path: '/moderation',
    name: 'Moderation',
    component: () => import('../views/Moderation.vue'),
    meta: { title: '内容审核', icon: 'View' },
  },
  {
    path: '/analytics',
    name: 'Analytics',
    component: () => import('../views/Analytics.vue'),
    meta: { title: '数据分析', icon: 'DataAnalysis' },
  },
  {
    path: '/settings',
    name: 'Settings',
    component: () => import('../views/Settings.vue'),
    meta: { title: '系统设置', icon: 'Setting' },
  },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

// 路由守卫
router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('admin_token');
  const userRole = localStorage.getItem('user_role');

  // 不需要认证的页面
  if (to.meta.noAuth) {
    next();
    return;
  }

  // 检查是否已登录
  if (!token) {
    next('/login');
    return;
  }

  // 检查角色权限
  if (to.meta.role && userRole !== to.meta.role) {
    next('/dashboard');
    return;
  }

  next();
});

export default router;

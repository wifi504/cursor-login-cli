import type { RouteRecordRaw } from 'vue-router'
/**
 * Vue Router 路由配置自动生成(v1.1)
 * @author WIFI连接超时
 * 请不要编辑此文件，因为在重新运行生成脚本后会覆盖，自定义配置请写在 options.ts 中
 */
export default [
  {
    path: '/',
    component: () => import('@/views/Index.vue'),
  },
  {
    path: '/accounts',
    component: () => import('@/views/accounts/Index.vue'),
    children: [
      {
        path: 'edit',
        component: () => import('@/views/accounts/edit/Index.vue'),
        name: 'accounts-edit',
        meta: {
          title: '编辑账号',
        },
      },
      {
        path: 'new',
        component: () => import('@/views/accounts/new/Index.vue'),
        name: 'accounts-new',
        meta: {
          title: '新建账号',
        },
      },
    ],
    name: 'accounts',
    meta: {
      title: '账号管理',
    },
  },
  {
    path: '/home',
    component: () => import('@/views/home/Index.vue'),
    name: 'home',
    meta: {
      title: '概览',
    },
  },
  {
    path: '/login',
    component: () => import('@/views/login/Index.vue'),
    name: 'login',
    meta: {
      title: '登录',
    },
  },
  {
    path: '/settings',
    component: () => import('@/views/settings/Index.vue'),
    name: 'settings',
    meta: {
      title: '系统设置',
    },
  },
  {
    path: '/setup',
    component: () => import('@/views/setup/Index.vue'),
    name: 'setup',
    meta: {
      title: '初始化',
    },
  },
] as RouteRecordRaw[]

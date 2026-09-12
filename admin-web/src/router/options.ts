import type { RouteRecordOption } from '@ezview/gen-vue-routes'

/**
 * 由 `pnpm gen-routes` 合并的路由增量配置（name / meta 等）。
 * 注意：部分 gen-vue-routes 版本若给 `/` 单独写 name，可能生成重复空路由；
 * 根路由可用路径跳转，或生成后再手工合并。
 *
 * 后台壳路由在 router/index.ts 中组装，这里只提供 name/meta。
 */
export default [
  {
    path: '/setup',
    name: 'setup',
    meta: {
      title: '初始化',
    },
  },
  {
    path: '/login',
    name: 'login',
    meta: {
      title: '登录',
    },
  },
  {
    path: '/home',
    name: 'home',
    meta: {
      title: '概览',
    },
  },
  {
    path: '/settings',
    name: 'settings',
    meta: {
      title: '系统设置',
    },
  },
  {
    path: '/accounts',
    name: 'accounts',
    meta: {
      title: '账号管理',
    },
  },
  {
    path: '/accounts/new',
    name: 'accounts-new',
    meta: {
      title: '新建账号',
    },
  },
  {
    path: '/accounts/edit',
    name: 'accounts-edit',
    meta: {
      title: '编辑账号',
    },
  },
] as RouteRecordOption[]

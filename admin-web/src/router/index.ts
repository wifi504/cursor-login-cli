import { createRouter, createWebHashHistory } from 'vue-router'
import routes from '@/router/routes.ts'

// Hash 历史：管理端 SPA 挂在动态安全入口路径下
const router = createRouter({
  history: createWebHashHistory(),
  routes,
})

export default router

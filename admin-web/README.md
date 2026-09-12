# 基于 https://github.com/wifi504/vue-ts-starter
# Vue3 + TS + Vite + @ezview/gen-vue-routes + ESLint + Stylelint + Arco Design

## 脚本

| 脚本 | 说明 |
|------|------|
| `pnpm dev` | 启动 Vite 开发服务 |
| `pnpm gen-routes` | 根据 `src/views/**/Index.vue` 生成 `src/router/routes.ts` |
| `pnpm lint:all` / `lint:all:fix` | ESLint + Stylelint |
| `pnpm build` | 类型检查后构建到本地 `dist/`（再由 VS Code 任务复制到服务端 embed） |

修改视图目录后请先执行 `pnpm gen-routes`，再在 `src/router/options.ts` 填写 name、meta 等增量配置。

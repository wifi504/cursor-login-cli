import path from 'node:path'
import vue from '@vitejs/plugin-vue'
import { defineConfig } from 'vite'

// 开发：页面只由 Vite 提供；/__dev__ 与公开 API 代理到 Go。
// 生产：构建到 dist/，再由任务复制到服务端 embed。
export default defineConfig({
  base: './',
  plugins: [vue()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    port: 5173,
    proxy: {
      '/__dev__': {
        target: 'http://127.0.0.1:8080',
        changeOrigin: true,
      },
      '/api': {
        target: 'http://127.0.0.1:8080',
        changeOrigin: true,
      },
      '/install.sh': {
        target: 'http://127.0.0.1:8080',
        changeOrigin: true,
      },
      '/install.ps1': {
        target: 'http://127.0.0.1:8080',
        changeOrigin: true,
      },
      '/uninstall.sh': {
        target: 'http://127.0.0.1:8080',
        changeOrigin: true,
      },
      '/uninstall.ps1': {
        target: 'http://127.0.0.1:8080',
        changeOrigin: true,
      },
      '/download': {
        target: 'http://127.0.0.1:8080',
        changeOrigin: true,
      },
    },
  },
  build: {
    outDir: 'dist',
    emptyOutDir: true,
  },
})

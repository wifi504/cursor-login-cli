---
name: arco-vue-getting-started
description: "Arco Design Vue 安装与接入指南。用于安装 `@arco-design/web-vue`、注册 ArcoVue、引入 CSS 或 Less 样式、配置 Vite 按需加载、图标和 TypeScript。"
user-invocable: false
---

# 快速开始

来源文档：
- 上游 `packages/web-vue/README.zh-CN.md`
- 上游 `packages/arco-vue-docs/docs/start.zh-CN.md`

## 版本

Arco Design Vue 面向 Vue 3。当前提供的源码包版本是 `@arco-design/web-vue@2.58.0-beta.1`，peer dependency 声明 Vue `>=3.1.0`，官方快速上手文档建议 Vue `>=3.2.0`。

Vue 3 不再支持 IE，Arco Design Vue 也不支持 IE。

## 安装

```bash
npm install @arco-design/web-vue
yarn add @arco-design/web-vue
pnpm add @arco-design/web-vue
```

## 完整引入

```ts
import { createApp } from 'vue';
import ArcoVue from '@arco-design/web-vue';
import '@arco-design/web-vue/dist/arco.css';
import App from './App.vue';

const app = createApp(App);
app.use(ArcoVue);
app.mount('#app');
```

完整注册后，默认使用 `a-` 前缀的全局组件标签。

```vue
<template>
  <a-space>
    <a-button type="primary">提交</a-button>
    <a-input v-model="keyword" placeholder="搜索" />
  </a-space>
</template>
```

## 全局配置

`app.use(ArcoVue, options)` 可以接收全局配置对象。通过 `componentPrefix` 可修改全局组件前缀。

```ts
app.use(ArcoVue, {
  componentPrefix: 'arco',
});
```

设置后，`<a-button>` 会变成 `<arco-button>`。同一个应用中不要混用多个前缀。

## 按需加载

模板开发场景优先使用 `unplugin-vue-components` 和 `ArcoResolver`。

```ts
import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';
import AutoImport from 'unplugin-auto-import/vite';
import Components from 'unplugin-vue-components/vite';
import { ArcoResolver } from 'unplugin-vue-components/resolvers';

export default defineConfig({
  plugins: [
    vue(),
    AutoImport({
      resolvers: [ArcoResolver()],
    }),
    Components({
      resolvers: [
        ArcoResolver({
          sideEffect: true,
        }),
      ],
    }),
  ],
});
```

如果在 `<script>` 中手动导入组件或服务，也需要手动导入对应样式。

```ts
import { Message } from '@arco-design/web-vue';
import '@arco-design/web-vue/es/message/style/css.js';
```

## Arco Vite 插件

也可以使用 `@arco-plugins/vite-vue` 完成按需加载和组件库样式配置。

```ts
import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';
import { vitePluginForArco } from '@arco-plugins/vite-vue';

export default defineConfig({
  plugins: [
    vue(),
    vitePluginForArco({
      style: 'css',
    }),
  ],
});
```

## 组件和图标导入

组件和全局服务从根包导入。

```ts
import { Button, Table, Form, Message, Modal } from '@arco-design/web-vue';
```

图标从图标入口导入。

```ts
import { IconSearch, IconPlus } from '@arco-design/web-vue/es/icon';
```

## TypeScript

组件库使用 TypeScript 编写。业务代码中优先使用有类型的 `ref`、`reactive`、`computed`、`defineProps` 和 `defineEmits`。

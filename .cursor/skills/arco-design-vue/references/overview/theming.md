---
name: arco-vue-theming
description: "Arco Design Vue 主题指南。用于 CSS 引入、Less 变量定制、组件 token、Vite/Webpack `modifyVars`、主题包和暗黑模式。"
user-invocable: false
---

# 主题定制

来源文档：
- 上游 `packages/arco-vue-docs/docs/theme.zh-CN.md`
- 上游 `packages/arco-vue-docs/docs/dark.zh-CN.md`

## 样式引入

完整 CSS 引入最简单：

```ts
import '@arco-design/web-vue/dist/arco.css';
```

需要定制 Less 变量时，引入 Less 样式：

```ts
import '@arco-design/web-vue/dist/arco.less';
```

手动按需引入组件样式：

```ts
import '@arco-design/web-vue/es/button/style/css.js';
```

## Less 变量

全局变量位于 `@arco-design/web-vue/es/style/theme/global.less`，组件级 token 位于类似 `@arco-design/web-vue/es/button/style/token.less` 的路径。

Vite 示例：

```ts
export default {
  css: {
    preprocessorOptions: {
      less: {
        modifyVars: {
          'arcoblue-6': '#f85959',
        },
        javascriptEnabled: true,
      },
    },
  },
};
```

Webpack 示例：

```js
module.exports = {
  module: {
    rules: [
      {
        test: /\.less$/,
        use: [
          'style-loader',
          'css-loader',
          {
            loader: 'less-loader',
            options: {
              lessOptions: {
                modifyVars: {
                  'arcoblue-6': '#f85959',
                },
                javascriptEnabled: true,
              },
            },
          },
        ],
      },
    ],
  },
};
```

## 暗黑模式

Arco Design Vue 通过 `body` 上的 `arco-theme` 属性切换主题。

```ts
document.body.setAttribute('arco-theme', 'dark');
document.body.removeAttribute('arco-theme');
```

把主题状态集中在布局 store 或 `useTheme` composable 中，不要让多个无关组件直接修改 `body` 属性。

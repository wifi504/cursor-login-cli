export default {
  extends: [
    'stylelint-config-standard',
    'stylelint-config-recommended-vue',
  ],
  ignoreFiles: [
    '**/node_modules/**',
    '**/dist/**',
  ],
  rules: {
    // 禁止空块
    'block-no-empty': true,
    // 禁止无效十六进制
    'color-no-invalid-hex': true,
    // 允许直接用 "@import 'xxx'"，不强制 url()
    'import-notation': null,
  },
  overrides: [
    {
      files: ['**/*.(less|css|vue|html)'],
      customSyntax: 'postcss-less',
    },
    {
      files: ['**/*.(html|vue)'],
      customSyntax: 'postcss-html',
    },
  ],
}

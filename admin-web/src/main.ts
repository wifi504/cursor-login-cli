import ArcoVue from '@arco-design/web-vue'
import { createPinia } from 'pinia'
import { createApp } from 'vue'
import router from '@/router'
import App from './App.vue'
import '@arco-design/web-vue/dist/arco.css'
import './style.css'

createApp(App)
  .use(createPinia())
  .use(router)
  .use(ArcoVue)
  .mount('#app')

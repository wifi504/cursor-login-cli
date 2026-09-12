<template>
  <div class="page">
    <a-card title="初始化 Cursor Login" class="card" :loading="loading">
      <a-alert type="warning" style="margin-bottom: 16px">
        首次启动请在此创建管理员、自定义安全入口与对外访问域名。完成后临时入口将失效。
      </a-alert>
      <a-form :model="form" layout="vertical" @submit="onSubmit">
        <a-form-item field="username" label="管理员用户名" required>
          <a-input v-model="form.username" allow-clear />
        </a-form-item>
        <a-form-item field="password" label="管理员密码" required>
          <a-input-password v-model="form.password" allow-clear />
        </a-form-item>
        <a-form-item field="admin_entry" label="Cursor Login Admin 安全入口（路径）" required>
          <a-input v-model="form.admin_entry" placeholder="/console" allow-clear />
        </a-form-item>
        <a-form-item field="public_base_url" label="对外访问地址（仅域名，勿带 path）" required>
          <a-input
            v-model="form.public_base_url"
            placeholder="https://cursor-login.example.com"
            allow-clear
          />
        </a-form-item>
        <a-button type="primary" html-type="submit" long :loading="submitting">
          完成初始化
        </a-button>
      </a-form>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import type { StatusResp } from '@/api/admin'
import { Message } from '@arco-design/web-vue'
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { api, rememberEntry } from '@/api/admin'

const router = useRouter()
const loading = ref(true)
const submitting = ref(false)
const form = reactive({
  username: 'admin',
  password: '',
  admin_entry: '/console',
  // 默认取当前浏览器访问源，管理员可再改成正式域名
  public_base_url: typeof window !== 'undefined' ? window.location.origin : '',
})

onMounted(async () => {
  try {
    const st = await api<StatusResp>('/status')
    rememberEntry(st.admin_entry)
    if (st.initialized) {
      await router.replace({ name: 'login' })
      return
    }
    if (st.public_base_url) {
      form.public_base_url = st.public_base_url
    } else {
      form.public_base_url = window.location.origin
    }
  } catch (e) {
    Message.error(e instanceof Error ? e.message : '无法读取状态')
    form.public_base_url = window.location.origin
  } finally {
    loading.value = false
  }
})

async function onSubmit() {
  submitting.value = true
  try {
    const res = await api<{ ok: boolean, admin_entry: string }>('/bootstrap', {
      method: 'POST',
      body: JSON.stringify(form),
    })
    rememberEntry(res.admin_entry)
    Message.success('初始化成功')
    if (import.meta.env.DEV) {
      await router.replace({ name: 'login' })
      return
    }
    const base = res.admin_entry.replace(/\/+$/, '')
    window.location.href = `${base}/#/login`
  } catch (e) {
    Message.error(e instanceof Error ? e.message : '初始化失败')
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped lang="less">
.page {
  display: flex;
  align-items: center;
  justify-content: center;
  box-sizing: border-box;
  min-height: 100%;
  padding: 24px 16px;
}

.card {
  width: 100%;
  max-width: 520px;
}
</style>

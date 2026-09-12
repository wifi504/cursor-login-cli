<template>
  <div class="page">
    <a-card title="登录 Cursor Login Admin" class="card" :loading="loading">
      <a-form :model="form" layout="vertical" @submit="onSubmit">
        <a-form-item field="username" label="用户名" required>
          <a-input v-model="form.username" allow-clear />
        </a-form-item>
        <a-form-item field="password" label="密码" required>
          <a-input-password v-model="form.password" allow-clear />
        </a-form-item>
        <a-button type="primary" html-type="submit" long :loading="submitting">
          登录
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
})

onMounted(async () => {
  try {
    const st = await api<StatusResp>('/status')
    rememberEntry(st.admin_entry)
    if (!st.initialized) {
      await router.replace({ name: 'setup' })
      return
    }
    try {
      await api<{ username: string }>('/me')
      await router.replace({ name: 'home' })
    } catch {
      // 尚未登录，停留在登录页
    }
  } catch (e) {
    Message.error(e instanceof Error ? e.message : '无法读取状态')
  } finally {
    loading.value = false
  }
})

async function onSubmit() {
  submitting.value = true
  try {
    await api('/login', {
      method: 'POST',
      body: JSON.stringify(form),
    })
    Message.success('登录成功')
    await router.push({ name: 'home' })
  } catch (e) {
    Message.error(e instanceof Error ? e.message : '登录失败')
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
  max-width: 420px;
}
</style>

<template>
  <div class="settings">
    <a-card title="安全入口" :loading="loading">
      <a-alert type="warning" style="margin-bottom: 16px">
        修改后旧入口立即失效。
        <template v-if="isDev">开发模式（Vite）仍走 /__dev__，正式访问请用新入口。</template>
      </a-alert>
      <a-form :model="entryForm" layout="vertical" @submit="saveEntry">
        <a-form-item field="admin_entry" label="Cursor Login Admin 安全入口（路径）" required>
          <a-input v-model="entryForm.admin_entry" placeholder="/console" allow-clear />
        </a-form-item>
        <a-form-item field="current_password" label="当前密码（确认）" required>
          <a-input-password v-model="entryForm.current_password" allow-clear />
        </a-form-item>
        <a-button type="primary" html-type="submit" :loading="savingEntry">
          保存入口
        </a-button>
      </a-form>
    </a-card>

    <a-card title="管理员账号" :loading="loading">
      <a-alert type="info" style="margin-bottom: 16px">
        修改用户名或密码后需要重新登录。新密码留空表示不修改密码。
      </a-alert>
      <a-form :model="accountForm" layout="vertical" @submit="saveAccount">
        <a-form-item field="username" label="用户名" required>
          <a-input v-model="accountForm.username" allow-clear />
        </a-form-item>
        <a-form-item field="password" label="新密码">
          <a-input-password v-model="accountForm.password" placeholder="留空则不修改" allow-clear />
        </a-form-item>
        <a-form-item field="current_password" label="当前密码（确认）" required>
          <a-input-password v-model="accountForm.current_password" allow-clear />
        </a-form-item>
        <a-button type="primary" html-type="submit" :loading="savingAccount">
          保存账号
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
const savingEntry = ref(false)
const savingAccount = ref(false)
const isDev = import.meta.env.DEV

const entryForm = reactive({
  admin_entry: '',
  current_password: '',
})

const accountForm = reactive({
  username: '',
  password: '',
  current_password: '',
})

onMounted(async () => {
  try {
    const st = await api<StatusResp>('/status')
    rememberEntry(st.admin_entry)
    entryForm.admin_entry = st.admin_entry
    const me = await api<{ username: string }>('/me')
    accountForm.username = me.username
  } catch {
    Message.warning('请先登录')
    await router.replace({ name: 'login' })
  } finally {
    loading.value = false
  }
})

async function saveEntry() {
  savingEntry.value = true
  try {
    const res = await api<{ ok: boolean, admin_entry: string }>('/settings/entry', {
      method: 'PUT',
      body: JSON.stringify(entryForm),
    })
    rememberEntry(res.admin_entry)
    entryForm.current_password = ''
    Message.success('安全入口已更新')
    if (!import.meta.env.DEV) {
      const base = res.admin_entry.replace(/\/+$/, '')
      window.location.href = `${base}/#/settings`
    }
  } catch (e) {
    Message.error(e instanceof Error ? e.message : '保存失败')
  } finally {
    savingEntry.value = false
  }
}

async function saveAccount() {
  savingAccount.value = true
  try {
    const res = await api<{ ok: boolean, relogin?: boolean }>('/settings/account', {
      method: 'PUT',
      body: JSON.stringify(accountForm),
    })
    accountForm.password = ''
    accountForm.current_password = ''
    if (res.relogin) {
      Message.success('账号已更新，请重新登录')
      await router.replace({ name: 'login' })
      return
    }
    Message.success('已保存')
    const me = await api<{ username: string }>('/me')
    accountForm.username = me.username
  } catch (e) {
    Message.error(e instanceof Error ? e.message : '保存失败')
  } finally {
    savingAccount.value = false
  }
}
</script>

<style scoped lang="less">
.settings {
  display: grid;
  gap: 16px;
}
</style>

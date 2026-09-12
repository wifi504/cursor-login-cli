<template>
  <div class="boot">
    <a-spin tip="加载中..." />
  </div>
</template>

<script setup lang="ts">
import type { StatusResp } from '@/api/admin'
import { Message } from '@arco-design/web-vue'
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { api, rememberEntry } from '@/api/admin'

const router = useRouter()

onMounted(async () => {
  try {
    const st = await api<StatusResp>('/status')
    rememberEntry(st.admin_entry)
    if (st.initialized) {
      await router.replace({ name: 'login' })
      return
    }
    await router.replace({ name: 'setup' })
  } catch (e) {
    Message.error(e instanceof Error ? e.message : '无法读取状态')
  }
})
</script>

<style scoped lang="less">
.boot {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
}
</style>

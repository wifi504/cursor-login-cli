<template>
  <a-card title="新建账号">
    <account-form mode="create" :submitting="submitting" @submit="onCreate" @cancel="goBack" />
  </a-card>
</template>

<script setup lang="ts">
import type { AccountInput } from '@/api/accounts'
import { Message } from '@arco-design/web-vue'
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { createAccount } from '@/api/accounts'
import AccountForm from '@/components/AccountForm.vue'

const router = useRouter()
const submitting = ref(false)

function goBack() {
  router.push({ name: 'accounts' })
}

async function onCreate(body: AccountInput) {
  submitting.value = true
  try {
    await createAccount(body)
    Message.success('创建成功')
    goBack()
  } catch (e) {
    Message.error(e instanceof Error ? e.message : '创建失败')
  } finally {
    submitting.value = false
  }
}
</script>

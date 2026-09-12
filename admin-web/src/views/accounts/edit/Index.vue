<template>
  <a-card title="编辑账号">
    <template v-if="account" #extra>
      <a-button status="danger" @click="onDelete">
        删除
      </a-button>
    </template>
    <a-spin :loading="loading" style="width: 100%">
      <account-form
        v-if="account"
        mode="edit"
        :account="account"
        :submitting="submitting"
        @submit="onSave"
        @cancel="goBack"
        @refreshed="onRefreshed"
      />
      <a-empty v-else-if="!loading" description="账号不存在" />
    </a-spin>
  </a-card>
</template>

<script setup lang="ts">
import type { Account, AccountInput } from '@/api/accounts'
import { Message, Modal } from '@arco-design/web-vue'
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { deleteAccount, getAccount, updateAccount } from '@/api/accounts'
import AccountForm from '@/components/AccountForm.vue'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const submitting = ref(false)
const account = ref<Account | null>(null)

function goBack() {
  router.push({ name: 'accounts' })
}

function onRefreshed(a: Account) {
  account.value = a
}

async function load() {
  const id = Number(route.query.id)
  if (!Number.isFinite(id) || id <= 0) {
    Message.error('缺少账号 ID')
    goBack()
    return
  }
  loading.value = true
  try {
    account.value = await getAccount(id)
  } catch (e) {
    Message.error(e instanceof Error ? e.message : '加载失败')
    account.value = null
  } finally {
    loading.value = false
  }
}

async function onSave(body: AccountInput) {
  if (!account.value) { return }
  submitting.value = true
  try {
    account.value = await updateAccount(account.value.id, body)
    Message.success('已保存')
  } catch (e) {
    Message.error(e instanceof Error ? e.message : '保存失败')
  } finally {
    submitting.value = false
  }
}

function onDelete() {
  if (!account.value) { return }
  const name = account.value.name
  const id = account.value.id
  Modal.confirm({
    title: '删除账号',
    content: `确认删除「${name}」？此操作不可恢复。`,
    onOk: async () => {
      await deleteAccount(id)
      Message.success('已删除')
      goBack()
    },
  })
}

onMounted(load)
</script>

<template>
  <a-form :model="form" layout="vertical" @submit="onSubmit">
    <a-form-item field="name" label="展示名 / 备注" required>
      <a-input v-model="form.name" allow-clear placeholder="CLI 展示用名称" />
    </a-form-item>
    <a-form-item field="email" label="邮箱">
      <a-input v-model="form.email" allow-clear placeholder="账号邮箱" />
    </a-form-item>
    <a-form-item field="access_token" label="AccessToken">
      <a-space direction="vertical" fill style="width: 100%">
        <a-input-password v-model="form.access_token" allow-clear />
        <a-space wrap>
          <a-button type="outline" size="small" @click="copyText(form.access_token)">
            复制
          </a-button>
          <a-button type="outline" size="small" @click="tokenModalVisible = true">
            获取 AccessToken
          </a-button>
        </a-space>
      </a-space>
    </a-form-item>
    <a-form-item field="claim_code" label="上号码" required>
      <a-space direction="vertical" fill style="width: 100%">
        <a-input-password v-model="form.claim_code" allow-clear />
        <a-space wrap>
          <a-button type="outline" size="small" @click="copyText(form.claim_code)">
            复制
          </a-button>
          <a-button type="outline" size="small" @click="form.claim_code = randomClaimCode()">
            随机生成
          </a-button>
        </a-space>
      </a-space>
    </a-form-item>
    <a-form-item v-if="mode === 'create'" field="claim_limit" label="允许领取次数" required>
      <a-input-number v-model="form.claim_limit" :min="0" :precision="0" style="width: 160px" />
    </a-form-item>

    <template v-if="mode === 'edit' && account">
      <a-divider orientation="left">
        领取配额
      </a-divider>
      <a-descriptions :column="1" bordered size="small" style="margin-bottom: 16px">
        <a-descriptions-item label="已领取">
          {{ account.claimed_count }}
        </a-descriptions-item>
        <a-descriptions-item label="剩余">
          {{ Math.max(0, limitDraft - account.claimed_count) }}
        </a-descriptions-item>
      </a-descriptions>
      <a-form-item label="允许领取次数（总配额）">
        <a-space>
          <a-input-number v-model="limitDraft" :min="0" :precision="0" style="width: 160px" />
          <a-button type="outline" :loading="savingLimit" @click="saveLimit">
            保存配额
          </a-button>
          <a-button
            status="warning"
            :loading="resetting"
            :disabled="account.claimed_count === 0"
            @click="doResetClaimed"
          >
            重置已领取
          </a-button>
        </a-space>
      </a-form-item>
    </template>

    <a-space style="margin-top: 8px">
      <a-button type="primary" html-type="submit" :loading="submitting">
        {{ mode === 'create' ? '创建' : '保存' }}
      </a-button>
      <a-button @click="emit('cancel')">
        返回
      </a-button>
    </a-space>
  </a-form>

  <fetch-access-token-modal v-model:visible="tokenModalVisible" @apply="onTokenApply" />
</template>

<script setup lang="ts">
import type { Account, AccountInput } from '@/api/accounts'
import { Message, Modal } from '@arco-design/web-vue'
import { reactive, ref, watch } from 'vue'
import { resetClaimed, updateClaimLimit } from '@/api/accounts'
import FetchAccessTokenModal from '@/components/FetchAccessTokenModal.vue'
import { randomClaimCode } from '@/utils/crypto'

const props = defineProps<{
  mode: 'create' | 'edit'
  account?: Account | null
  submitting?: boolean
}>()

const emit = defineEmits<{
  submit: [body: AccountInput]
  cancel: []
  refreshed: [account: Account]
}>()

const form = reactive({
  name: '',
  email: '',
  access_token: '',
  claim_code: '',
  claim_limit: 1,
})

const limitDraft = ref(1)
const savingLimit = ref(false)
const resetting = ref(false)
const tokenModalVisible = ref(false)

function fillFromAccount(a: Account) {
  form.name = a.name
  form.email = a.email
  form.access_token = a.access_token
  form.claim_code = a.claim_code
  form.claim_limit = a.claim_limit
  limitDraft.value = a.claim_limit
}

watch(
  () => props.account,
  (a) => {
    if (a) { fillFromAccount(a) }
  },
  { immediate: true },
)

watch(
  () => props.mode,
  (m) => {
    if (m === 'create' && !form.claim_code) {
      form.claim_code = randomClaimCode()
      form.claim_limit = 1
    }
  },
  { immediate: true },
)

async function copyText(text: string) {
  if (!text) {
    Message.warning('内容为空')
    return
  }
  await navigator.clipboard.writeText(text)
  Message.success('已复制')
}

function onTokenApply(token: string) {
  form.access_token = token
  Message.success('已应用到表单，请记得保存账号')
}

function onSubmit() {
  const body: AccountInput = {
    name: form.name,
    email: form.email,
    access_token: form.access_token,
    claim_code: form.claim_code,
  }
  if (props.mode === 'create') {
    body.claim_limit = form.claim_limit
  }
  emit('submit', body)
}

async function saveLimit() {
  if (!props.account) { return }
  savingLimit.value = true
  try {
    const a = await updateClaimLimit(props.account.id, props.account.claim_limit, limitDraft.value)
    Message.success('配额已更新')
    emit('refreshed', a)
  } catch (e) {
    Message.error(e instanceof Error ? e.message : '更新失败')
  } finally {
    savingLimit.value = false
  }
}

function doResetClaimed() {
  if (!props.account) { return }
  Modal.confirm({
    title: '重置已领取次数',
    content: `将已领取次数从 ${props.account.claimed_count} 重置为 0？`,
    onOk: async () => {
      resetting.value = true
      try {
        const a = await resetClaimed(props.account!.id, props.account!.claimed_count)
        Message.success('已重置')
        emit('refreshed', a)
      } catch (e) {
        Message.error(e instanceof Error ? e.message : '重置失败')
      } finally {
        resetting.value = false
      }
    },
  })
}
</script>

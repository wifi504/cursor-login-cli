<template>
  <a-modal
    :visible="visible"
    title="获取 AccessToken"
    :mask-closable="false"
    :esc-to-close="false"
    width="520px"
    unmount-on-close
    @cancel="onCancel"
  >
    <a-space direction="vertical" fill style="width: 100%">
      <a-alert type="info">
        粘贴浏览器中的 WorkosCursorSessionToken，开始后会打开 Cursor 确认窗口；请在弹窗中点击 Yes, Log In，本页将自动轮询结果。
      </a-alert>
      <a-form :model="{ sessionToken }" layout="vertical">
        <a-form-item label="WorkosCursorSessionToken" required>
          <a-input-password
            v-model="sessionToken"
            allow-clear
            placeholder="WorkosCursorSessionToken"
          />
        </a-form-item>
      </a-form>
      <a-space>
        <a-button type="primary" :loading="starting" :disabled="polling" @click="startAuth">
          开始认证
        </a-button>
        <a-tag v-if="status" :color="statusColor">
          {{ status }}
        </a-tag>
      </a-space>
      <div v-if="resultToken" class="result">
        <div class="result-label">
          已获取 AccessToken
        </div>
        <a-textarea :model-value="resultToken" :auto-size="{ minRows: 3, maxRows: 6 }" readonly />
        <div v-if="resultUserId" class="muted">
          userId: {{ resultUserId }}
        </div>
      </div>
    </a-space>
    <template #footer>
      <a-button @click="onCancel">
        取消
      </a-button>
      <a-button type="primary" :disabled="!resultToken" @click="onApply">
        应用
      </a-button>
    </template>
  </a-modal>
</template>

<script setup lang="ts">
import { Message } from '@arco-design/web-vue'
import { computed, ref, watch } from 'vue'
import { pollCursorAuth } from '@/api/accounts'
import { generatePKCEPair, generateUUID } from '@/utils/crypto'

const props = defineProps<{
  visible: boolean
}>()

const emit = defineEmits<{
  'update:visible': [boolean]
  'apply': [token: string]
}>()

const sessionToken = ref('')
const starting = ref(false)
const polling = ref(false)
const status = ref('')
const statusType = ref<'info' | 'success' | 'error'>('info')
const resultToken = ref('')
const resultUserId = ref('')
let pollAbort = false

const statusColor = computed(() => {
  if (statusType.value === 'success') { return 'green' }
  if (statusType.value === 'error') { return 'red' }
  return 'arcoblue'
})

watch(() => props.visible, (v) => {
  if (v) {
    pollAbort = false
    starting.value = false
    polling.value = false
    status.value = ''
    resultToken.value = ''
    resultUserId.value = ''
    const saved = localStorage.getItem('cursor_session_token')
    if (saved) { sessionToken.value = saved }
  } else {
    pollAbort = true
  }
})

function onCancel() {
  pollAbort = true
  emit('update:visible', false)
}

function onApply() {
  if (!resultToken.value) { return }
  emit('apply', resultToken.value)
  emit('update:visible', false)
}

async function startAuth() {
  const token = sessionToken.value.trim()
  if (!token) {
    Message.warning('请先输入 Session Token')
    return
  }
  localStorage.setItem('cursor_session_token', token)
  starting.value = true
  statusType.value = 'info'
  status.value = '正在初始化…'
  resultToken.value = ''
  resultUserId.value = ''
  try {
    const { codeVerifier, codeChallenge } = await generatePKCEPair()
    const uuid = generateUUID()
    const clientLoginUrl = `https://www.cursor.com/cn/loginDeepControl?challenge=${encodeURIComponent(codeChallenge)}&uuid=${encodeURIComponent(uuid)}&mode=login`
    const width = 500
    const height = 600
    const left = Math.round((window.screen.width - width) / 2)
    const top = Math.round((window.screen.height - height) / 2)
    window.open(
      clientLoginUrl,
      'CursorAuth',
      `width=${width},height=${height},left=${left},top=${top},toolbar=no,menubar=no,scrollbars=yes,resizable=yes`,
    )
    status.value = '请在弹窗中确认登录，正在轮询…'
    starting.value = false
    await runPoll(uuid, codeVerifier)
  } catch (e) {
    statusType.value = 'error'
    status.value = e instanceof Error ? e.message : '初始化失败'
  } finally {
    starting.value = false
  }
}

async function runPoll(uuid: string, verifier: string, maxAttempts = 40) {
  polling.value = true
  pollAbort = false
  try {
    for (let i = 0; i < maxAttempts; i++) {
      if (pollAbort) { return }
      try {
        const data = await pollCursorAuth(uuid, verifier)
        const accessToken = typeof data.accessToken === 'string' ? data.accessToken : ''
        const authId = typeof data.authId === 'string' ? data.authId : ''
        if (accessToken) {
          resultToken.value = accessToken
          resultUserId.value = authId.includes('|') ? authId.split('|')[1] || '' : ''
          statusType.value = 'success'
          status.value = '获取成功'
          return
        }
      } catch {
        // 继续重试
      }
      await new Promise(r => setTimeout(r, 2000))
    }
    statusType.value = 'error'
    status.value = '轮询超时，请确认已在弹窗中登录后重试'
  } finally {
    polling.value = false
  }
}
</script>

<style scoped lang="less">
.result {
  display: grid;
  gap: 8px;
}

.result-label {
  font-weight: 600;
}

.muted {
  color: var(--color-text-3);
  font-size: 12px;
}
</style>

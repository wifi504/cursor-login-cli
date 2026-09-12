<template>
  <div v-if="!loading" class="home">
    <a-card title="服务信息">
      <a-descriptions :column="1" bordered size="large">
        <a-descriptions-item label="对外地址">
          {{ publicBase || '-' }}
        </a-descriptions-item>
        <a-descriptions-item label="Cursor Login Admin 安全入口">
          {{ adminEntry || '-' }}
        </a-descriptions-item>
      </a-descriptions>
    </a-card>

    <a-card title="一键安装 / 卸载命令">
      <a-space direction="vertical" fill :size="16" style="width: 100%">
        <a-card v-for="plat in platforms" :key="plat.key" size="small" :bordered="true">
          <template #title>
            <a-space :size="8" align="center">
              <img :src="plat.icon" alt="" class="plat-icon" :class="`plat-icon--${plat.key}`">
              <span>{{ plat.title }}</span>
            </a-space>
          </template>
          <a-list :bordered="false" :split="true">
            <a-list-item v-for="item in plat.items" :key="item.label">
              <a-list-item-meta :title="item.label" />
              <a-space fill :size="12" align="center" style="width: 100%">
                <a-input :model-value="item.cmd" readonly />
                <a-button type="outline" @click="copyText(item.cmd)">
                  复制
                </a-button>
              </a-space>
            </a-list-item>
          </a-list>
        </a-card>
      </a-space>
    </a-card>
  </div>
  <div v-else class="page-loading">
    <a-spin tip="加载中..." />
  </div>
</template>

<script setup lang="ts">
import type { StatusResp } from '@/api/admin'
import { Message } from '@arco-design/web-vue'
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { api, rememberEntry } from '@/api/admin'
import unixIcon from '@/assets/icons/unix.svg'
import windowsIcon from '@/assets/icons/windows.svg'

const router = useRouter()
const loading = ref(true)
const publicBase = ref('')
const adminEntry = ref('')
const commands = ref<Record<string, string>>({})

const platforms = computed(() => [
  {
    key: 'windows',
    title: 'Windows',
    icon: windowsIcon,
    items: [
      { label: '安装（PowerShell）', cmd: commands.value.install_windows || '' },
      { label: '卸载（PowerShell）', cmd: commands.value.uninstall_windows || '' },
    ],
  },
  {
    key: 'unix',
    title: 'macOS / Linux',
    icon: unixIcon,
    items: [
      { label: '安装（Shell）', cmd: commands.value.install_unix || '' },
      { label: '卸载（Shell）', cmd: commands.value.uninstall_unix || '' },
    ],
  },
])

onMounted(async () => {
  try {
    const st = await api<StatusResp>('/status')
    rememberEntry(st.admin_entry)
    publicBase.value = st.public_base_url
    adminEntry.value = st.admin_entry
    commands.value = await api<Record<string, string>>('/install-commands')
  } catch {
    Message.warning('请先登录')
    await router.replace({ name: 'login' })
  } finally {
    loading.value = false
  }
})

async function copyText(text: string) {
  await navigator.clipboard.writeText(text)
  Message.success('已复制')
}
</script>

<style scoped lang="less">
.home {
  display: grid;
  gap: 16px;
}

.page-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: calc(100vh - 56px - 40px);
}

.plat-icon {
  display: block;
  width: 18px;
  height: 18px;
  opacity: 0.85;
}

:deep(.arco-list-item-content .arco-space) {
  width: 100%;
}

:deep(.arco-list-item-content .arco-space-item:first-child) {
  flex: 1;
  min-width: 0;
}
</style>

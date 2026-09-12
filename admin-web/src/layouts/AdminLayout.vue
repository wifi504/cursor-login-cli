<template>
  <div class="admin-shell">
    <a-layout class="admin-layout">
      <a-layout-sider :width="200" :collapsible="false">
        <div class="brand">
          <span class="brand-text">Cursor Login Admin</span>
        </div>
        <a-menu
          :selected-keys="selectedKeys"
          :style="{ width: '100%' }"
          @menu-item-click="onMenuClick"
        >
          <a-menu-item key="home">
            <template #icon>
              <icon-home />
            </template>
            概览
          </a-menu-item>
          <a-menu-item key="accounts">
            <template #icon>
              <icon-user />
            </template>
            账号管理
          </a-menu-item>
          <a-menu-item key="settings">
            <template #icon>
              <icon-settings />
            </template>
            系统设置
          </a-menu-item>
        </a-menu>
      </a-layout-sider>
      <a-layout>
        <a-layout-header class="header">
          <div class="header-title">
            {{ pageTitle }}
          </div>
          <div class="header-right">
            <span class="user">{{ username }}</span>
            <a-button type="text" @click="logout">
              退出
            </a-button>
          </div>
        </a-layout-header>
        <a-layout-content class="content">
          <router-view />
        </a-layout-content>
      </a-layout>
    </a-layout>
  </div>
</template>

<script setup lang="ts">
import { Message } from '@arco-design/web-vue'
import { IconHome, IconSettings, IconUser } from '@arco-design/web-vue/es/icon'
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api } from '@/api/admin'

const route = useRoute()
const router = useRouter()
const username = ref('')

const selectedKeys = computed(() => {
  const name = String(route.name || '')
  if (name === 'accounts' || name === 'accounts-new' || name === 'accounts-edit') {
    return ['accounts']
  }
  return name ? [name] : []
})

const pageTitle = computed(() => String(route.meta.title || 'Cursor Login Admin'))

onMounted(async () => {
  try {
    const me = await api<{ username: string }>('/me')
    username.value = me.username
  } catch {
    Message.warning('请先登录')
    await router.replace({ name: 'login' })
  }
})

watch(
  () => route.name,
  async () => {
    try {
      const me = await api<{ username: string }>('/me')
      username.value = me.username
    } catch {
      // 设置页改密后会跳登录，这里忽略
    }
  },
)

function onMenuClick(key: string) {
  if (route.name !== key) {
    void router.push({ name: key })
  }
}

async function logout() {
  await api('/logout', { method: 'POST' })
  await router.replace({ name: 'login' })
}
</script>

<style scoped lang="less">
.admin-shell {
  display: flex;
  justify-content: center;
  box-sizing: border-box;
  min-height: 100%;
  padding: 0;
  background: #fff;
}

.admin-layout {
  width: 100%;
  max-width: none;
  min-height: 100vh;
  overflow: hidden;
  background: var(--color-bg-1);
  border-right: none;
  border-left: none;
}

@media (width >= 1100px) {
  .admin-shell {
    padding: 0 16px;
  }

  .admin-layout {
    max-width: 1080px;
    border-right: 1px solid var(--color-border);
    border-left: 1px solid var(--color-border);
  }
}

.brand {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 56px;
  padding: 0 8px;
  border-bottom: 1px solid var(--color-border);
}

.brand-text {
  overflow: hidden;
  font-weight: 700;
  font-size: 16px;
  letter-spacing: 0;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 56px;
  padding: 0 20px;
  background: var(--color-bg-2);
  border-bottom: 1px solid var(--color-border);
}

.header-title {
  font-weight: 600;
  font-size: 16px;
}

.header-right {
  display: flex;
  gap: 8px;
  align-items: center;
}

.user {
  color: var(--color-text-2);
  font-size: 13px;
}

.content {
  box-sizing: border-box;
  min-height: calc(100vh - 56px);
  padding: 20px;
  background: var(--color-fill-2);
}
</style>

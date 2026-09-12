<template>
  <router-view v-if="isChildRoute" />
  <a-card v-else title="账号管理">
    <template #extra>
      <a-button type="primary" @click="router.push({ name: 'accounts-new' })">
        新建账号
      </a-button>
    </template>
    <a-table
      row-key="id"
      :loading="loading"
      :data="items"
      :pagination="false"
      :bordered="{ cell: true }"
      :scroll="{ x: tableScrollX }"
      :table-layout-fixed="true"
      :hoverable="true"
      :row-class="() => 'account-row'"
      @row-click="onRowClick"
    >
      <template #columns>
        <a-table-column title="ID" data-index="id" :width="72" />
        <a-table-column title="展示名" data-index="name" :width="160" :ellipsis="true" tooltip />
        <a-table-column title="邮箱" data-index="email" :width="220" :ellipsis="true" tooltip />
        <a-table-column title="上号码" data-index="claim_code" :width="200" :ellipsis="true" tooltip />
        <a-table-column title="配额" :width="100">
          <template #cell="{ record }">
            {{ record.claimed_count }} / {{ record.claim_limit }}
          </template>
        </a-table-column>
        <a-table-column title="更新时间" :width="180">
          <template #cell="{ record }">
            {{ formatShanghai(record.updated_at) }}
          </template>
        </a-table-column>
      </template>
    </a-table>
  </a-card>
</template>

<script setup lang="ts">
import type { TableData } from '@arco-design/web-vue'
import type { Account } from '@/api/accounts'
import { Message } from '@arco-design/web-vue'
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { listAccounts } from '@/api/accounts'
import { formatShanghai } from '@/utils/datetime'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const items = ref<Account[]>([])

/** 各列 width 之和，驱动横向滚动 */
const tableScrollX = 72 + 160 + 220 + 200 + 100 + 180

const isChildRoute = computed(() => route.name === 'accounts-new' || route.name === 'accounts-edit')

async function load() {
  loading.value = true
  try {
    const res = await listAccounts()
    items.value = res.items || []
  } catch (e) {
    Message.error(e instanceof Error ? e.message : '加载失败')
  } finally {
    loading.value = false
  }
}

function onRowClick(record: TableData) {
  router.push({ name: 'accounts-edit', query: { id: String(record.id) } })
}

watch(isChildRoute, (child) => {
  if (!child) { void load() }
})

onMounted(() => {
  if (!isChildRoute.value) { void load() }
})
</script>

<style scoped lang="less">
:deep(.account-row) {
  cursor: pointer;
}
</style>

<template>
  <admin-layout v-if="useAdminShell" />
  <router-view v-else />
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { RouterView, useRoute } from 'vue-router'
import AdminLayout from '@/layouts/AdminLayout.vue'

const route = useRoute()
const useAdminShell = computed(() => {
  const name = String(route.name || '')
  return name === 'home' || name === 'settings'
    || name === 'accounts' || name === 'accounts-new' || name === 'accounts-edit'
})
</script>

<script setup lang="ts">
/**
 * Example Feature Component
 * Place domain-specific UI components in features/<domain>/components.
 */
import Avatar from 'primevue/avatar'

import type { User } from '@/types'

defineProps<{
  user: User
}>()
</script>

<template>
  <div class="flex items-center gap-3">
    <Avatar 
      :image="user.avatarUrl" 
      :label="!user.avatarUrl ? user.name.charAt(0) : undefined" 
      shape="circle" 
      class="bg-sky-100 dark:bg-sky-950 text-sky-700 dark:text-sky-300"
    />
    <div>
      <div class="font-medium text-sm text-slate-900 dark:text-slate-50">{{ user.name }}</div>
      <div class="text-xs text-slate-500">{{ user.email }}</div>
    </div>
  </div>
</template>

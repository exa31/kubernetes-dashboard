<script setup lang="ts">
import ConfirmDialog from 'primevue/confirmdialog'
import Toast from 'primevue/toast'

import { useDarkMode } from '@/composables'

// Initialize dark mode reactive listener
useDarkMode()
</script>

<template>
  <Toast />
  <ConfirmDialog />
  <router-view />
</template>

<style scoped></style>

<script setup lang="ts">
import { useDarkMode } from '@/composables'
import Toast from 'primevue/toast'

// Initialize dark mode reactive listener
useDarkMode()
</script>

<template>
  <Toast />
  <router-view />
</template>

<style scoped>
</style>

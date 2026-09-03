<script setup lang="ts">
import { storeToRefs } from 'pinia'
import Button from 'primevue/button'
import IconField from 'primevue/iconfield'
import InputIcon from 'primevue/inputicon'
import InputText from 'primevue/inputtext'
import { ref } from 'vue'
import { useRouter } from 'vue-router'

import { useAuthStore } from '@/stores'

const router = useRouter()
const authStore = useAuthStore()
const { isLoading } = storeToRefs(authStore)

const email = ref('')
const password = ref('')
const errorMessage = ref('')

const handleLogin = async () => {
  errorMessage.value = ''
  try {
    await authStore.login({
      email: email.value.trim(),
      password: password.value,
    })
    router.push('/')
  } catch (err: unknown) {
    if (err && typeof err === 'object' && 'response' in err) {
      const res = (err as { response?: { data?: { message?: string } } }).response
      errorMessage.value = res?.data?.message || 'Login failed. Check credentials.'
    } else if (err instanceof Error) {
      errorMessage.value = err.message
    } else {
      errorMessage.value = 'Login failed. Check credentials.'
    }
  }
}
</script>

<template>
  <div class="p-8">
    <div class="mb-6">
      <h2 class="text-xl font-bold text-slate-900 dark:text-slate-100">Sign in to KubeNexus</h2>
      <p class="text-xs text-slate-500 mt-1">Enterprise Kubernetes Orchestration & Observability</p>
    </div>

    <!-- Error message banner -->
    <div
      v-if="errorMessage"
      class="mb-4 p-3 rounded-lg bg-rose-50 dark:bg-rose-950/40 border border-rose-200 dark:border-rose-800 text-rose-700 dark:text-rose-300 text-xs flex items-center gap-2"
    >
      <i class="pi pi-exclamation-circle"></i>
      <span>{{ errorMessage }}</span>
    </div>

    <form class="space-y-4" @submit.prevent="handleLogin">
      <div>
        <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1.5">
          Email
        </label>
        <IconField class="w-full">
          <InputIcon class="pi pi-envelope text-slate-400" />
          <InputText
            v-model="email"
            type="email"
            class="w-full text-sm"
            placeholder="name@company.com"
            required
            autocomplete="email"
          />
        </IconField>
      </div>

      <div>
        <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1.5">
          Password
        </label>
        <IconField class="w-full">
          <InputIcon class="pi pi-lock text-slate-400" />
          <InputText
            v-model="password"
            type="password"
            class="w-full text-sm"
            placeholder="••••••••"
            required
            autocomplete="current-password"
          />
        </IconField>
      </div>

      <Button
        type="submit"
        label="Sign In"
        :loading="isLoading"
        class="w-full bg-emerald-600 hover:bg-emerald-500 text-white font-semibold border-none mt-2 shadow-xs cursor-pointer"
      />
    </form>
  </div>
</template>

<style scoped>
</style>

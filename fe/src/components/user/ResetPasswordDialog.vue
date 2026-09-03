<script setup lang="ts">
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import { useToast } from 'primevue/usetoast'
import { ref, watch } from 'vue'

import { useUserStore } from '@/stores'
import type { User } from '@/types'

const props = defineProps<{
  visible: boolean
  user: User | null
}>()

const emit = defineEmits<{
  (e: 'update:visible', val: boolean): void
  (e: 'success'): void
}>()

const userStore = useUserStore()
const toast = useToast()

const newPassword = ref('')
const confirmPassword = ref('')
const errorMessage = ref('')

watch(
  () => props.visible,
  (isOpen) => {
    if (isOpen) {
      newPassword.value = ''
      confirmPassword.value = ''
      errorMessage.value = ''
    }
  },
)

const handleClose = () => {
  emit('update:visible', false)
}

const handleSubmit = async () => {
  errorMessage.value = ''
  if (!newPassword.value || newPassword.value.length < 6) {
    errorMessage.value = 'Password must be at least 6 characters long'
    return
  }
  if (newPassword.value !== confirmPassword.value) {
    errorMessage.value = 'Passwords do not match'
    return
  }
  if (!props.user) return

  try {
    await userStore.resetPassword(props.user.id, { new_password: newPassword.value })
    toast.add({
      severity: 'success',
      summary: 'Password Reset',
      detail: `Password for ${props.user.name} has been reset successfully.`,
      life: 3000,
    })
    emit('success')
    handleClose()
  } catch (err: unknown) {
    if (err instanceof Error) {
      errorMessage.value = err.message
    } else {
      errorMessage.value = 'Failed to reset password'
    }
  }
}
</script>

<template>
  <Dialog
    :visible="visible"
    modal
    header="Reset User Password"
    :style="{ width: '440px', maxWidth: '95vw' }"
    :closable="!userStore.isActionLoading"
    class="border border-slate-800 rounded-2xl bg-slate-950 text-slate-100 shadow-2xl"
    @update:visible="handleClose"
  >
    <div class="space-y-4 pt-2">
      <!-- Info notice -->
      <div class="p-3 rounded-xl bg-sky-950/40 border border-sky-800/60 text-xs text-sky-300 flex items-start gap-2.5">
        <i class="pi pi-key text-sky-400 mt-0.5"></i>
        <div>
          <span>Resetting password for <b class="text-white">{{ user?.name }}</b> ({{ user?.email }}).</span>
          <p class="text-[11px] text-sky-400/80 mt-1">The user will be required to authenticate with this new password on next sign in.</p>
        </div>
      </div>

      <!-- Error banner -->
      <div
        v-if="errorMessage"
        class="p-3 rounded-xl bg-rose-950/50 border border-rose-800/80 text-rose-300 text-xs flex items-center gap-2"
      >
        <i class="pi pi-exclamation-circle text-rose-400"></i>
        <span>{{ errorMessage }}</span>
      </div>

      <!-- New Password -->
      <div>
        <label class="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
          New Password <span class="text-rose-400">*</span>
        </label>
        <InputText
          v-model="newPassword"
          type="password"
          placeholder="Enter new password (min 6 characters)"
          class="w-full text-sm bg-slate-900 border-slate-700 text-slate-100"
          required
        />
      </div>

      <!-- Confirm Password -->
      <div>
        <label class="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
          Confirm Password <span class="text-rose-400">*</span>
        </label>
        <InputText
          v-model="confirmPassword"
          type="password"
          placeholder="Repeat new password"
          class="w-full text-sm bg-slate-900 border-slate-700 text-slate-100"
          required
        />
      </div>
    </div>

    <template #footer>
      <div class="flex items-center justify-end gap-2 pt-4 border-t border-slate-800/80">
        <Button
          label="Cancel"
          severity="secondary"
          text
          :disabled="userStore.isActionLoading"
          class="text-slate-400 hover:text-white text-xs cursor-pointer"
          @click="handleClose"
        />
        <Button
          label="Reset Password"
          :loading="userStore.isActionLoading"
          icon="pi pi-check"
          class="bg-amber-600 hover:bg-amber-500 text-white font-semibold text-xs border-none cursor-pointer"
          @click="handleSubmit"
        />
      </div>
    </template>
  </Dialog>
</template>

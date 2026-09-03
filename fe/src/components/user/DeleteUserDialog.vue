<script setup lang="ts">
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import { useToast } from 'primevue/usetoast'
import { ref } from 'vue'

import { useUserStore } from '@/stores'
import type { User } from '@/types'

const props = defineProps<{
  visible: boolean
  user: User | null
}>()

const emit = defineEmits<{
  (e: 'update:visible', val: boolean): void
  (e: 'deleted'): void
}>()

const userStore = useUserStore()
const toast = useToast()
const isHardDelete = ref(false)
const errorMessage = ref('')

const handleClose = () => {
  isHardDelete.value = false
  errorMessage.value = ''
  emit('update:visible', false)
}

const handleConfirm = async () => {
  if (!props.user) return
  errorMessage.value = ''

  try {
    if (isHardDelete.value) {
      await userStore.hardDeleteUser(props.user.id)
      toast.add({
        severity: 'info',
        summary: 'User Deleted',
        detail: `User ${props.user.name} permanently removed.`,
        life: 3000,
      })
    } else {
      await userStore.deleteUser(props.user.id)
      toast.add({
        severity: 'warn',
        summary: 'User Deactivated',
        detail: `User ${props.user.name} deactivated.`,
        life: 3000,
      })
    }

    emit('deleted')
    handleClose()
  } catch (err: unknown) {
    if (err instanceof Error) {
      errorMessage.value = err.message
    } else {
      errorMessage.value = 'Failed to delete user'
    }
  }
}
</script>

<template>
  <Dialog
    :visible="visible"
    modal
    header="Delete User Account"
    :style="{ width: '460px', maxWidth: '95vw' }"
    :closable="!userStore.isActionLoading"
    class="border border-slate-800 rounded-2xl bg-slate-950 text-slate-100 shadow-2xl"
    @update:visible="handleClose"
  >
    <div class="space-y-4 pt-2">
      <div class="flex items-start gap-3 p-4 rounded-xl bg-rose-950/40 border border-rose-800/60 text-xs text-rose-200">
        <div class="w-8 h-8 rounded-lg bg-rose-500/20 text-rose-400 flex items-center justify-center shrink-0">
          <i class="pi pi-exclamation-triangle text-base"></i>
        </div>
        <div>
          <h4 class="font-bold text-sm text-rose-300">Are you sure?</h4>
          <p class="mt-1 text-slate-300">
            You are about to remove access for user <b class="text-white">{{ user?.name }}</b> (<span class="font-mono">{{ user?.email }}</span>).
          </p>
        </div>
      </div>

      <!-- Error banner -->
      <div
        v-if="errorMessage"
        class="p-3 rounded-xl bg-rose-950/60 border border-rose-800 text-rose-300 text-xs flex items-center gap-2"
      >
        <i class="pi pi-exclamation-circle text-rose-400"></i>
        <span>{{ errorMessage }}</span>
      </div>

      <!-- Mode selection -->
      <div class="space-y-2 pt-1">
        <label
          class="flex items-center gap-3 p-3 rounded-xl border transition cursor-pointer"
          :class="!isHardDelete
            ? 'bg-slate-900 border-amber-500/50 text-white'
            : 'bg-slate-950 border-slate-800 text-slate-400 hover:border-slate-700'"
        >
          <input v-model="isHardDelete" type="radio" :value="false" class="text-amber-500 focus:ring-0" />
          <div class="text-xs">
            <span class="font-bold text-amber-400">Deactivate Account (Soft Delete - Recommended)</span>
            <p class="text-slate-400 mt-0.5">Revokes login credentials and tokens while preserving audit history.</p>
          </div>
        </label>

        <label
          class="flex items-center gap-3 p-3 rounded-xl border transition cursor-pointer"
          :class="isHardDelete
            ? 'bg-slate-900 border-rose-500/50 text-white'
            : 'bg-slate-950 border-slate-800 text-slate-400 hover:border-slate-700'"
        >
          <input v-model="isHardDelete" type="radio" :value="true" class="text-rose-500 focus:ring-0" />
          <div class="text-xs">
            <span class="font-bold text-rose-400">Permanently Delete (Hard Delete)</span>
            <p class="text-slate-400 mt-0.5">Completely purges user records from the database. Cannot be undone.</p>
          </div>
        </label>
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
          :label="isHardDelete ? 'Permanently Delete' : 'Deactivate User'"
          :loading="userStore.isActionLoading"
          icon="pi pi-trash"
          :class="isHardDelete
            ? 'bg-rose-600 hover:bg-rose-500 text-white font-semibold text-xs border-none cursor-pointer'
            : 'bg-amber-600 hover:bg-amber-500 text-white font-semibold text-xs border-none cursor-pointer'"
          @click="handleConfirm"
        />
      </div>
    </template>
  </Dialog>
</template>

<script setup lang="ts">
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import { useToast } from 'primevue/usetoast'
import { computed, reactive, ref, watch } from 'vue'

import { useUserStore } from '@/stores'
import type { CreateUserPayload, UpdateUserPayload, User, UserRole } from '@/types'

const props = defineProps<{
  visible: boolean
  user: User | null
}>()

const emit = defineEmits<{
  (e: 'update:visible', val: boolean): void
  (e: 'saved'): void
}>()

const userStore = useUserStore()
const toast = useToast()

const isEditMode = computed(() => !!props.user)
const dialogTitle = computed(() => (isEditMode.value ? 'Edit User' : 'Create New User'))

interface FormState {
  name: string
  email: string
  password: string
  role: UserRole
  phone: string
  is_active: boolean
}

const form = reactive<FormState>({
  name: '',
  email: '',
  password: '',
  role: 'viewer',
  phone: '',
  is_active: true,
})

const errorMessage = ref('')

const roleOptions = [
  {
    label: 'Cluster Admin (Full Access)',
    value: 'admin' as UserRole,
    desc: 'Full cluster control, RBAC, credentials & user management',
    badgeClass: 'bg-purple-500/10 text-purple-400 border-purple-500/30',
  },
  {
    label: 'DevOps Engineer (Read/Write)',
    value: 'devops' as UserRole,
    desc: 'Deploy, scale, restart workloads, inspect logs & web terminal',
    badgeClass: 'bg-sky-500/10 text-sky-400 border-sky-500/30',
  },
  {
    label: 'Viewer (Read-Only)',
    value: 'viewer' as UserRole,
    desc: 'Observe cluster telemetry, pods status, events & resource metrics',
    badgeClass: 'bg-slate-500/10 text-slate-400 border-slate-500/30',
  },
]

watch(
  () => props.visible,
  (isOpen) => {
    if (isOpen) {
      errorMessage.value = ''
      if (props.user) {
        form.name = props.user.name
        form.email = props.user.email
        form.password = ''
        form.role = props.user.role
        form.phone = props.user.phone || ''
        form.is_active = props.user.is_active
      } else {
        form.name = ''
        form.email = ''
        form.password = ''
        form.role = 'viewer'
        form.phone = ''
        form.is_active = true
      }
    }
  },
)

const handleClose = () => {
  emit('update:visible', false)
}

const handleSubmit = async () => {
  errorMessage.value = ''
  if (!form.name.trim()) {
    errorMessage.value = 'Full name is required'
    return
  }
  if (!form.email.trim()) {
    errorMessage.value = 'Email address is required'
    return
  }

  if (!isEditMode.value && (!form.password || form.password.length < 6)) {
    errorMessage.value = 'Password must be at least 6 characters'
    return
  }

  try {
    if (isEditMode.value && props.user) {
      const payload: UpdateUserPayload = {
        name: form.name.trim(),
        email: form.email.trim(),
        role: form.role,
        phone: form.phone.trim() || undefined,
        is_active: form.is_active,
      }
      await userStore.updateUser(props.user.id, payload)
      toast.add({
        severity: 'success',
        summary: 'User Updated',
        detail: `User ${form.name} updated successfully`,
        life: 3000,
      })
    } else {
      const payload: CreateUserPayload = {
        name: form.name.trim(),
        email: form.email.trim(),
        password: form.password,
        role: form.role,
        phone: form.phone.trim() || undefined,
      }
      await userStore.createUser(payload)
      toast.add({
        severity: 'success',
        summary: 'User Created',
        detail: `User ${form.name} created successfully`,
        life: 3000,
      })
    }

    emit('saved')
    handleClose()
  } catch (err: unknown) {
    if (err instanceof Error) {
      errorMessage.value = err.message
    } else {
      errorMessage.value = 'Failed to save user'
    }
  }
}
</script>

<template>
  <Dialog
    :visible="visible"
    modal
    :header="dialogTitle"
    :style="{ width: '520px', maxWidth: '95vw' }"
    :closable="!userStore.isActionLoading"
    class="border border-slate-800 rounded-2xl bg-slate-950 text-slate-100 shadow-2xl"
    @update:visible="handleClose"
  >
    <div class="space-y-4 pt-2">
      <!-- Error banner -->
      <div
        v-if="errorMessage"
        class="p-3 rounded-xl bg-rose-950/50 border border-rose-800/80 text-rose-300 text-xs flex items-center gap-2"
      >
        <i class="pi pi-exclamation-circle text-rose-400"></i>
        <span>{{ errorMessage }}</span>
      </div>

      <!-- Full Name -->
      <div>
        <label class="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
          Full Name <span class="text-rose-400">*</span>
        </label>
        <InputText
          v-model="form.name"
          placeholder="e.g. Alex Morgan"
          class="w-full text-sm bg-slate-900 border-slate-700 text-slate-100"
          required
        />
      </div>

      <!-- Email -->
      <div>
        <label class="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
          Email Address <span class="text-rose-400">*</span>
        </label>
        <InputText
          v-model="form.email"
          type="email"
          placeholder="alex@example.com"
          class="w-full text-sm bg-slate-900 border-slate-700 text-slate-100"
          required
        />
      </div>

      <!-- Password (Create only) -->
      <div v-if="!isEditMode">
        <label class="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
          Password <span class="text-rose-400">*</span>
        </label>
        <InputText
          v-model="form.password"
          type="password"
          placeholder="Minimum 6 characters"
          class="w-full text-sm bg-slate-900 border-slate-700 text-slate-100"
          required
        />
      </div>

      <!-- Role Selection -->
      <div>
        <label class="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
          Role / Authorization Level <span class="text-rose-400">*</span>
        </label>
        <Select
          v-model="form.role"
          :options="roleOptions"
          option-label="label"
          option-value="value"
          placeholder="Select a role"
          class="w-full bg-slate-900 border-slate-700 text-slate-100 text-sm"
        >
          <template #option="slotProps">
            <div class="py-1">
              <div class="flex items-center gap-2">
                <span class="font-semibold text-sm text-slate-100">{{ slotProps.option.label }}</span>
                <span
                  class="px-2 py-0.5 rounded text-[10px] font-mono border"
                  :class="slotProps.option.badgeClass"
                >
                  {{ slotProps.option.value }}
                </span>
              </div>
              <p class="text-xs text-slate-400 mt-0.5">{{ slotProps.option.desc }}</p>
            </div>
          </template>
        </Select>
      </div>

      <!-- Phone -->
      <div>
        <label class="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
          Phone Number <span class="text-slate-500 font-normal">(Optional)</span>
        </label>
        <InputText
          v-model="form.phone"
          placeholder="e.g. 08123456789"
          class="w-full text-sm bg-slate-900 border-slate-700 text-slate-100"
        />
      </div>

      <!-- Status Toggle (Edit only) -->
      <div v-if="isEditMode" class="pt-2 border-t border-slate-800 flex items-center justify-between">
        <div>
          <span class="text-sm font-semibold text-slate-200">Account Status</span>
          <p class="text-xs text-slate-400">Deactivated users cannot authenticate or access cluster API.</p>
        </div>
        <button
          type="button"
          class="px-3 py-1.5 rounded-lg text-xs font-semibold border transition cursor-pointer"
          :class="form.is_active
            ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/30 hover:bg-emerald-500/20'
            : 'bg-rose-500/10 text-rose-400 border-rose-500/30 hover:bg-rose-500/20'"
          @click="form.is_active = !form.is_active"
        >
          <i class="pi" :class="form.is_active ? 'pi-check-circle' : 'pi-ban'"></i>
          {{ form.is_active ? 'Active' : 'Inactive' }}
        </button>
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
          :label="isEditMode ? 'Save Changes' : 'Create User'"
          :loading="userStore.isActionLoading"
          icon="pi pi-check"
          class="bg-emerald-600 hover:bg-emerald-500 text-white font-semibold text-xs border-none cursor-pointer"
          @click="handleSubmit"
        />
      </div>
    </template>
  </Dialog>
</template>

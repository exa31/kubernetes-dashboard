<script setup lang="ts">
import Button from 'primevue/button'
import Column from 'primevue/column'
import DataTable from 'primevue/datatable'
import IconField from 'primevue/iconfield'
import InputIcon from 'primevue/inputicon'
import InputText from 'primevue/inputtext'
import Toast from 'primevue/toast'
import { onMounted, ref } from 'vue'

import { DeleteUserDialog, ResetPasswordDialog, UserFormDialog } from '@/components/user'
import { useAuthStore, useUserStore } from '@/stores'
import type { User, UserRole } from '@/types'

const userStore = useUserStore()
const authStore = useAuthStore()

// Dialog states
const formDialogVisible = ref(false)
const resetPasswordDialogVisible = ref(false)
const deleteDialogVisible = ref(false)
const selectedUser = ref<User | null>(null)

onMounted(() => {
  userStore.fetchUsers()
})

const handleRefresh = () => {
  userStore.fetchUsers()
}

const openCreateDialog = () => {
  selectedUser.value = null
  formDialogVisible.value = true
}

const openEditDialog = (user: User) => {
  selectedUser.value = user
  formDialogVisible.value = true
}

const openResetPasswordDialog = (user: User) => {
  selectedUser.value = user
  resetPasswordDialogVisible.value = true
}

const openDeleteDialog = (user: User) => {
  selectedUser.value = user
  deleteDialogVisible.value = true
}

const getInitials = (name: string) => {
  if (!name) return 'U'
  return name
    .split(' ')
    .map((n) => n[0])
    .join('')
    .substring(0, 2)
    .toUpperCase()
}

const formatDate = (dateStr?: string) => {
  if (!dateStr) return '-'
  try {
    const d = new Date(dateStr)
    return d.toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
    })
  } catch {
    return dateStr
  }
}

const roleBadgeConfig = (role: UserRole) => {
  switch (role) {
    case 'admin':
      return {
        label: 'Admin',
        icon: 'pi-shield',
        classes: 'bg-purple-500/10 text-purple-400 border-purple-500/30',
      }
    case 'devops':
      return {
        label: 'DevOps',
        icon: 'pi-wrench',
        classes: 'bg-sky-500/10 text-sky-400 border-sky-500/30',
      }
    case 'viewer':
    default:
      return {
        label: 'Viewer',
        icon: 'pi-eye',
        classes: 'bg-slate-500/10 text-slate-400 border-slate-500/30',
      }
  }
}
</script>

<template>
  <div class="space-y-6">
    <Toast position="top-right" />

    <!-- Hero Header Banner -->
    <div
      class="relative overflow-hidden rounded-2xl bg-gradient-to-r from-slate-900 via-slate-900/90 to-purple-950/40 border border-slate-800/80 p-6 shadow-xl"
    >
      <div class="relative z-10 flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <div class="flex items-center gap-3">
            <div
              class="w-10 h-10 rounded-xl bg-purple-500/10 border border-purple-500/30 flex items-center justify-center text-purple-400 shadow-inner"
            >
              <i class="pi pi-users text-lg"></i>
            </div>
            <div>
              <div class="flex items-center gap-2.5">
                <h1 class="text-xl font-extrabold text-white tracking-tight">Team & Access Control (RBAC)</h1>
                <span
                  class="px-2.5 py-0.5 rounded-full text-xs font-semibold bg-purple-500/10 text-purple-400 border border-purple-500/30 flex items-center gap-1.5"
                >
                  <span class="w-2 h-2 rounded-full bg-purple-400 animate-pulse"></span>
                  Role-Based Security
                </span>
              </div>
              <p class="text-xs text-slate-400 mt-1 font-mono">
                Manage team authorizations, cluster privileges, and user access levels across environments.
              </p>
            </div>
          </div>
        </div>

        <!-- Action Buttons -->
        <div class="flex items-center gap-2.5">
          <Button
            icon="pi pi-refresh"
            :loading="userStore.isLoading"
            severity="secondary"
            class="bg-slate-800/80 hover:bg-slate-700 text-slate-300 border-slate-700 text-xs cursor-pointer"
            label="Refresh"
            @click="handleRefresh"
          />

          <Button
            v-if="authStore.isAdmin"
            icon="pi pi-user-plus"
            label="Add Member"
            class="bg-emerald-600 hover:bg-emerald-500 text-white font-semibold text-xs border-none shadow-sm cursor-pointer"
            @click="openCreateDialog"
          />
        </div>
      </div>
    </div>

    <!-- Non-admin Notice Banner -->
    <div
      v-if="!authStore.isAdmin"
      class="p-4 rounded-xl bg-slate-900/60 border border-slate-800 flex items-center justify-between text-xs text-slate-300"
    >
      <div class="flex items-center gap-2.5">
        <i class="pi pi-info-circle text-sky-400 text-sm"></i>
        <span>
          You are signed in as <b class="text-white uppercase font-mono">{{ authStore.userRole }}</b>. User creation and privilege management are restricted to Cluster Administrators.
        </span>
      </div>
      <span class="px-2 py-0.5 rounded text-[11px] font-mono bg-slate-800 text-slate-400 border border-slate-700">
        Read-Only Access
      </span>
    </div>

    <!-- Stats Overview Cards -->
    <div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-3.5">
      <!-- Total Users -->
      <div class="p-4 rounded-xl bg-slate-900/70 border border-slate-800 shadow-sm">
        <div class="flex items-center justify-between text-slate-400 text-xs mb-1.5">
          <span>Total Members</span>
          <i class="pi pi-users text-slate-500"></i>
        </div>
        <div class="text-2xl font-bold font-mono text-white">{{ userStore.totalUsers }}</div>
        <div class="text-[11px] text-slate-500 mt-1">Platform accounts</div>
      </div>

      <!-- Admins -->
      <div class="p-4 rounded-xl bg-slate-900/70 border border-purple-900/30 shadow-sm">
        <div class="flex items-center justify-between text-purple-300 text-xs mb-1.5">
          <span>Cluster Admins</span>
          <i class="pi pi-shield text-purple-400"></i>
        </div>
        <div class="text-2xl font-bold font-mono text-purple-400">{{ userStore.adminCount }}</div>
        <div class="text-[11px] text-slate-500 mt-1">Full cluster root</div>
      </div>

      <!-- DevOps -->
      <div class="p-4 rounded-xl bg-slate-900/70 border border-sky-900/30 shadow-sm">
        <div class="flex items-center justify-between text-sky-300 text-xs mb-1.5">
          <span>DevOps Engineers</span>
          <i class="pi pi-wrench text-sky-400"></i>
        </div>
        <div class="text-2xl font-bold font-mono text-sky-400">{{ userStore.devopsCount }}</div>
        <div class="text-[11px] text-slate-500 mt-1">Deploy & Scale</div>
      </div>

      <!-- Viewers -->
      <div class="p-4 rounded-xl bg-slate-900/70 border border-slate-800 shadow-sm">
        <div class="flex items-center justify-between text-slate-400 text-xs mb-1.5">
          <span>Viewers</span>
          <i class="pi pi-eye text-slate-500"></i>
        </div>
        <div class="text-2xl font-bold font-mono text-slate-300">{{ userStore.viewerCount }}</div>
        <div class="text-[11px] text-slate-500 mt-1">Telemetry observe</div>
      </div>

      <!-- Active Users -->
      <div class="p-4 rounded-xl bg-slate-900/70 border border-emerald-900/30 shadow-sm col-span-2 sm:col-span-1">
        <div class="flex items-center justify-between text-emerald-300 text-xs mb-1.5">
          <span>Active Status</span>
          <i class="pi pi-check-circle text-emerald-400"></i>
        </div>
        <div class="text-2xl font-bold font-mono text-emerald-400">{{ userStore.activeCount }}</div>
        <div class="text-[11px] text-slate-500 mt-1">Enabled credentials</div>
      </div>
    </div>

    <!-- Filter & Search Toolbar -->
    <div class="p-4 rounded-2xl bg-slate-900/80 border border-slate-800 shadow-sm flex flex-col md:flex-row items-center justify-between gap-4">
      <!-- Search Input -->
      <IconField class="w-full md:w-80">
        <InputIcon class="pi pi-search text-xs" />
        <InputText
          v-model="userStore.searchQuery"
          placeholder="Search member by name or email..."
          class="w-full text-xs"
        />
      </IconField>

      <!-- Filters -->
      <div class="flex items-center gap-2 flex-wrap w-full md:w-auto justify-start md:justify-end">
        <!-- Role filter tabs -->
        <div class="p-1 rounded-xl bg-slate-950 border border-slate-800 flex items-center text-xs">
          <button
            type="button"
            class="px-3 py-1 rounded-lg font-medium transition cursor-pointer"
            :class="userStore.roleFilter === 'all'
              ? 'bg-slate-800 text-white font-semibold shadow-xs'
              : 'text-slate-400 hover:text-slate-200'"
            @click="userStore.roleFilter = 'all'"
          >
            All Roles
          </button>
          <button
            type="button"
            class="px-3 py-1 rounded-lg font-medium transition cursor-pointer"
            :class="userStore.roleFilter === 'admin'
              ? 'bg-purple-900/50 text-purple-300 font-semibold shadow-xs'
              : 'text-slate-400 hover:text-slate-200'"
            @click="userStore.roleFilter = 'admin'"
          >
            Admin
          </button>
          <button
            type="button"
            class="px-3 py-1 rounded-lg font-medium transition cursor-pointer"
            :class="userStore.roleFilter === 'devops'
              ? 'bg-sky-900/50 text-sky-300 font-semibold shadow-xs'
              : 'text-slate-400 hover:text-slate-200'"
            @click="userStore.roleFilter = 'devops'"
          >
            DevOps
          </button>
          <button
            type="button"
            class="px-3 py-1 rounded-lg font-medium transition cursor-pointer"
            :class="userStore.roleFilter === 'viewer'
              ? 'bg-slate-800 text-slate-300 font-semibold shadow-xs'
              : 'text-slate-400 hover:text-slate-200'"
            @click="userStore.roleFilter = 'viewer'"
          >
            Viewer
          </button>
        </div>

        <!-- Status Filter -->
        <div class="p-1 rounded-xl bg-slate-950 border border-slate-800 flex items-center text-xs">
          <button
            type="button"
            class="px-2.5 py-1 rounded-lg font-medium transition cursor-pointer"
            :class="userStore.statusFilter === 'all'
              ? 'bg-slate-800 text-white font-semibold'
              : 'text-slate-400 hover:text-slate-200'"
            @click="userStore.statusFilter = 'all'"
          >
            All
          </button>
          <button
            type="button"
            class="px-2.5 py-1 rounded-lg font-medium transition cursor-pointer"
            :class="userStore.statusFilter === 'active'
              ? 'bg-emerald-900/50 text-emerald-300 font-semibold'
              : 'text-slate-400 hover:text-slate-200'"
            @click="userStore.statusFilter = 'active'"
          >
            Active
          </button>
          <button
            type="button"
            class="px-2.5 py-1 rounded-lg font-medium transition cursor-pointer"
            :class="userStore.statusFilter === 'inactive'
              ? 'bg-rose-900/50 text-rose-300 font-semibold'
              : 'text-slate-400 hover:text-slate-200'"
            @click="userStore.statusFilter = 'inactive'"
          >
            Inactive
          </button>
        </div>
      </div>
    </div>

    <!-- Data Table Container -->
    <div class="rounded-2xl bg-slate-900/90 border border-slate-800 overflow-hidden shadow-xl">
      <DataTable
        :value="userStore.filteredUsers"
        :loading="userStore.isLoading"
        responsive-layout="scroll"
        class="text-xs"
        row-hover
      >
        <template #empty>
          <div class="py-12 text-center text-slate-500">
            <i class="pi pi-users text-3xl text-slate-600 mb-2"></i>
            <p class="font-medium text-sm text-slate-400">No team members found</p>
            <p class="text-xs text-slate-500 mt-1">Try adjusting your search query or filter settings.</p>
          </div>
        </template>

        <!-- Member Column -->
        <Column header="Member">
          <template #body="{ data }">
            <div class="flex items-center gap-3 py-1">
              <div
                class="w-8 h-8 rounded-full flex items-center justify-center text-xs font-bold font-mono shrink-0 shadow-inner"
                :class="data.role === 'admin'
                  ? 'bg-purple-500/20 text-purple-300 border border-purple-500/40'
                  : data.role === 'devops'
                    ? 'bg-sky-500/20 text-sky-300 border border-sky-500/40'
                    : 'bg-slate-800 text-slate-300 border border-slate-700'"
              >
                {{ getInitials(data.name) }}
              </div>
              <div>
                <div class="font-bold text-white flex items-center gap-2">
                  <span>{{ data.name }}</span>
                  <span
                    v-if="data.email === authStore.user?.email"
                    class="px-1.5 py-0.2 rounded text-[10px] font-mono bg-emerald-500/20 text-emerald-300 border border-emerald-500/30"
                  >
                    You
                  </span>
                </div>
                <div class="text-[11px] font-mono text-slate-400">{{ data.email }}</div>
              </div>
            </div>
          </template>
        </Column>

        <!-- Role / RBAC Column -->
        <Column header="Role & Privileges">
          <template #body="{ data }">
            <span
              class="px-2.5 py-1 rounded-full text-xs font-semibold border flex items-center gap-1.5 w-fit"
              :class="roleBadgeConfig(data.role).classes"
            >
              <i class="pi" :class="roleBadgeConfig(data.role).icon"></i>
              <span>{{ roleBadgeConfig(data.role).label }}</span>
            </span>
          </template>
        </Column>

        <!-- Status Column -->
        <Column header="Status">
          <template #body="{ data }">
            <span
              class="px-2 py-0.5 rounded text-[11px] font-semibold border inline-flex items-center gap-1.5"
              :class="data.is_active
                ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/30'
                : 'bg-rose-500/10 text-rose-400 border-rose-500/30'"
            >
              <span
                class="w-1.5 h-1.5 rounded-full"
                :class="data.is_active ? 'bg-emerald-400 animate-pulse' : 'bg-rose-500'"
              ></span>
              {{ data.is_active ? 'Active' : 'Disabled' }}
            </span>
          </template>
        </Column>

        <!-- Phone Column -->
        <Column header="Phone">
          <template #body="{ data }">
            <span class="font-mono text-slate-400 text-xs">{{ data.phone || '-' }}</span>
          </template>
        </Column>

        <!-- Created At Column -->
        <Column header="Joined">
          <template #body="{ data }">
            <span class="font-mono text-slate-400 text-xs">{{ formatDate(data.created_at) }}</span>
          </template>
        </Column>

        <!-- Action Column -->
        <Column header="Actions" align-frozen="right" frozen>
          <template #body="{ data }">
            <div class="flex items-center gap-1.5 justify-end">
              <!-- Edit -->
              <Button
                icon="pi pi-pencil"
                text
                rounded
                severity="secondary"
                size="small"
                v-tooltip.top="'Edit Member'"
                :disabled="!authStore.isAdmin"
                class="text-slate-400 hover:text-white hover:bg-slate-800/80 cursor-pointer"
                @click="openEditDialog(data)"
              />

              <!-- Reset Password -->
              <Button
                icon="pi pi-key"
                text
                rounded
                severity="secondary"
                size="small"
                v-tooltip.top="'Reset Password'"
                :disabled="!authStore.isAdmin"
                class="text-slate-400 hover:text-amber-400 hover:bg-slate-800/80 cursor-pointer"
                @click="openResetPasswordDialog(data)"
              />

              <!-- Delete -->
              <Button
                icon="pi pi-trash"
                text
                rounded
                severity="secondary"
                size="small"
                v-tooltip.top="'Delete / Deactivate'"
                :disabled="!authStore.isAdmin || data.email === authStore.user?.email"
                class="text-slate-400 hover:text-rose-400 hover:bg-slate-800/80 cursor-pointer"
                @click="openDeleteDialog(data)"
              />
            </div>
          </template>
        </Column>
      </DataTable>
    </div>

    <!-- Modals -->
    <UserFormDialog
      v-model:visible="formDialogVisible"
      :user="selectedUser"
      @saved="handleRefresh"
    />

    <ResetPasswordDialog
      v-model:visible="resetPasswordDialogVisible"
      :user="selectedUser"
      @success="handleRefresh"
    />

    <DeleteUserDialog
      v-model:visible="deleteDialogVisible"
      :user="selectedUser"
      @deleted="handleRefresh"
    />
  </div>
</template>

<style scoped>
:deep(.p-datatable-header) {
  background: transparent;
}
:deep(.p-datatable-thead > tr > th) {
  background: rgba(2, 6, 23, 0.7);
  color: #94a3b8;
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  border-bottom: 1px solid #1e293b;
  padding: 12px 16px;
}
:deep(.p-datatable-tbody > tr) {
  background: transparent;
  border-bottom: 1px solid rgba(30, 41, 59, 0.6);
  transition: background-color 0.15s;
}
:deep(.p-datatable-tbody > tr:hover) {
  background: rgba(30, 41, 59, 0.4);
}
:deep(.p-datatable-tbody > tr > td) {
  padding: 10px 16px;
  border: none;
}
</style>

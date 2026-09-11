<script setup lang="ts">
import { storeToRefs } from 'pinia'
import Button from 'primevue/button'
import Column from 'primevue/column'
import DataTable from 'primevue/datatable'
import Dialog from 'primevue/dialog'
import IconField from 'primevue/iconfield'
import InputIcon from 'primevue/inputicon'
import InputText from 'primevue/inputtext'
import Tag from 'primevue/tag'
import { useConfirm } from 'primevue/useconfirm'
import { useToast } from 'primevue/usetoast'
import { computed, onMounted, ref } from 'vue'

import EnvEditor from '@/features/k8s/EnvEditor.vue'
import { useAuthStore, useK8sStore } from '@/stores'
import type { SecretDetail, SecretItem } from '@/types'

const authStore = useAuthStore()
const k8sStore = useK8sStore()
const { secrets, selectedNamespace, isLoading, isActionLoading } = storeToRefs(k8sStore)
const confirm = useConfirm()
const toast = useToast()

const canMutate = computed(() => authStore.canMutateNamespace(selectedNamespace.value))

const searchQuery = ref('')
const activeSecretDetail = ref<SecretDetail | null>(null)
const isEditorOpen = ref(false)

// Create Dialog
const isCreateOpen = ref(false)
const newSecretName = ref('')
const newSecretType = ref<
  | 'Opaque'
  | 'kubernetes.io/tls'
  | 'kubernetes.io/dockerconfigjson'
  | 'kubernetes.io/basic-auth'
  | 'kubernetes.io/ssh-auth'
>('Opaque')

// Opaque state
const opaqueMode = ref<'keyvalue' | 'env'>('keyvalue')
const opaqueKeyValues = ref<Array<{ key: string; value: string; hidden: boolean }>>([
  { key: '', value: '', hidden: false }
])
const newSecretEnvText = ref('KEY=value\nANOTHER_KEY=123')

// TLS state
const tlsCert = ref('')
const tlsKey = ref('')

// Docker registry state
const dockerServer = ref('https://index.docker.io/v1/')
const dockerUsername = ref('')
const dockerPassword = ref('')
const dockerEmail = ref('')
const dockerPasswordHidden = ref(true)

const dockerPresets = [
  { label: 'Docker Hub', server: 'https://index.docker.io/v1/' },
  { label: 'GitHub (ghcr.io)', server: 'ghcr.io' },
  { label: 'Google Container Registry', server: 'gcr.io' },
  { label: 'Quay.io', server: 'quay.io' }
]

// Basic Auth state
const basicAuthUsername = ref('')
const basicAuthPassword = ref('')
const basicAuthPasswordHidden = ref(true)

// SSH Auth state
const sshPrivateKey = ref('')
const sshPublicKey = ref('')
const sshKnownHosts = ref('')

function addOpaqueRow() {
  opaqueKeyValues.value.push({ key: '', value: '', hidden: false })
}

function removeOpaqueRow(idx: number) {
  opaqueKeyValues.value.splice(idx, 1)
}

function resetCreateForm() {
  newSecretName.value = ''
  newSecretType.value = 'Opaque'
  opaqueMode.value = 'keyvalue'
  opaqueKeyValues.value = [{ key: '', value: '', hidden: false }]
  newSecretEnvText.value = 'KEY=value\nANOTHER_KEY=123'
  tlsCert.value = ''
  tlsKey.value = ''
  dockerServer.value = 'https://index.docker.io/v1/'
  dockerUsername.value = ''
  dockerPassword.value = ''
  dockerEmail.value = ''
  dockerPasswordHidden.value = true
  basicAuthUsername.value = ''
  basicAuthPassword.value = ''
  basicAuthPasswordHidden.value = true
  sshPrivateKey.value = ''
  sshPublicKey.value = ''
  sshKnownHosts.value = ''
}

onMounted(() => {
  k8sStore.fetchSecrets()
})

const filteredSecrets = computed(() => {
  const q = searchQuery.value.trim().toLowerCase()
  if (!q) return secrets.value
  return secrets.value.filter(
    (s) => s.name.toLowerCase().includes(q) || s.type.toLowerCase().includes(q)
  )
})

const openSecret = async (item: SecretItem) => {
  try {
    const detail = await k8sStore.getSecretDetail(item.name)
    activeSecretDetail.value = detail
    isEditorOpen.value = true
  } catch {
    // handled in store
  }
}

const deleteSecret = (item: SecretItem) => {
  confirm.require({
    message: `Are you sure you want to delete secret '${item.name}' from namespace '${selectedNamespace.value}'?`,
    header: 'Delete Secret',
    icon: 'pi pi-exclamation-triangle',
    rejectProps: {
      label: 'Cancel',
      severity: 'secondary',
      outlined: true
    },
    acceptProps: {
      label: 'Delete',
      severity: 'danger'
    },
    accept: async () => {
      try {
        await k8sStore.deleteSecret(item.name)
        toast.add({
          severity: 'success',
          summary: 'Deleted',
          detail: `Secret '${item.name}' deleted successfully`,
          life: 3000
        })
      } catch (err: unknown) {
        toast.add({
          severity: 'error',
          summary: 'Delete Failed',
          detail: err instanceof Error ? err.message : 'Failed to delete Secret',
          life: 4000
        })
      }
    }
  })
}

const createSecret = async () => {
  if (!newSecretName.value.trim()) return

  let data: Record<string, string> = {}

  if (newSecretType.value === 'Opaque') {
    if (opaqueMode.value === 'keyvalue') {
      for (const row of opaqueKeyValues.value) {
        if (row.key.trim()) {
          data[row.key.trim()] = row.value
        }
      }
    } else {
      const lines = newSecretEnvText.value.split('\n')
      for (const line of lines) {
        const trimmed = line.trim()
        if (!trimmed || trimmed.startsWith('#')) continue
        const eqIdx = trimmed.indexOf('=')
        if (eqIdx > 0) {
          const k = trimmed.slice(0, eqIdx).trim()
          let v = trimmed.slice(eqIdx + 1).trim()
          if ((v.startsWith('"') && v.endsWith('"')) || (v.startsWith("'") && v.endsWith("'"))) {
            v = v.slice(1, -1)
          }
          data[k] = v
        }
      }
    }
  } else if (newSecretType.value === 'kubernetes.io/tls') {
    if (!tlsCert.value.trim() || !tlsKey.value.trim()) {
      toast.add({
        severity: 'warn',
        summary: 'Missing Fields',
        detail: 'Both TLS Certificate (tls.crt) and Private Key (tls.key) are required.',
        life: 4000
      })
      return
    }
    data = {
      'tls.crt': tlsCert.value.trim(),
      'tls.key': tlsKey.value.trim()
    }
  } else if (newSecretType.value === 'kubernetes.io/dockerconfigjson') {
    if (!dockerServer.value.trim() || !dockerUsername.value.trim() || !dockerPassword.value.trim()) {
      toast.add({
        severity: 'warn',
        summary: 'Missing Fields',
        detail: 'Registry Server, Username, and Password are required.',
        life: 4000
      })
      return
    }
    const srv = dockerServer.value.trim()
    const usr = dockerUsername.value.trim()
    const pwd = dockerPassword.value.trim()
    const eml = dockerEmail.value.trim()
    const authString = btoa(`${usr}:${pwd}`)
    const authConfig = {
      auths: {
        [srv]: {
          username: usr,
          password: pwd,
          email: eml,
          auth: authString
        }
      }
    }
    data = {
      '.dockerconfigjson': JSON.stringify(authConfig, null, 2)
    }
  } else if (newSecretType.value === 'kubernetes.io/basic-auth') {
    if (!basicAuthUsername.value.trim() || !basicAuthPassword.value.trim()) {
      toast.add({
        severity: 'warn',
        summary: 'Missing Fields',
        detail: 'Username and Password are required for Basic Auth.',
        life: 4000
      })
      return
    }
    data = {
      username: basicAuthUsername.value.trim(),
      password: basicAuthPassword.value.trim()
    }
  } else if (newSecretType.value === 'kubernetes.io/ssh-auth') {
    if (!sshPrivateKey.value.trim()) {
      toast.add({
        severity: 'warn',
        summary: 'Missing Field',
        detail: 'SSH Private Key (ssh-privatekey) is required.',
        life: 4000
      })
      return
    }
    data = {
      'ssh-privatekey': sshPrivateKey.value.trim()
    }
    if (sshPublicKey.value.trim()) {
      data['ssh-publickey'] = sshPublicKey.value.trim()
    }
    if (sshKnownHosts.value.trim()) {
      data['known_hosts'] = sshKnownHosts.value.trim()
    }
  }

  try {
    await k8sStore.saveSecret({
      name: newSecretName.value.trim().toLowerCase(),
      namespace: selectedNamespace.value,
      type: newSecretType.value,
      data
    })
    toast.add({
      severity: 'success',
      summary: 'Success',
      detail: `Secret '${newSecretName.value.trim()}' (${newSecretType.value}) created successfully`,
      life: 3000
    })
    isCreateOpen.value = false
    resetCreateForm()
  } catch (err: unknown) {
    toast.add({
      severity: 'error',
      summary: 'Create Failed',
      detail: err instanceof Error ? err.message : 'Failed to create Secret',
      life: 4000
    })
  }
}
</script>

<template>
  <div class="space-y-4">
    <!-- Top toolbar -->
    <div class="flex flex-wrap items-center justify-between gap-4">
      <div>
        <h1
          class="text-2xl font-bold tracking-tight text-slate-900 dark:text-slate-100 flex items-center gap-2.5"
        >
          <i class="pi pi-lock text-amber-500"></i>
          <span>Secrets</span>
        </h1>
        <p class="text-slate-500 dark:text-slate-400 text-sm mt-1">
          Manage decoded environment variables and credentials in
          <strong class="text-slate-700 dark:text-slate-300 font-mono">{{
            selectedNamespace
          }}</strong>
        </p>
      </div>

      <div class="flex items-center gap-3">
        <IconField>
          <InputIcon class="pi pi-search" />
          <InputText v-model="searchQuery" placeholder="Search secrets..." class="text-sm w-64" />
        </IconField>

        <Button
          label="Refresh"
          icon="pi pi-refresh"
          severity="secondary"
          outlined
          size="small"
          :loading="isLoading"
          @click="k8sStore.fetchSecrets()"
        />

        <Button
          v-if="canMutate"
          label="Create Secret"
          icon="pi pi-plus"
          size="small"
          class="btn-emerald text-xs shadow-xs cursor-pointer"
          @click="isCreateOpen = true"
        />
      </div>
    </div>

    <!-- Read-Only Notice for restricted users / viewers -->
    <div
      v-if="!canMutate"
      class="bg-amber-500/10 border border-amber-500/20 rounded-xl p-3.5 flex items-center justify-between gap-3 text-amber-600 dark:text-amber-400 text-xs"
    >
      <div class="flex items-center gap-2">
        <i class="pi pi-lock text-sm"></i>
        <span>
          <strong>Read-Only Mode:</strong> You do not have permission to modify secrets in namespace
          <strong>{{ selectedNamespace }}</strong
          >.
        </span>
      </div>
      <span
        class="px-2 py-0.5 rounded text-[10px] uppercase font-mono font-semibold bg-amber-500/20 border border-amber-500/30"
      >
        {{ authStore.user?.role || 'Viewer' }}
      </span>
    </div>

    <!-- PrimeVue DataTable for Secrets -->
    <div
      class="w-full bg-white dark:bg-slate-950 rounded-xl border border-slate-200 dark:border-slate-800 shadow-sm overflow-hidden"
    >
      <DataTable
        :value="filteredSecrets"
        :loading="isLoading"
        striped-rows
        paginator
        :rows="10"
        :rows-per-page-options="[10, 20, 50]"
        table-style="min-width: 100%"
        class="p-datatable-sm w-full"
        row-hover
        @row-click="(e) => openSecret(e.data)"
      >
        <!-- Name Column -->
        <Column field="name" header="Name" sortable>
          <template #body="{ data }">
            <div class="flex items-center gap-3 py-1 cursor-pointer">
              <div
                class="w-8 h-8 rounded-lg bg-amber-500/10 text-amber-500 flex items-center justify-center font-bold shrink-0"
              >
                <i class="pi pi-key text-xs"></i>
              </div>
              <div>
                <div
                  class="font-semibold text-slate-900 dark:text-slate-100 font-mono text-sm hover:text-sky-600 transition-colors"
                >
                  {{ data.name }}
                </div>
                <div class="text-xs text-slate-400 mt-0.5 truncate font-mono">
                  {{ data.keys.slice(0, 8).join(', ')
                  }}{{ data.keys.length > 8 ? ` +${data.keys.length - 8} more` : '' }}
                </div>
              </div>
            </div>
          </template>
        </Column>

        <!-- Type Column -->
        <Column field="type" header="Type" sortable style="width: 190px">
          <template #body="{ data }">
            <span
              v-if="data.type === 'Opaque'"
              class="px-2.5 py-1 rounded-md text-xs font-mono font-medium bg-sky-500/10 text-sky-400 border border-sky-500/20 flex items-center gap-1.5 w-fit"
            >
              <i class="pi pi-file text-[10px]"></i>
              <span>Opaque</span>
            </span>
            <span
              v-else-if="data.type === 'kubernetes.io/tls'"
              class="px-2.5 py-1 rounded-md text-xs font-mono font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 flex items-center gap-1.5 w-fit"
            >
              <i class="pi pi-shield text-[10px]"></i>
              <span>TLS Cert</span>
            </span>
            <span
              v-else-if="data.type === 'kubernetes.io/dockerconfigjson'"
              class="px-2.5 py-1 rounded-md text-xs font-mono font-medium bg-purple-500/10 text-purple-400 border border-purple-500/20 flex items-center gap-1.5 w-fit"
            >
              <i class="pi pi-box text-[10px]"></i>
              <span>Registry Creds</span>
            </span>
            <span
              v-else-if="data.type === 'kubernetes.io/basic-auth'"
              class="px-2.5 py-1 rounded-md text-xs font-mono font-medium bg-amber-500/10 text-amber-400 border border-amber-500/20 flex items-center gap-1.5 w-fit"
            >
              <i class="pi pi-user text-[10px]"></i>
              <span>Basic Auth</span>
            </span>
            <span
              v-else-if="data.type === 'kubernetes.io/ssh-auth'"
              class="px-2.5 py-1 rounded-md text-xs font-mono font-medium bg-rose-500/10 text-rose-400 border border-rose-500/20 flex items-center gap-1.5 w-fit"
            >
              <i class="pi pi-key text-[10px]"></i>
              <span>SSH Auth</span>
            </span>
            <Tag
              v-else
              :value="data.type"
              severity="secondary"
              class="font-mono text-xs"
            />
          </template>
        </Column>

        <!-- Variables / Key count -->
        <Column field="key_count" header="Variables" sortable style="width: 140px">
          <template #body="{ data }">
            <span class="font-semibold font-mono text-slate-800 dark:text-slate-200">
              {{ data.key_count }}
            </span>
            <span class="text-xs text-slate-400 ml-1">keys</span>
          </template>
        </Column>

        <!-- Age Column -->
        <Column field="age" header="Age" sortable style="width: 120px">
          <template #body="{ data }">
            <span class="text-xs text-slate-500 font-mono">{{ data.age }}</span>
          </template>
        </Column>

        <!-- Actions Column -->
        <Column
          header="Actions"
          header-style="text-align: right"
          body-style="text-align: right"
          style="width: 180px"
        >
          <template #body="{ data }">
            <div class="flex items-center justify-end gap-1.5" @click.stop>
              <Button
                :label="canMutate ? 'Edit Env' : 'View Env'"
                :icon="canMutate ? 'pi pi-file-edit' : 'pi pi-eye'"
                size="small"
                class="btn-blue text-xs px-3 py-1.5 rounded-lg active:scale-95 cursor-pointer"
                @click="openSecret(data)"
              />
              <Button
                v-if="canMutate"
                icon="pi pi-trash"
                size="small"
                class="btn-rose text-xs px-2.5 py-1.5 rounded-lg active:scale-95 cursor-pointer"
                :disabled="isActionLoading"
                title="Delete secret"
                @click="deleteSecret(data)"
              />
            </div>
          </template>
        </Column>

        <!-- Empty state -->
        <template #empty>
          <div class="py-16 text-center text-slate-400">
            <i class="pi pi-shield text-4xl mb-3 text-slate-300 dark:text-slate-700"></i>
            <h3 class="font-semibold text-slate-700 dark:text-slate-300">No Secrets Found</h3>
            <p class="text-xs text-slate-500 mt-1">
              There are no secrets matching your query in namespace {{ selectedNamespace }}.
            </p>
            <Button
              v-if="canMutate"
              label="Create Secret"
              icon="pi pi-plus"
              size="small"
              class="mt-4 btn-emerald text-xs shadow-xs cursor-pointer"
              @click="isCreateOpen = true"
            />
          </div>
        </template>
      </DataTable>
    </div>

    <!-- Env Editor Dialog Modal -->
    <Dialog
      v-model:visible="isEditorOpen"
      modal
      :show-header="false"
      class="w-[95vw] max-w-6xl h-[85vh] rounded-2xl overflow-hidden shadow-2xl border border-slate-200 dark:border-slate-800"
      :pt="{
        root: { class: 'border-none p-0 overflow-hidden' },
        content: { class: 'p-0 h-full overflow-hidden' }
      }"
    >
      <EnvEditor
        v-if="activeSecretDetail"
        resource-type="secret"
        :detail="activeSecretDetail"
        :read-only="!canMutate"
        @close="isEditorOpen = false"
        @saved="() => k8sStore.fetchSecrets()"
      />
    </Dialog>

    <!-- Create Secret Dialog -->
    <Dialog
      v-model:visible="isCreateOpen"
      modal
      :show-header="false"
      class="w-[92vw] max-w-2xl max-h-[88vh] rounded-2xl overflow-hidden shadow-2xl border border-slate-200 dark:border-slate-800"
      :pt="{
        root: { class: 'border-none p-0 overflow-hidden' },
        content: { class: 'p-0 h-full flex flex-col overflow-hidden bg-white dark:bg-slate-900 text-slate-800 dark:text-slate-100' }
      }"
    >
      <!-- Custom Header -->
      <div
        class="px-6 py-4 bg-slate-50 dark:bg-slate-900/90 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between shrink-0"
      >
        <div class="flex items-center gap-3">
          <div
            class="w-9 h-9 rounded-xl bg-amber-500/10 text-amber-500 flex items-center justify-center font-bold text-base shrink-0"
          >
            <i class="pi pi-key"></i>
          </div>
          <div>
            <h2 class="font-bold text-base text-slate-900 dark:text-slate-100">
              Create Kubernetes Secret
            </h2>
            <p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
              Configure credentials, TLS certs, or environment variables visually without manual YAML.
            </p>
          </div>
        </div>
        <button
          type="button"
          class="w-8 h-8 rounded-lg flex items-center justify-center text-slate-400 hover:text-slate-700 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors cursor-pointer"
          @click="isCreateOpen = false"
        >
          <i class="pi pi-times"></i>
        </button>
      </div>

      <form class="flex-1 overflow-y-auto p-6 space-y-5" @submit.prevent="createSecret">
        <!-- Secret Name & Namespace -->
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div>
            <label
              class="block text-xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1.5"
            >
              Secret Name *
            </label>
            <InputText
              v-model="newSecretName"
              placeholder="e.g. app-credentials or tls-cert"
              class="w-full font-mono text-xs py-2"
              required
            />
          </div>

          <div>
            <label
              class="block text-xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1.5"
            >
              Namespace
            </label>
            <div
              class="w-full font-mono text-xs py-2 px-3 bg-slate-100 dark:bg-slate-800 rounded-lg text-slate-600 dark:text-slate-300 border border-slate-200 dark:border-slate-700"
            >
              {{ selectedNamespace }}
            </div>
          </div>
        </div>

        <!-- Secret Type Selector -->
        <div>
          <label
            class="block text-xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1.5"
          >
            Secret Type *
          </label>
          <select
            v-model="newSecretType"
            class="w-full p-2.5 text-xs font-mono bg-white dark:bg-slate-950 border border-slate-300 dark:border-slate-700 rounded-lg text-slate-800 dark:text-slate-200 focus:outline-none"
          >
            <option value="Opaque">Opaque (Generic Environment Variables / Key-Values)</option>
            <option value="kubernetes.io/tls">kubernetes.io/tls (TLS Certificate & Private Key)</option>
            <option value="kubernetes.io/dockerconfigjson">kubernetes.io/dockerconfigjson (Container Registry imagePullSecret)</option>
            <option value="kubernetes.io/basic-auth">kubernetes.io/basic-auth (HTTP Basic Authentication)</option>
            <option value="kubernetes.io/ssh-auth">kubernetes.io/ssh-auth (SSH Keypair Authentication)</option>
          </select>
        </div>

        <!-- TYPE 1: Opaque -->
        <div v-if="newSecretType === 'Opaque'" class="space-y-3 pt-1">
          <div class="flex items-center justify-between">
            <span class="text-xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider">
              Secret Keys & Values
            </span>

            <div class="flex items-center gap-2">
              <Button
                :label="opaqueMode === 'keyvalue' ? 'Switch to .env Text' : 'Switch to Key-Value Form'"
                size="small"
                outlined
                severity="secondary"
                class="text-xs py-1"
                @click="opaqueMode = opaqueMode === 'keyvalue' ? 'env' : 'keyvalue'"
              />
              <Button
                v-if="opaqueMode === 'keyvalue'"
                label="Add Row"
                icon="pi pi-plus"
                size="small"
                class="btn-sky text-xs py-1"
                @click="addOpaqueRow"
              />
            </div>
          </div>

          <!-- Key-Value Table -->
          <div v-if="opaqueMode === 'keyvalue'" class="space-y-2">
            <div
              v-for="(row, idx) in opaqueKeyValues"
              :key="idx"
              class="flex items-center gap-2 p-2 rounded-lg bg-slate-50 dark:bg-slate-950 border border-slate-200 dark:border-slate-800"
            >
              <InputText
                v-model="row.key"
                placeholder="KEY_NAME"
                class="w-1/3 font-mono text-xs py-1.5"
              />
              <span class="text-slate-400 font-mono">=</span>
              <div class="flex-1 relative flex items-center">
                <InputText
                  v-model="row.value"
                  :type="row.hidden ? 'password' : 'text'"
                  placeholder="secret value"
                  class="w-full font-mono text-xs py-1.5 pr-8"
                />
                <button
                  type="button"
                  class="absolute right-2 text-slate-400 hover:text-slate-200 cursor-pointer"
                  @click="row.hidden = !row.hidden"
                >
                  <i :class="row.hidden ? 'pi pi-eye' : 'pi pi-eye-slash'" class="text-xs"></i>
                </button>
              </div>
              <Button
                icon="pi pi-trash"
                severity="danger"
                text
                rounded
                size="small"
                class="text-slate-400 hover:text-rose-500"
                @click="removeOpaqueRow(idx)"
              />
            </div>
          </div>

          <!-- .env textarea -->
          <div v-else>
            <textarea
              v-model="newSecretEnvText"
              rows="6"
              placeholder="DATABASE_PASSWORD=super_secret&#10;API_TOKEN=xyz123&#10;ENCRYPTION_KEY=..."
              class="w-full p-3 font-mono text-xs bg-slate-950 text-emerald-400 rounded-lg border border-slate-800 focus:outline-none resize-none"
            ></textarea>
            <p class="text-[11px] text-slate-400 mt-1">
              Supports standard <code>KEY=value</code> pairs separated by newlines. Comments starting with <code>#</code> will be ignored.
            </p>
          </div>
        </div>

        <!-- TYPE 2: kubernetes.io/tls -->
        <div v-else-if="newSecretType === 'kubernetes.io/tls'" class="space-y-4 pt-1">
          <div class="p-3 rounded-lg bg-emerald-500/10 border border-emerald-500/20 text-xs text-emerald-400 flex items-center gap-2">
            <i class="pi pi-shield text-base"></i>
            <span>TLS certificates will be saved as <strong>tls.crt</strong> and <strong>tls.key</strong> compatible with Ingress and cert-manager.</span>
          </div>

          <div>
            <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1">
              Certificate (PEM format - tls.crt) *
            </label>
            <textarea
              v-model="tlsCert"
              rows="4"
              placeholder="-----BEGIN CERTIFICATE-----&#10;MIIDrzCCApegAwIBAgIQ...&#10;-----END CERTIFICATE-----"
              class="w-full p-2.5 font-mono text-xs bg-slate-950 text-emerald-400 rounded-lg border border-slate-800 focus:outline-none resize-none"
              required
            ></textarea>
          </div>

          <div>
            <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1">
              Private Key (PEM format - tls.key) *
            </label>
            <textarea
              v-model="tlsKey"
              rows="4"
              placeholder="-----BEGIN RSA PRIVATE KEY-----&#10;MIIEowIBAAKCAQEA0...&#10;-----END RSA PRIVATE KEY-----"
              class="w-full p-2.5 font-mono text-xs bg-slate-950 text-amber-400 rounded-lg border border-slate-800 focus:outline-none resize-none"
              required
            ></textarea>
          </div>
        </div>

        <!-- TYPE 3: kubernetes.io/dockerconfigjson -->
        <div v-else-if="newSecretType === 'kubernetes.io/dockerconfigjson'" class="space-y-4 pt-1">
          <div class="p-3 rounded-lg bg-purple-500/10 border border-purple-500/20 text-xs text-purple-400 flex items-center gap-2">
            <i class="pi pi-box text-base"></i>
            <span>Image pull credentials will be formatted into <strong>.dockerconfigjson</strong> for private registry pulling.</span>
          </div>

          <!-- Presets -->
          <div>
            <label class="block text-[11px] font-semibold text-slate-400 uppercase tracking-wider mb-1.5">
              Popular Registry Presets:
            </label>
            <div class="flex flex-wrap gap-2">
              <button
                v-for="p in dockerPresets"
                :key="p.server"
                type="button"
                class="px-2.5 py-1 rounded-md text-xs font-mono border transition cursor-pointer"
                :class="
                  dockerServer === p.server
                    ? 'bg-purple-500/20 text-purple-300 border-purple-500/40'
                    : 'bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400 border-slate-300 dark:border-slate-700'
                "
                @click="dockerServer = p.server"
              >
                {{ p.label }}
              </button>
            </div>
          </div>

          <div>
            <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1">
              Registry Server *
            </label>
            <InputText
              v-model="dockerServer"
              placeholder="https://index.docker.io/v1/ or ghcr.io"
              class="w-full font-mono text-xs py-2"
              required
            />
          </div>

          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1">
                Username *
              </label>
              <InputText
                v-model="dockerUsername"
                placeholder="e.g. eka-dev or _json_key"
                class="w-full font-mono text-xs py-2"
                required
              />
            </div>

            <div>
              <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1">
                Password / Access Token *
              </label>
              <div class="relative flex items-center">
                <InputText
                  v-model="dockerPassword"
                  :type="dockerPasswordHidden ? 'password' : 'text'"
                  placeholder="token or password"
                  class="w-full font-mono text-xs py-2 pr-8"
                  required
                />
                <button
                  type="button"
                  class="absolute right-2 text-slate-400 hover:text-slate-200 cursor-pointer"
                  @click="dockerPasswordHidden = !dockerPasswordHidden"
                >
                  <i :class="dockerPasswordHidden ? 'pi pi-eye' : 'pi pi-eye-slash'" class="text-xs"></i>
                </button>
              </div>
            </div>
          </div>

          <div>
            <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1">
              Email (Optional)
            </label>
            <InputText
              v-model="dockerEmail"
              placeholder="e.g. user@example.com"
              class="w-full font-mono text-xs py-2"
            />
          </div>
        </div>

        <!-- TYPE 4: kubernetes.io/basic-auth -->
        <div v-else-if="newSecretType === 'kubernetes.io/basic-auth'" class="space-y-4 pt-1">
          <div class="p-3 rounded-lg bg-amber-500/10 border border-amber-500/20 text-xs text-amber-400 flex items-center gap-2">
            <i class="pi pi-user text-base"></i>
            <span>HTTP Basic credentials stored as <strong>username</strong> and <strong>password</strong>.</span>
          </div>

          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1">
                Username *
              </label>
              <InputText
                v-model="basicAuthUsername"
                placeholder="admin"
                class="w-full font-mono text-xs py-2"
                required
              />
            </div>

            <div>
              <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1">
                Password *
              </label>
              <div class="relative flex items-center">
                <InputText
                  v-model="basicAuthPassword"
                  :type="basicAuthPasswordHidden ? 'password' : 'text'"
                  placeholder="password"
                  class="w-full font-mono text-xs py-2 pr-8"
                  required
                />
                <button
                  type="button"
                  class="absolute right-2 text-slate-400 hover:text-slate-200 cursor-pointer"
                  @click="basicAuthPasswordHidden = !basicAuthPasswordHidden"
                >
                  <i :class="basicAuthPasswordHidden ? 'pi pi-eye' : 'pi pi-eye-slash'" class="text-xs"></i>
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- TYPE 5: kubernetes.io/ssh-auth -->
        <div v-else-if="newSecretType === 'kubernetes.io/ssh-auth'" class="space-y-4 pt-1">
          <div class="p-3 rounded-lg bg-rose-500/10 border border-rose-500/20 text-xs text-rose-400 flex items-center gap-2">
            <i class="pi pi-key text-base"></i>
            <span>SSH credentials for git-sync and remote access stored as <strong>ssh-privatekey</strong>.</span>
          </div>

          <div>
            <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1">
              SSH Private Key (ssh-privatekey) *
            </label>
            <textarea
              v-model="sshPrivateKey"
              rows="4"
              placeholder="-----BEGIN OPENSSH PRIVATE KEY-----&#10;b3BlbnNzaC1rZXktdjEAAAA...&#10;-----END OPENSSH PRIVATE KEY-----"
              class="w-full p-2.5 font-mono text-xs bg-slate-950 text-rose-400 rounded-lg border border-slate-800 focus:outline-none resize-none"
              required
            ></textarea>
          </div>

          <div>
            <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1">
              SSH Public Key (ssh-publickey, optional)
            </label>
            <textarea
              v-model="sshPublicKey"
              rows="2"
              placeholder="ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI..."
              class="w-full p-2.5 font-mono text-xs bg-slate-950 text-slate-300 rounded-lg border border-slate-800 focus:outline-none resize-none"
            ></textarea>
          </div>

          <div>
            <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1">
              Known Hosts (known_hosts, optional)
            </label>
            <textarea
              v-model="sshKnownHosts"
              rows="2"
              placeholder="github.com ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl"
              class="w-full p-2.5 font-mono text-xs bg-slate-950 text-slate-400 rounded-lg border border-slate-800 focus:outline-none resize-none"
            ></textarea>
          </div>
        </div>

        <div class="flex justify-end gap-2 pt-3 border-t border-slate-200 dark:border-slate-800">
          <Button label="Cancel" severity="secondary" text size="small" @click="isCreateOpen = false" />
          <Button
            type="submit"
            label="Create Secret"
            icon="pi pi-check"
            class="btn-emerald text-xs font-semibold shadow-xs cursor-pointer px-4 py-2"
            :loading="isActionLoading"
          />
        </div>
      </form>
    </Dialog>
  </div>
</template>

<style scoped></style>

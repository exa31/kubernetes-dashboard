<script setup lang="ts">
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import Select from 'primevue/select'
import Tag from 'primevue/tag'
import { computed, ref, watch } from 'vue'

import type { ApplyYAMLResult } from '@/api/k8s'
import { useK8sStore } from '@/stores'

const props = defineProps<{
  visible: boolean
}>()

const emit = defineEmits<{
  (e: 'update:visible', val: boolean): void
  (e: 'applied'): void
}>()

const k8sStore = useK8sStore()

const activeNamespace = computed(() => k8sStore.selectedNamespace || 'default')
const availableNamespaces = computed(() =>
  k8sStore.namespaces.map((ns) => ({ label: ns.name, value: ns.name })),
)

const targetNamespace = ref(activeNamespace.value)
const yamlContent = ref('')
const isApplying = ref(false)
const isDryRunning = ref(false)
const executionResult = ref<ApplyYAMLResult | null>(null)
const executionError = ref<string | null>(null)
const selectedTemplateKey = ref<string>('custom')
const fileInputRef = ref<HTMLInputElement | null>(null)

// Watch for namespace change
watch(
  () => k8sStore.selectedNamespace,
  (newNs) => {
    if (newNs) targetNamespace.value = newNs
  },
)

// Watch visible to reset results
watch(
  () => props.visible,
  (vis) => {
    if (vis) {
      targetNamespace.value = activeNamespace.value
      executionResult.value = null
      executionError.value = null
      if (!yamlContent.value.trim()) {
        loadTemplate('deployment_service')
      }
    }
  },
)

const TEMPLATES: Record<string, { label: string; icon: string; yaml: (ns: string) => string }> = {
  custom: {
    label: 'Custom (Blank)',
    icon: 'pi pi-file',
    yaml: () => `# Enter your Kubernetes YAML manifest below.
# Multi-document manifests are supported using '---'.
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-sample-config
  namespace: default
data:
  APP_ENV: "production"
`,
  },
  deployment_service: {
    label: 'Deployment & Service (Full Web Stack)',
    icon: 'pi pi-server',
    yaml: (ns) => `apiVersion: apps/v1
kind: Deployment
metadata:
  name: demo-web-app
  namespace: ${ns}
  labels:
    app: demo-web-app
spec:
  replicas: 2
  selector:
    matchLabels:
      app: demo-web-app
  template:
    metadata:
      labels:
        app: demo-web-app
    spec:
      containers:
        - name: web
          image: nginx:alpine
          ports:
            - containerPort: 80
          resources:
            requests:
              cpu: 50m
              memory: 64Mi
            limits:
              cpu: 200m
              memory: 256Mi
---
apiVersion: v1
kind: Service
metadata:
  name: demo-web-service
  namespace: ${ns}
  labels:
    app: demo-web-app
spec:
  type: ClusterIP
  ports:
    - port: 80
      targetPort: 80
      protocol: TCP
  selector:
    app: demo-web-app
`,
  },
  cronjob: {
    label: 'CronJob (Scheduled Task)',
    icon: 'pi pi-clock',
    yaml: (ns) => `apiVersion: batch/v1
kind: CronJob
metadata:
  name: demo-scheduled-task
  namespace: ${ns}
spec:
  schedule: "0 1 * * *"
  successfulJobsHistoryLimit: 3
  failedJobsHistoryLimit: 1
  jobTemplate:
    spec:
      template:
        spec:
          containers:
            - name: task-runner
              image: busybox:latest
              command:
                - /bin/sh
                - -c
                - "echo 'Running scheduled batch task...'; date; sleep 3; echo 'Finished!'"
          restartPolicy: OnFailure
`,
  },
  configmap: {
    label: 'ConfigMap (App Configuration)',
    icon: 'pi pi-file-edit',
    yaml: (ns) => `apiVersion: v1
kind: ConfigMap
metadata:
  name: app-runtime-config
  namespace: ${ns}
data:
  APP_ENV: "production"
  LOG_LEVEL: "info"
  PORT: "8080"
  ENABLE_METRICS: "true"
`,
  },
  secret: {
    label: 'Secret (Opaque Credentials)',
    icon: 'pi pi-key',
    yaml: (ns) => `apiVersion: v1
kind: Secret
metadata:
  name: app-credentials
  namespace: ${ns}
type: Opaque
stringData:
  API_KEY: "super-secret-production-token-12345"
  DB_PASSWORD: "database-secure-password-abc"
`,
  },
  ingress: {
    label: 'Ingress (HTTP Route & Domain)',
    icon: 'pi pi-globe',
    yaml: (ns) => `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: demo-ingress
  namespace: ${ns}
  annotations:
    kubernetes.io/ingress.class: "nginx"
spec:
  rules:
    - host: app.kubeenv.local
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: demo-web-service
                port:
                  number: 80
`,
  },
  pvc: {
    label: 'PersistentVolumeClaim (Storage)',
    icon: 'pi pi-database',
    yaml: (ns) => `apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: app-data-pvc
  namespace: ${ns}
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 2Gi
`,
  },
  pod: {
    label: 'Pod (Single Container)',
    icon: 'pi pi-box',
    yaml: (ns) => `apiVersion: v1
kind: Pod
metadata:
  name: toolkit-pod
  namespace: ${ns}
spec:
  containers:
    - name: tools
      image: curlimages/curl:latest
      command: ["sleep", "3600"]
`,
  },
}

const templateOptions = Object.entries(TEMPLATES).map(([key, item]) => ({
  key,
  label: item.label,
  icon: item.icon,
}))

function loadTemplate(key: string) {
  selectedTemplateKey.value = key
  const t = TEMPLATES[key]
  if (t) {
    yamlContent.value = t.yaml(targetNamespace.value)
    executionResult.value = null
    executionError.value = null
  }
}

// Line numbers computation
const lineNumbersRef = ref<HTMLElement | null>(null)
const textareaRef = ref<HTMLTextAreaElement | null>(null)

const lineCount = computed(() => {
  return Math.max(yamlContent.value.split('\n').length, 1)
})

function handleScroll() {
  if (textareaRef.value && lineNumbersRef.value) {
    lineNumbersRef.value.scrollTop = textareaRef.value.scrollTop
  }
}

// Handle Tab key in textarea to insert 2 spaces
function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'Tab') {
    e.preventDefault()
    const target = e.target as HTMLTextAreaElement
    const start = target.selectionStart
    const end = target.selectionEnd
    yamlContent.value =
      yamlContent.value.substring(0, start) + '  ' + yamlContent.value.substring(end)
    setTimeout(() => {
      target.selectionStart = target.selectionEnd = start + 2
    }, 0)
  }
}

// File Upload Handler
function handleFileUpload(e: Event) {
  const target = e.target as HTMLInputElement
  if (!target.files || target.files.length === 0) return
  const file = target.files[0]
  const reader = new FileReader()
  reader.onload = (event) => {
    if (event.target?.result) {
      yamlContent.value = event.target.result as string
      executionResult.value = null
      executionError.value = null
    }
  }
  reader.readAsText(file)
  target.value = ''
}

function triggerFileUpload() {
  fileInputRef.value?.click()
}

function clearEditor() {
  yamlContent.value = ''
  executionResult.value = null
  executionError.value = null
  selectedTemplateKey.value = 'custom'
}

// Execute Apply / Dry Run
async function executeYAML(dryRun: boolean) {
  if (!yamlContent.value.trim()) {
    executionError.value = 'Please provide YAML manifest content before executing.'
    return
  }

  if (dryRun) {
    isDryRunning.value = true
  } else {
    isApplying.value = true
  }
  executionError.value = null
  executionResult.value = null

  try {
    const res = await k8sStore.applyYAML(yamlContent.value, targetNamespace.value, dryRun)
    executionResult.value = res
    if (!dryRun) {
      emit('applied')
    }
  } catch (err: any) {
    executionError.value = err.message || 'Failed to apply YAML manifest to Kubernetes'
  } finally {
    isDryRunning.value = false
    isApplying.value = false
  }
}

function closeDialog() {
  emit('update:visible', false)
}
</script>

<template>
  <Dialog
    :visible="visible"
    modal
    dismissable-mask
    :show-header="false"
    class="w-full max-w-5xl rounded-2xl overflow-hidden shadow-2xl bg-white dark:bg-slate-950 border border-slate-200 dark:border-slate-800"
    :pt="{
      root: { class: 'border-none p-0 overflow-hidden' },
      header: { class: 'hidden' },
      content: { class: 'p-0 flex flex-col max-h-[92vh] overflow-hidden' }
    }"
    @update:visible="closeDialog"
  >
    <!-- Custom Header -->
    <div class="px-6 py-4 bg-white dark:bg-slate-900 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between shrink-0">
      <div class="flex items-center gap-3">
        <div class="w-10 h-10 rounded-xl bg-sky-500/10 text-sky-500 flex items-center justify-center font-bold text-lg shrink-0">
          <i class="pi pi-code"></i>
        </div>
        <div>
          <h2 class="font-bold text-lg text-slate-900 dark:text-slate-100 flex items-center gap-2">
            <span>Apply Kubernetes YAML Manifest</span>
            <span class="px-2 py-0.5 rounded text-[11px] font-mono bg-sky-100 dark:bg-sky-950/60 text-sky-700 dark:text-sky-300 font-semibold border border-sky-200 dark:border-sky-800">
              kubectl apply -f
            </span>
          </h2>
          <p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
            Directly create, update, or validate multi-document Kubernetes resources in your cluster.
          </p>
        </div>
      </div>

      <button
        type="button"
        class="w-8 h-8 rounded-lg flex items-center justify-center text-slate-400 hover:text-slate-700 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors cursor-pointer"
        @click="closeDialog"
      >
        <i class="pi pi-times"></i>
      </button>
    </div>

    <!-- Controls Toolbar -->
    <div class="px-6 py-3 bg-slate-50/80 dark:bg-slate-900/60 border-b border-slate-200 dark:border-slate-800 flex flex-wrap items-center justify-between gap-3 shrink-0">
      <div class="flex flex-wrap items-center gap-3">
        <!-- Preset Templates Selector -->
        <div class="flex items-center gap-2">
          <label class="text-xs font-semibold text-slate-600 dark:text-slate-300 uppercase tracking-wider">
            Template:
          </label>
          <Select
            v-model="selectedTemplateKey"
            :options="templateOptions"
            option-label="label"
            option-value="key"
            placeholder="Select a template"
            class="text-xs w-64 h-9"
            @change="loadTemplate(selectedTemplateKey)"
          >
            <template #value="slotProps">
              <div v-if="slotProps.value" class="flex items-center gap-2 text-xs">
                <i :class="TEMPLATES[slotProps.value]?.icon || 'pi pi-file'" class="text-sky-500"></i>
                <span class="truncate">{{ TEMPLATES[slotProps.value]?.label }}</span>
              </div>
            </template>
            <template #option="slotProps">
              <div class="flex items-center gap-2 text-xs py-0.5">
                <i :class="slotProps.option.icon" class="text-sky-500"></i>
                <span>{{ slotProps.option.label }}</span>
              </div>
            </template>
          </Select>
        </div>

        <!-- Target Namespace -->
        <div class="flex items-center gap-2">
          <label class="text-xs font-semibold text-slate-600 dark:text-slate-300 uppercase tracking-wider">
            Target Namespace:
          </label>
          <Select
            v-model="targetNamespace"
            :options="availableNamespaces"
            option-label="label"
            option-value="value"
            placeholder="Namespace"
            class="text-xs w-40 h-9 font-mono"
          />
        </div>
      </div>

      <!-- Action buttons on toolbar -->
      <div class="flex items-center gap-2">
        <input
          ref="fileInputRef"
          type="file"
          accept=".yaml,.yml"
          class="hidden"
          @change="handleFileUpload"
        />
        <Button
          label="Upload File"
          icon="pi pi-upload"
          size="small"
          severity="secondary"
          outlined
          class="text-xs px-2.5 py-1.5 cursor-pointer"
          title="Upload a .yaml or .yml file from your machine"
          @click="triggerFileUpload"
        />
        <Button
          label="Clear"
          icon="pi pi-trash"
          size="small"
          severity="secondary"
          text
          class="text-xs px-2.5 py-1.5 cursor-pointer text-slate-500 hover:text-rose-500"
          title="Clear editor contents"
          @click="clearEditor"
        />
      </div>
    </div>

    <!-- Editor Body -->
    <div class="flex-1 flex flex-col min-h-0 bg-slate-950 overflow-hidden">
      <!-- Editor Code Box with synchronized line numbers -->
      <div class="relative flex w-full h-[450px] min-h-[380px] bg-slate-950 overflow-hidden">
        <!-- Line Numbers Strip -->
        <div
          ref="lineNumbersRef"
          class="w-12 py-3 pr-3 text-right font-mono text-xs select-none bg-slate-900/90 text-slate-500 border-r border-slate-800/80 shrink-0 overflow-hidden pointer-events-none"
        >
          <div v-for="n in lineCount" :key="n" class="leading-6 h-6">
            {{ n }}
          </div>
        </div>

        <!-- Textarea -->
        <textarea
          ref="textareaRef"
          v-model="yamlContent"
          placeholder="Paste or write Kubernetes manifest YAML here..."
          spellcheck="false"
          wrap="off"
          class="flex-1 w-full h-full py-3 px-3.5 font-mono text-xs leading-6 bg-transparent text-slate-100 placeholder-slate-600 focus:outline-none resize-none overflow-auto whitespace-pre selection:bg-sky-600 selection:text-white"
          @scroll="handleScroll"
          @keydown="handleKeydown"
        ></textarea>
      </div>

      <!-- Execution Feedback Drawer / Result Banner -->
      <div
        v-if="executionResult || executionError"
        class="border-t border-slate-800 max-h-48 overflow-y-auto p-4 shrink-0 transition-all"
        :class="executionError || executionResult?.error_count ? 'bg-rose-950/40' : 'bg-emerald-950/40'"
      >
        <!-- Error Banner -->
        <div v-if="executionError" class="flex items-start gap-2.5 text-rose-300 text-xs">
          <i class="pi pi-exclamation-triangle text-rose-400 mt-0.5 text-sm shrink-0"></i>
          <div>
            <div class="font-bold">Execution Failed</div>
            <div class="font-mono mt-0.5">{{ executionError }}</div>
          </div>
        </div>

        <!-- Structured Results -->
        <div v-else-if="executionResult" class="space-y-2">
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-2">
              <i
                :class="executionResult.error_count > 0 ? 'pi pi-exclamation-circle text-amber-400' : 'pi pi-check-circle text-emerald-400'"
                class="text-base"
              ></i>
              <span class="text-xs font-bold text-slate-200">
                {{ executionResult.dry_run ? 'Dry-Run Pre-Flight Validation:' : 'Apply Execution Result:' }}
              </span>
              <Tag
                :value="`${executionResult.success_count} / ${executionResult.total} Succeeded`"
                :severity="executionResult.error_count > 0 ? 'warn' : 'success'"
                class="text-[10px]"
              />
            </div>
            <span v-if="executionResult.dry_run" class="text-[11px] font-mono text-sky-400 italic">
              No cluster state was altered (Dry Run)
            </span>
          </div>

          <!-- Applied Resources List -->
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-2 mt-2">
            <div
              v-for="(item, idx) in executionResult.results"
              :key="idx"
              class="p-2 rounded-lg border text-xs font-mono flex items-center justify-between"
              :class="
                item.status === 'success'
                  ? 'bg-emerald-900/20 border-emerald-800/50 text-emerald-300'
                  : 'bg-rose-900/20 border-rose-800/50 text-rose-300'
              "
            >
              <div class="truncate mr-2">
                <span class="font-bold">{{ item.kind }}/{{ item.name }}</span>
                <span v-if="item.namespace" class="text-slate-400 ml-1 text-[11px]">({{ item.namespace }})</span>
              </div>
              <div class="flex items-center gap-1.5 shrink-0">
                <Tag
                  :value="item.action || item.status"
                  :severity="item.status === 'success' ? 'success' : 'danger'"
                  class="text-[10px] uppercase"
                />
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Footer Controls -->
    <div class="px-6 py-4 bg-white dark:bg-slate-900 border-t border-slate-200 dark:border-slate-800 flex flex-wrap items-center justify-between gap-3 shrink-0">
      <div class="text-xs text-slate-400 font-mono flex items-center gap-2">
        <i class="pi pi-info-circle text-sky-500"></i>
        <span>Supports multiple documents separated by <code class="text-sky-400">---</code></span>
      </div>

      <div class="flex items-center gap-2">
        <Button
          label="Cancel"
          severity="secondary"
          text
          size="small"
          class="text-xs px-3 py-1.5 cursor-pointer"
          @click="closeDialog"
        />

        <!-- Dry Run Button -->
        <Button
          label="Dry Run (Validate)"
          icon="pi pi-shield"
          size="small"
          class="btn-sky text-xs px-3.5 py-1.5 rounded-lg active:scale-95 cursor-pointer"
          :loading="isDryRunning"
          :disabled="isApplying"
          title="Test validate syntax and cluster schemas without changing resources"
          @click="executeYAML(true)"
        />

        <!-- Apply Button -->
        <Button
          label="Apply Manifest"
          icon="pi pi-play"
          size="small"
          class="btn-emerald text-xs px-4 py-1.5 rounded-lg active:scale-95 cursor-pointer font-bold"
          :loading="isApplying"
          :disabled="isDryRunning"
          title="Apply manifest directly to Kubernetes cluster (kubectl apply -f)"
          @click="executeYAML(false)"
        />
      </div>
    </div>
  </Dialog>
</template>

<style scoped>
/* Scrollbar styling for code editor */
textarea::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}
textarea::-webkit-scrollbar-track {
  background: #090d16;
}
textarea::-webkit-scrollbar-thumb {
  background: #1e293b;
  border-radius: 4px;
}
textarea::-webkit-scrollbar-thumb:hover {
  background: #334155;
}
</style>

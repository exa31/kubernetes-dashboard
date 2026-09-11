<script setup lang="ts">
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import InputNumber from 'primevue/inputnumber'
import InputText from 'primevue/inputtext'
import Message from 'primevue/message'
import { computed, ref, watch } from 'vue'

import { useK8sStore } from '@/stores'
import type { ContainerEnvVar, CreateDeploymentPayload } from '@/types'
import { logger } from '@/utils'

const props = defineProps<{
  visible: boolean
  namespace?: string
}>()

const emit = defineEmits<{
  (e: 'update:visible', val: boolean): void
  (e: 'deployed', name: string): void
}>()

const k8sStore = useK8sStore()

// Tabs
type ActiveTab = 'container' | 'capacity' | 'scaling' | 'env' | 'yaml'
const activeTab = ref<ActiveTab>('container')

// Form State
const workloadName = ref('')
const selectedNs = ref(props.namespace || k8sStore.selectedNamespace || 'default')
const image = ref('')
const containerPort = ref<number | null>(8080)
const exposeService = ref(true)
const servicePort = ref<number | null>(80)
const serviceType = ref<'ClusterIP' | 'NodePort' | 'LoadBalancer'>('ClusterIP')

// Capacity / Tiers
type ResourceTier = 'micro' | 'standard' | 'performance' | 'high' | 'custom'
const selectedTier = ref<ResourceTier>('standard')
const cpuRequest = ref('250m')
const cpuLimit = ref('500m')
const memoryRequest = ref('256Mi')
const memoryLimit = ref('512Mi')

// Scaling
const scalingMode = ref<'keda' | 'hpa' | 'manual'>('keda')
const manualReplicas = ref(1)
const minReplicas = ref(0)
const maxReplicas = ref(3)
const targetConcurrency = ref(30)
const targetCpuPct = ref(75)

// Environment Variables
const envVars = ref<Array<{ name: string; value: string; secretRef?: string }>>([])
const envPasteMode = ref(false)
const envPasteText = ref('')

// Submission & Feedback
const isDeploying = ref(false)
const errorMessage = ref<string | null>(null)
const yamlCopied = ref(false)

// Presets for popular images
const imagePresets = [
  { label: 'Nginx (Alpine)', image: 'nginx:alpine', port: 80, name: 'nginx-web' },
  { label: 'Node.js (Alpine)', image: 'node:20-alpine', port: 3000, name: 'node-app' },
  { label: 'Python FastAPI', image: 'tiangolo/uvicorn-gunicorn-fastapi:python3.10-alpine', port: 80, name: 'fastapi-app' },
  { label: 'Go Echo / Fiber', image: 'golang:1.24-alpine', port: 8080, name: 'go-service' },
  { label: 'Redis (Alpine)', image: 'redis:7-alpine', port: 6379, name: 'redis-cache' },
  { label: 'PostgreSQL 16', image: 'postgres:16-alpine', port: 5432, name: 'postgres-db' }
]

function applyPreset(preset: typeof imagePresets[0]) {
  image.value = preset.image
  containerPort.value = preset.port
  servicePort.value = preset.port === 8080 ? 80 : preset.port
  if (!workloadName.value) {
    workloadName.value = preset.name
  }
}

function applyTier(tier: ResourceTier) {
  selectedTier.value = tier
  switch (tier) {
    case 'micro':
      cpuRequest.value = '100m'
      cpuLimit.value = '250m'
      memoryRequest.value = '128Mi'
      memoryLimit.value = '256Mi'
      break
    case 'standard':
      cpuRequest.value = '250m'
      cpuLimit.value = '500m'
      memoryRequest.value = '256Mi'
      memoryLimit.value = '512Mi'
      break
    case 'performance':
      cpuRequest.value = '500m'
      cpuLimit.value = '1000m'
      memoryRequest.value = '512Mi'
      memoryLimit.value = '1Gi'
      break
    case 'high':
      cpuRequest.value = '1000m'
      cpuLimit.value = '2000m'
      memoryRequest.value = '1Gi'
      memoryLimit.value = '2Gi'
      break
  }
}

// Env helpers
function addEnvVar() {
  envVars.value.push({ name: '', value: '' })
}

function removeEnvVar(idx: number) {
  envVars.value.splice(idx, 1)
}

function importPastedEnv() {
  const lines = envPasteText.value.split('\n')
  for (const line of lines) {
    const trimmed = line.trim()
    if (!trimmed || trimmed.startsWith('#')) continue
    const eq = trimmed.indexOf('=')
    if (eq > 0) {
      const k = trimmed.slice(0, eq).trim()
      let v = trimmed.slice(eq + 1).trim()
      if ((v.startsWith('"') && v.endsWith('"')) || (v.startsWith("'") && v.endsWith("'"))) {
        v = v.slice(1, -1)
      }
      envVars.value.push({ name: k, value: v })
    }
  }
  envPasteText.value = ''
  envPasteMode.value = false
}

// Live YAML generation
const generatedYaml = computed(() => {
  const name = workloadName.value.trim() || 'my-workload'
  const ns = selectedNs.value || 'default'
  const img = image.value.trim() || 'nginx:alpine'
  const reps = scalingMode.value === 'manual' ? manualReplicas.value : minReplicas.value

  let yaml = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: ${name}
  namespace: ${ns}
  labels:
    app: ${name}
    app.kubernetes.io/name: ${name}
    app.kubernetes.io/managed-by: kubenexus
  annotations:
    kubenexus.io/deployment-style: cloud-run
spec:
  replicas: ${reps}
  selector:
    matchLabels:
      app: ${name}
  template:
    metadata:
      labels:
        app: ${name}
    spec:
      containers:
      - name: ${name}
        image: ${img}
`
  if (containerPort.value) {
    yaml += `        ports:
        - containerPort: ${containerPort.value}
          name: http
          protocol: TCP
`
  }

  if (cpuRequest.value || memoryRequest.value || cpuLimit.value || memoryLimit.value) {
    yaml += `        resources:
          requests:
            cpu: "${cpuRequest.value}"
            memory: "${memoryRequest.value}"
          limits:
            cpu: "${cpuLimit.value}"
            memory: "${memoryLimit.value}"
`
  }

  const validEnvs = envVars.value.filter((e) => e.name.trim() !== '')
  if (validEnvs.length > 0) {
    yaml += `        env:
`
    for (const e of validEnvs) {
      if (e.secretRef) {
        const [sName, sKey] = e.secretRef.split(':')
        yaml += `        - name: ${e.name.trim()}
          valueFrom:
            secretKeyRef:
              name: ${sName}
              key: ${sKey || 'key'}
`
      } else {
        yaml += `        - name: ${e.name.trim()}
          value: "${e.value}"
`
      }
    }
  }

  if (exposeService.value && containerPort.value) {
    yaml += `---
apiVersion: v1
kind: Service
metadata:
  name: ${name}
  namespace: ${ns}
  labels:
    app: ${name}
spec:
  type: ${serviceType.value}
  selector:
    app: ${name}
  ports:
  - name: http
    port: ${servicePort.value || containerPort.value}
    targetPort: ${containerPort.value}
    protocol: TCP
`
  }

  if (scalingMode.value === 'hpa') {
    yaml += `---
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: ${name}-hpa
  namespace: ${ns}
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: ${name}
  minReplicas: ${minReplicas.value}
  maxReplicas: ${maxReplicas.value}
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: ${targetCpuPct.value}
`
  }

  return yaml
})

async function copyYaml() {
  try {
    await navigator.clipboard.writeText(generatedYaml.value)
    yamlCopied.value = true
    setTimeout(() => {
      yamlCopied.value = false
    }, 2000)
  } catch (err) {
    logger.error('Failed to copy YAML', err)
  }
}

// Deploy Handler
async function deployWorkload() {
  if (!workloadName.value.trim()) {
    errorMessage.value = 'Workload name is required'
    activeTab.value = 'container'
    return
  }
  if (!image.value.trim()) {
    errorMessage.value = 'Container image is required'
    activeTab.value = 'container'
    return
  }

  isDeploying.value = true
  errorMessage.value = null

  const formattedEnvs: ContainerEnvVar[] = envVars.value
    .filter((e) => e.name.trim() !== '')
    .map((e) => ({
      name: e.name.trim(),
      value: e.value,
      secret_ref: e.secretRef
    }))

  const payload: CreateDeploymentPayload = {
    name: workloadName.value.trim().toLowerCase(),
    namespace: selectedNs.value,
    image: image.value.trim(),
    replicas:
      scalingMode.value === 'manual'
        ? manualReplicas.value
        : minReplicas.value > 0
        ? minReplicas.value
        : 1,
    port: containerPort.value ?? undefined,
    cpu_request: cpuRequest.value,
    cpu_limit: cpuLimit.value,
    memory_request: memoryRequest.value,
    memory_limit: memoryLimit.value,
    env: formattedEnvs.length > 0 ? formattedEnvs : undefined,
    create_service: exposeService.value,
    service_type: serviceType.value,
    service_port: servicePort.value ?? undefined
  }

  if (scalingMode.value === 'hpa') {
    payload.autoscale = {
      namespace: selectedNs.value,
      target_name: workloadName.value.trim().toLowerCase(),
      target_kind: 'Deployment',
      min_replicas: Math.max(1, minReplicas.value),
      max_replicas: maxReplicas.value,
      target_cpu: targetCpuPct.value
    }
  }

  try {
    await k8sStore.createDeployment(payload)

    // If KEDA mode selected, create HTTPScaledObject
    if (scalingMode.value === 'keda') {
      try {
        await k8sStore.saveKedaHTTPScaledObject({
          namespace: selectedNs.value,
          target_name: payload.name,
          target_kind: 'Deployment',
          target_service: exposeService.value ? payload.name : undefined,
          target_port: servicePort.value || containerPort.value || 80,
          min_replicas: minReplicas.value,
          max_replicas: maxReplicas.value,
          concurrency: targetConcurrency.value || 30,
          scaledown_period: 300
        })
      } catch (kedaErr) {
        logger.warn('Workload deployed, but KEDA autoscaling setup had warning:', kedaErr)
      }
    }

    emit('deployed', payload.name)
    emit('update:visible', false)
  } catch (err: unknown) {
    logger.error('Failed to deploy workload', err)
    let msg = 'Failed to deploy workload'
    if (err && typeof err === 'object' && 'response' in err) {
      const res = (err as { response?: { data?: { message?: string } } }).response
      if (res?.data?.message) msg = res.data.message
    } else if (err instanceof Error) {
      msg = err.message
    }
    errorMessage.value = msg
  } finally {
    isDeploying.value = false
  }
}

watch(
  () => props.visible,
  (open) => {
    if (open) {
      selectedNs.value = props.namespace || k8sStore.selectedNamespace || 'default'
      errorMessage.value = null
      activeTab.value = 'container'
      if (!workloadName.value) {
        applyTier('standard')
      }
    }
  }
)
</script>

<template>
  <Dialog
    :visible="visible"
    modal
    :show-header="false"
    class="w-[95vw] max-w-5xl h-[88vh] rounded-2xl overflow-hidden shadow-2xl border border-slate-200 dark:border-slate-800"
    :pt="{
      root: { class: 'border-none p-0 overflow-hidden' },
      content: { class: 'p-0 h-full flex flex-col overflow-hidden bg-slate-900 text-slate-100' }
    }"
  >
    <!-- Header: Cloud Run Banner -->
    <div
      class="px-6 py-4 bg-gradient-to-r from-slate-900 via-sky-950/40 to-slate-900 border-b border-slate-800 flex items-center justify-between shrink-0"
    >
      <div class="flex items-center gap-3.5">
        <div
          class="w-10 h-10 rounded-xl bg-gradient-to-tr from-sky-500 to-cyan-400 text-slate-950 flex items-center justify-center font-bold text-lg shadow-lg shadow-sky-500/20"
        >
          <i class="pi pi-cloud text-xl"></i>
        </div>
        <div>
          <div class="flex items-center gap-2">
            <h2 class="font-bold text-base text-slate-100 tracking-tight">Deploy Workload</h2>
            <span
              class="px-2 py-0.5 rounded-full text-[10px] font-semibold bg-sky-500/20 text-sky-300 border border-sky-500/30"
            >
              Cloud Run Style
            </span>
          </div>
          <p class="text-xs text-slate-400 mt-0.5">
            Deploy containers visually with automatic networking, resource limits, and autoscaling.
          </p>
        </div>
      </div>

      <div class="flex items-center gap-3">
        <!-- Namespace Indicator -->
        <div
          class="px-3 py-1.5 rounded-lg bg-slate-800/80 border border-slate-700/60 text-xs font-mono text-slate-300 flex items-center gap-2"
        >
          <i class="pi pi-box text-sky-400 text-xs"></i>
          <span>ns: <strong>{{ selectedNs }}</strong></span>
        </div>

        <Button
          icon="pi pi-times"
          severity="secondary"
          text
          rounded
          size="small"
          class="text-slate-400 hover:text-white"
          @click="emit('update:visible', false)"
        />
      </div>
    </div>

    <!-- Error Banner -->
    <div v-if="errorMessage" class="px-6 pt-4 shrink-0">
      <Message severity="error" :closable="false" class="text-xs">
        {{ errorMessage }}
      </Message>
    </div>

    <!-- Cloud Run Navigation Tabs -->
    <div
      class="px-6 border-b border-slate-800 bg-slate-950/70 flex items-center gap-2 shrink-0 overflow-x-auto"
    >
      <button
        type="button"
        class="py-3 px-3.5 text-xs font-medium border-b-2 transition flex items-center gap-2 cursor-pointer"
        :class="
          activeTab === 'container'
            ? 'border-sky-500 text-sky-400 font-semibold'
            : 'border-transparent text-slate-400 hover:text-slate-200'
        "
        @click="activeTab = 'container'"
      >
        <i class="pi pi-box"></i>
        <span>1. Container & Image</span>
      </button>

      <button
        type="button"
        class="py-3 px-3.5 text-xs font-medium border-b-2 transition flex items-center gap-2 cursor-pointer"
        :class="
          activeTab === 'capacity'
            ? 'border-sky-500 text-sky-400 font-semibold'
            : 'border-transparent text-slate-400 hover:text-slate-200'
        "
        @click="activeTab = 'capacity'"
      >
        <i class="pi pi-bolt"></i>
        <span>2. Capacity & Resources</span>
      </button>

      <button
        type="button"
        class="py-3 px-3.5 text-xs font-medium border-b-2 transition flex items-center gap-2 cursor-pointer"
        :class="
          activeTab === 'scaling'
            ? 'border-sky-500 text-sky-400 font-semibold'
            : 'border-transparent text-slate-400 hover:text-slate-200'
        "
        @click="activeTab = 'scaling'"
      >
        <i class="pi pi-sliders-h"></i>
        <span>3. Scaling & Autoscaling</span>
      </button>

      <button
        type="button"
        class="py-3 px-3.5 text-xs font-medium border-b-2 transition flex items-center gap-2 cursor-pointer"
        :class="
          activeTab === 'env'
            ? 'border-sky-500 text-sky-400 font-semibold'
            : 'border-transparent text-slate-400 hover:text-slate-200'
        "
        @click="activeTab = 'env'"
      >
        <i class="pi pi-lock"></i>
        <span>4. Environment & Secrets</span>
        <span
          v-if="envVars.filter((e) => e.name).length > 0"
          class="px-1.5 py-0.2 rounded-full bg-sky-500/20 text-sky-300 text-[10px]"
        >
          {{ envVars.filter((e) => e.name).length }}
        </span>
      </button>

      <button
        type="button"
        class="py-3 px-3.5 text-xs font-medium border-b-2 transition flex items-center gap-2 cursor-pointer ml-auto"
        :class="
          activeTab === 'yaml'
            ? 'border-purple-500 text-purple-400 font-semibold'
            : 'border-transparent text-slate-400 hover:text-slate-200'
        "
        @click="activeTab = 'yaml'"
      >
        <i class="pi pi-code"></i>
        <span>Live YAML Preview</span>
      </button>
    </div>

    <!-- Tab Content -->
    <div class="flex-1 overflow-y-auto p-6 bg-slate-950/40 space-y-6">
      <!-- TAB 1: Container & Image -->
      <div v-show="activeTab === 'container'" class="space-y-6 max-w-3xl">
        <!-- Workload Name & Presets -->
        <div class="p-5 rounded-xl bg-slate-900/80 border border-slate-800 space-y-4 shadow-sm">
          <h3 class="text-sm font-semibold text-slate-200 flex items-center gap-2">
            <i class="pi pi-id-card text-sky-400"></i>
            <span>Workload Identity</span>
          </h3>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label class="block text-xs font-semibold text-slate-400 uppercase tracking-wider mb-1.5">
                Workload Name *
              </label>
              <InputText
                v-model="workloadName"
                placeholder="e.g. auth-api, web-frontend"
                class="w-full font-mono text-xs py-2 bg-slate-950 border-slate-700 text-slate-100"
                required
              />
              <p class="text-[11px] text-slate-500 mt-1">
                Lowercase letters, numbers, and hyphens (DNS-1123).
              </p>
            </div>

            <div>
              <label class="block text-xs font-semibold text-slate-400 uppercase tracking-wider mb-1.5">
                Target Namespace
              </label>
              <InputText
                v-model="selectedNs"
                class="w-full font-mono text-xs py-2 bg-slate-950 border-slate-700 text-slate-100"
              />
            </div>
          </div>
        </div>

        <!-- Container Image & Presets -->
        <div class="p-5 rounded-xl bg-slate-900/80 border border-slate-800 space-y-4 shadow-sm">
          <div class="flex items-center justify-between">
            <h3 class="text-sm font-semibold text-slate-200 flex items-center gap-2">
              <i class="pi pi-box text-sky-400"></i>
              <span>Container Image *</span>
            </h3>
            <span class="text-xs text-slate-400">Docker Hub / GHCR / Artifact Registry</span>
          </div>

          <div>
            <InputText
              v-model="image"
              placeholder="e.g. nginx:alpine, ghcr.io/org/repo:tag, or gcr.io/..."
              class="w-full font-mono text-xs py-2.5 bg-slate-950 border-slate-700 text-slate-100"
              required
            />
          </div>

          <!-- Quick Presets -->
          <div>
            <label class="block text-[11px] font-semibold text-slate-400 uppercase tracking-wider mb-2">
              Popular Container Presets:
            </label>
            <div class="flex flex-wrap gap-2">
              <button
                v-for="p in imagePresets"
                :key="p.name"
                type="button"
                class="px-2.5 py-1.5 rounded-lg text-xs font-mono transition border cursor-pointer flex items-center gap-1.5"
                :class="
                  image === p.image
                    ? 'bg-sky-500/20 text-sky-300 border-sky-500/40'
                    : 'bg-slate-800/80 text-slate-300 border-slate-700/60 hover:border-slate-600 hover:text-white'
                "
                @click="applyPreset(p)"
              >
                <i class="pi pi-download text-[10px] text-sky-400"></i>
                <span>{{ p.label }}</span>
              </button>
            </div>
          </div>
        </div>

        <!-- Networking & Service Exposure -->
        <div class="p-5 rounded-xl bg-slate-900/80 border border-slate-800 space-y-4 shadow-sm">
          <div class="flex items-center justify-between">
            <h3 class="text-sm font-semibold text-slate-200 flex items-center gap-2">
              <i class="pi pi-globe text-cyan-400"></i>
              <span>Container Port & Service Networking</span>
            </h3>

            <label class="flex items-center gap-2 cursor-pointer text-xs font-semibold text-slate-300">
              <input
                v-model="exposeService"
                type="checkbox"
                class="rounded bg-slate-950 border-slate-700 text-sky-500 focus:ring-0 cursor-pointer"
              />
              <span>Auto-create Service</span>
            </label>
          </div>

          <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
            <div>
              <label class="block text-xs font-semibold text-slate-400 uppercase tracking-wider mb-1.5">
                Container Port
              </label>
              <InputNumber
                v-model="containerPort"
                placeholder="e.g. 8080"
                :min="1"
                :max="65535"
                class="w-full font-mono text-xs"
              />
              <p class="text-[11px] text-slate-500 mt-1">Port listened inside container.</p>
            </div>

            <div v-if="exposeService">
              <label class="block text-xs font-semibold text-slate-400 uppercase tracking-wider mb-1.5">
                Service Port
              </label>
              <InputNumber
                v-model="servicePort"
                placeholder="e.g. 80"
                :min="1"
                :max="65535"
                class="w-full font-mono text-xs"
              />
              <p class="text-[11px] text-slate-500 mt-1">Port exposed to other services.</p>
            </div>

            <div v-if="exposeService">
              <label class="block text-xs font-semibold text-slate-400 uppercase tracking-wider mb-1.5">
                Service Type
              </label>
              <select
                v-model="serviceType"
                class="w-full p-2 text-xs font-mono bg-slate-950 border border-slate-700 rounded-lg text-slate-200 focus:outline-none"
              >
                <option value="ClusterIP">ClusterIP (Internal)</option>
                <option value="NodePort">NodePort (Port 30000+)</option>
                <option value="LoadBalancer">LoadBalancer (Public IP)</option>
              </select>
              <p class="text-[11px] text-slate-500 mt-1">Routing type in cluster.</p>
            </div>
          </div>
        </div>
      </div>

      <!-- TAB 2: Capacity & Resources -->
      <div v-show="activeTab === 'capacity'" class="space-y-6 max-w-3xl">
        <div class="p-5 rounded-xl bg-slate-900/80 border border-slate-800 space-y-5 shadow-sm">
          <div>
            <h3 class="text-sm font-semibold text-slate-200 flex items-center gap-2">
              <i class="pi pi-bolt text-amber-400"></i>
              <span>Hardware Allocation (Cloud Run Tiers)</span>
            </h3>
            <p class="text-xs text-slate-400 mt-1">
              Select a pre-tuned compute profile or customize CPU and Memory limits directly.
            </p>
          </div>

          <!-- Tier Cards -->
          <div class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-4 gap-3">
            <!-- Micro -->
            <button
              type="button"
              class="p-4 rounded-xl text-left border transition cursor-pointer flex flex-col justify-between"
              :class="
                selectedTier === 'micro'
                  ? 'bg-sky-500/10 border-sky-500/60 ring-1 ring-sky-500/40'
                  : 'bg-slate-950/60 border-slate-800 hover:border-slate-700'
              "
              @click="applyTier('micro')"
            >
              <div>
                <span class="text-xs font-bold text-slate-300">Tier 1: Micro</span>
                <div class="text-lg font-mono font-bold text-sky-400 mt-1">0.25 vCPU</div>
                <div class="text-xs font-mono text-slate-400">256 MiB RAM</div>
              </div>
              <span class="text-[10px] text-slate-500 mt-3">Ideal for small APIs, webhooks</span>
            </button>

            <!-- Standard -->
            <button
              type="button"
              class="p-4 rounded-xl text-left border transition cursor-pointer flex flex-col justify-between"
              :class="
                selectedTier === 'standard'
                  ? 'bg-sky-500/10 border-sky-500/60 ring-1 ring-sky-500/40'
                  : 'bg-slate-950/60 border-slate-800 hover:border-slate-700'
              "
              @click="applyTier('standard')"
            >
              <div>
                <span class="text-xs font-bold text-slate-300">Tier 2: Standard</span>
                <div class="text-lg font-mono font-bold text-sky-400 mt-1">0.5 vCPU</div>
                <div class="text-xs font-mono text-slate-400">512 MiB RAM</div>
              </div>
              <span class="text-[10px] text-slate-500 mt-3">Recommended for general web apps</span>
            </button>

            <!-- Performance -->
            <button
              type="button"
              class="p-4 rounded-xl text-left border transition cursor-pointer flex flex-col justify-between"
              :class="
                selectedTier === 'performance'
                  ? 'bg-sky-500/10 border-sky-500/60 ring-1 ring-sky-500/40'
                  : 'bg-slate-950/60 border-slate-800 hover:border-slate-700'
              "
              @click="applyTier('performance')"
            >
              <div>
                <span class="text-xs font-bold text-slate-300">Tier 3: Performance</span>
                <div class="text-lg font-mono font-bold text-sky-400 mt-1">1.0 vCPU</div>
                <div class="text-xs font-mono text-slate-400">1 GiB RAM</div>
              </div>
              <span class="text-[10px] text-slate-500 mt-3">Production workloads & backends</span>
            </button>

            <!-- High Compute -->
            <button
              type="button"
              class="p-4 rounded-xl text-left border transition cursor-pointer flex flex-col justify-between"
              :class="
                selectedTier === 'high'
                  ? 'bg-sky-500/10 border-sky-500/60 ring-1 ring-sky-500/40'
                  : 'bg-slate-950/60 border-slate-800 hover:border-slate-700'
              "
              @click="applyTier('high')"
            >
              <div>
                <span class="text-xs font-bold text-slate-300">Tier 4: High</span>
                <div class="text-lg font-mono font-bold text-sky-400 mt-1">2.0 vCPU</div>
                <div class="text-xs font-mono text-slate-400">2 GiB RAM</div>
              </div>
              <span class="text-[10px] text-slate-500 mt-3">Heavy computation & databases</span>
            </button>
          </div>

          <!-- Custom Limits Detail -->
          <div class="pt-4 border-t border-slate-800 space-y-4">
            <div class="flex items-center justify-between">
              <span class="text-xs font-semibold text-slate-300 uppercase tracking-wider">
                Fine-Grained Kubernetes Resource Allocation
              </span>
              <span class="text-[11px] text-slate-500">Requests (Guaranteed) vs Limits (Hard Cap)</span>
            </div>

            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <!-- CPU -->
              <div class="p-3.5 rounded-lg bg-slate-950/70 border border-slate-800 space-y-3">
                <div class="text-xs font-bold text-slate-200 flex items-center gap-1.5">
                  <i class="pi pi-microchip text-sky-400"></i>
                  <span>CPU (Cores / Millicores)</span>
                </div>
                <div class="grid grid-cols-2 gap-2">
                  <div>
                    <label class="block text-[10px] text-slate-400 uppercase">Request</label>
                    <InputText
                      v-model="cpuRequest"
                      placeholder="100m"
                      class="w-full font-mono text-xs py-1.5 bg-slate-900 border-slate-700"
                      @input="selectedTier = 'custom'"
                    />
                  </div>
                  <div>
                    <label class="block text-[10px] text-slate-400 uppercase">Limit</label>
                    <InputText
                      v-model="cpuLimit"
                      placeholder="500m"
                      class="w-full font-mono text-xs py-1.5 bg-slate-900 border-slate-700"
                      @input="selectedTier = 'custom'"
                    />
                  </div>
                </div>
              </div>

              <!-- Memory -->
              <div class="p-3.5 rounded-lg bg-slate-950/70 border border-slate-800 space-y-3">
                <div class="text-xs font-bold text-slate-200 flex items-center gap-1.5">
                  <i class="pi pi-database text-purple-400"></i>
                  <span>Memory (MiB / GiB)</span>
                </div>
                <div class="grid grid-cols-2 gap-2">
                  <div>
                    <label class="block text-[10px] text-slate-400 uppercase">Request</label>
                    <InputText
                      v-model="memoryRequest"
                      placeholder="256Mi"
                      class="w-full font-mono text-xs py-1.5 bg-slate-900 border-slate-700"
                      @input="selectedTier = 'custom'"
                    />
                  </div>
                  <div>
                    <label class="block text-[10px] text-slate-400 uppercase">Limit</label>
                    <InputText
                      v-model="memoryLimit"
                      placeholder="512Mi"
                      class="w-full font-mono text-xs py-1.5 bg-slate-900 border-slate-700"
                      @input="selectedTier = 'custom'"
                    />
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- TAB 3: Scaling & Autoscaling -->
      <div v-show="activeTab === 'scaling'" class="space-y-6 max-w-3xl">
        <div class="p-5 rounded-xl bg-slate-900/80 border border-slate-800 space-y-5 shadow-sm">
          <div>
            <h3 class="text-sm font-semibold text-slate-200 flex items-center gap-2">
              <i class="pi pi-sliders-h text-cyan-400"></i>
              <span>Replicas & Auto-Scaling Strategy</span>
            </h3>
            <p class="text-xs text-slate-400 mt-1">
              Choose between a fixed pod count or automatic scaling based on real-time traffic/CPU load.
            </p>
          </div>

          <!-- Strategy Toggle -->
          <div class="grid grid-cols-1 sm:grid-cols-3 gap-3">
            <button
              type="button"
              class="p-4 rounded-xl text-left border transition cursor-pointer"
              :class="
                scalingMode === 'keda'
                  ? 'bg-cyan-500/10 border-cyan-500/60 ring-1 ring-cyan-500/40'
                  : 'bg-slate-950/60 border-slate-800 hover:border-slate-700'
              "
              @click="scalingMode = 'keda'"
            >
              <div class="flex items-center gap-2">
                <i class="pi pi-bolt text-cyan-400"></i>
                <span class="font-bold text-xs text-slate-200">KEDA HTTP (Per-Request)</span>
              </div>
              <p class="text-[11px] text-slate-400 mt-1.5">
                Scale pods on request concurrency. Supports Scale-to-Zero when idle.
              </p>
            </button>

            <button
              type="button"
              class="p-4 rounded-xl text-left border transition cursor-pointer"
              :class="
                scalingMode === 'hpa'
                  ? 'bg-sky-500/10 border-sky-500/60 ring-1 ring-sky-500/40'
                  : 'bg-slate-950/60 border-slate-800 hover:border-slate-700'
              "
              @click="scalingMode = 'hpa'"
            >
              <div class="flex items-center gap-2">
                <i class="pi pi-chart-line text-emerald-400"></i>
                <span class="font-bold text-xs text-slate-200">Resource HPA (CPU)</span>
              </div>
              <p class="text-[11px] text-slate-400 mt-1.5">
                Scale pods between Min and Max limits based on CPU/RAM usage.
              </p>
            </button>

            <button
              type="button"
              class="p-4 rounded-xl text-left border transition cursor-pointer"
              :class="
                scalingMode === 'manual'
                  ? 'bg-sky-500/10 border-sky-500/60 ring-1 ring-sky-500/40'
                  : 'bg-slate-950/60 border-slate-800 hover:border-slate-700'
              "
              @click="scalingMode = 'manual'"
            >
              <div class="flex items-center gap-2">
                <i class="pi pi-clone text-sky-400"></i>
                <span class="font-bold text-xs text-slate-200">Manual Replicas</span>
              </div>
              <p class="text-[11px] text-slate-400 mt-1.5">
                Always run a static number of pod replicas regardless of traffic.
              </p>
            </button>
          </div>

          <!-- KEDA HTTP Form -->
          <div v-if="scalingMode === 'keda'" class="p-4 rounded-xl bg-slate-950 border border-slate-800 space-y-4">
            <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
              <div>
                <label class="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
                  Min Replicas (Scale-to-0)
                </label>
                <input
                  v-model.number="minReplicas"
                  type="number"
                  min="0"
                  max="100"
                  placeholder="0"
                  class="w-full h-9 px-3 font-mono text-xs rounded-lg border border-slate-700 bg-slate-900 text-slate-100 focus:outline-none focus:ring-1 focus:ring-cyan-500"
                />
                <p class="text-[11px] text-slate-500 mt-1">Set to 0 to shut down pods when idle</p>
              </div>

              <div>
                <label class="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
                  Max Replicas
                </label>
                <input
                  v-model.number="maxReplicas"
                  type="number"
                  :min="minReplicas"
                  max="100"
                  placeholder="3"
                  class="w-full h-9 px-3 font-mono text-xs rounded-lg border border-slate-700 bg-slate-900 text-slate-100 focus:outline-none focus:ring-1 focus:ring-cyan-500"
                />
                <p class="text-[11px] text-slate-500 mt-1">Maximum scale ceiling</p>
              </div>

              <div>
                <label class="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
                  Target Concurrency
                </label>
                <input
                  v-model.number="targetConcurrency"
                  type="number"
                  min="1"
                  max="1000"
                  placeholder="30"
                  class="w-full h-9 px-3 font-mono text-xs rounded-lg border border-slate-700 bg-slate-900 text-slate-100 focus:outline-none focus:ring-1 focus:ring-cyan-500"
                />
                <p class="text-[11px] text-slate-500 mt-1">Concurrent requests / pod before scale</p>
              </div>
            </div>
          </div>

          <!-- Manual Replicas Form -->
          <div v-if="scalingMode === 'manual'" class="p-4 rounded-xl bg-slate-950 border border-slate-800">
            <label class="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-2">
              Number of Replicas
            </label>
            <div class="flex items-center gap-4">
              <div class="inline-flex items-center rounded-xl border border-slate-700 bg-slate-900 p-1 shadow-xs">
                <button
                  type="button"
                  class="w-8 h-8 rounded-lg flex items-center justify-center text-slate-400 hover:text-white hover:bg-slate-800 transition cursor-pointer disabled:opacity-30"
                  :disabled="manualReplicas <= 1"
                  @click="manualReplicas = Math.max(1, manualReplicas - 1)"
                >
                  <i class="pi pi-minus text-xs"></i>
                </button>
                <input
                  v-model.number="manualReplicas"
                  type="number"
                  min="1"
                  max="100"
                  class="w-14 text-center font-mono font-bold text-sm bg-transparent border-0 text-slate-100 focus:outline-none"
                />
                <button
                  type="button"
                  class="w-8 h-8 rounded-lg flex items-center justify-center text-slate-400 hover:text-white hover:bg-slate-800 transition cursor-pointer disabled:opacity-30"
                  :disabled="manualReplicas >= 100"
                  @click="manualReplicas = Math.min(100, manualReplicas + 1)"
                >
                  <i class="pi pi-plus text-xs"></i>
                </button>
              </div>
              <span class="text-xs text-slate-400 font-mono">pod instance(s) running simultaneously</span>
            </div>
          </div>

          <!-- HPA Form -->
          <div v-if="scalingMode === 'hpa'" class="p-4 rounded-xl bg-slate-950 border border-slate-800 space-y-4">
            <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
              <div>
                <label class="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
                  Min Replicas
                </label>
                <input
                  v-model.number="minReplicas"
                  type="number"
                  min="1"
                  max="100"
                  placeholder="1"
                  class="w-full h-9 px-3 font-mono text-xs rounded-lg border border-slate-700 bg-slate-900 text-slate-100 focus:outline-none focus:ring-1 focus:ring-emerald-500"
                />
                <p class="text-[11px] text-slate-500 mt-1">Minimum active pods (idle load)</p>
              </div>

              <div>
                <label class="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
                  Max Replicas
                </label>
                <input
                  v-model.number="maxReplicas"
                  type="number"
                  :min="minReplicas"
                  max="100"
                  placeholder="5"
                  class="w-full h-9 px-3 font-mono text-xs rounded-lg border border-slate-700 bg-slate-900 text-slate-100 focus:outline-none focus:ring-1 focus:ring-emerald-500"
                />
                <p class="text-[11px] text-slate-500 mt-1">Ceiling during peak traffic spikes</p>
              </div>

              <div>
                <label class="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
                  Target CPU Utilization (%)
                </label>
                <input
                  v-model.number="targetCpuPct"
                  type="number"
                  min="10"
                  max="100"
                  placeholder="75"
                  class="w-full h-9 px-3 font-mono text-xs rounded-lg border border-slate-700 bg-slate-900 text-slate-100 focus:outline-none focus:ring-1 focus:ring-emerald-500"
                />
                <p class="text-[11px] text-slate-500 mt-1">Trigger scale up when avg CPU crosses this</p>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- TAB 4: Environment & Secrets -->
      <div v-show="activeTab === 'env'" class="space-y-6 max-w-3xl">
        <div class="p-5 rounded-xl bg-slate-900/80 border border-slate-800 space-y-4 shadow-sm">
          <div class="flex items-center justify-between">
            <div>
              <h3 class="text-sm font-semibold text-slate-200 flex items-center gap-2">
                <i class="pi pi-lock text-amber-400"></i>
                <span>Environment Variables & Secret Injections</span>
              </h3>
              <p class="text-xs text-slate-400 mt-0.5">
                Pass configuration keys or bind sensitive values directly.
              </p>
            </div>

            <div class="flex items-center gap-2">
              <Button
                :label="envPasteMode ? 'Key-Value View' : 'Paste .env'"
                :icon="envPasteMode ? 'pi pi-list' : 'pi pi-file-edit'"
                size="small"
                outlined
                severity="secondary"
                class="text-xs"
                @click="envPasteMode = !envPasteMode"
              />
              <Button
                label="Add Variable"
                icon="pi pi-plus"
                size="small"
                class="btn-sky text-xs"
                @click="addEnvVar"
              />
            </div>
          </div>

          <!-- Paste Mode -->
          <div v-if="envPasteMode" class="space-y-3">
            <textarea
              v-model="envPasteText"
              rows="6"
              placeholder="PORT=8080&#10;NODE_ENV=production&#10;DATABASE_URL=postgres://..."
              class="w-full p-3 font-mono text-xs bg-slate-950 text-emerald-400 rounded-lg border border-slate-800 focus:outline-none resize-none"
            ></textarea>
            <div class="flex justify-end gap-2">
              <Button label="Cancel" severity="secondary" text size="small" @click="envPasteMode = false" />
              <Button label="Append Variables" size="small" class="btn-emerald text-xs" @click="importPastedEnv" />
            </div>
          </div>

          <!-- Visual Key-Value Editor -->
          <div v-else class="space-y-2.5">
            <div
              v-for="(e, idx) in envVars"
              :key="idx"
              class="flex items-center gap-2 p-2 rounded-lg bg-slate-950/60 border border-slate-800"
            >
              <InputText
                v-model="e.name"
                placeholder="VARIABLE_NAME"
                class="w-1/3 font-mono text-xs py-1.5 bg-slate-900 border-slate-700"
              />
              <span class="text-slate-600 font-mono">=</span>
              <InputText
                v-model="e.value"
                placeholder="value or plain text"
                class="flex-1 font-mono text-xs py-1.5 bg-slate-900 border-slate-700"
              />
              <Button
                icon="pi pi-trash"
                severity="danger"
                text
                rounded
                size="small"
                class="text-slate-500 hover:text-rose-400"
                @click="removeEnvVar(idx)"
              />
            </div>

            <div
              v-if="envVars.length === 0"
              class="py-8 text-center border border-dashed border-slate-800 rounded-xl text-slate-500 text-xs"
            >
              <i class="pi pi-info-circle mb-1 text-sm text-slate-600"></i>
              <p>No environment variables defined yet.</p>
              <button
                type="button"
                class="text-sky-400 hover:underline mt-1 inline-block cursor-pointer"
                @click="addEnvVar"
              >
                + Add your first variable
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- TAB 5: Live YAML Preview -->
      <div v-show="activeTab === 'yaml'" class="space-y-4 max-w-4xl">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-2 text-xs text-slate-400">
            <i class="pi pi-info-circle text-purple-400"></i>
            <span>This manifest is dynamically generated from your visual configuration.</span>
          </div>

          <Button
            :label="yamlCopied ? 'Copied!' : 'Copy YAML'"
            :icon="yamlCopied ? 'pi pi-check' : 'pi pi-copy'"
            size="small"
            severity="secondary"
            class="text-xs"
            @click="copyYaml"
          />
        </div>

        <pre
          class="p-4 rounded-xl bg-slate-950 border border-slate-800 text-purple-300 font-mono text-xs overflow-x-auto leading-relaxed shadow-inner"
        ><code>{{ generatedYaml }}</code></pre>
      </div>
    </div>

    <!-- Footer Summary Strip & Action Buttons -->
    <div
      class="px-6 py-3.5 bg-slate-900 border-t border-slate-800 flex flex-col sm:flex-row items-center justify-between gap-3 shrink-0"
    >
      <!-- Summary Chips -->
      <div class="flex flex-wrap items-center gap-2 text-xs text-slate-400">
        <span class="font-mono text-slate-200 font-semibold">
          {{ workloadName || 'unnamed-workload' }}
        </span>
        <span class="text-slate-600">•</span>
        <span class="font-mono text-sky-400">
          {{ cpuLimit }} / {{ memoryLimit }}
        </span>
        <span class="text-slate-600">•</span>
        <span class="font-mono text-cyan-400">
          Port {{ containerPort || 'none' }} ({{ exposeService ? serviceType : 'No Service' }})
        </span>
        <span class="text-slate-600">•</span>
        <span class="font-mono text-emerald-400">
          {{ scalingMode === 'manual' ? `${manualReplicas} Replicas` : `HPA: ${minReplicas}-${maxReplicas} (CPU ${targetCpuPct}%)` }}
        </span>
      </div>

      <!-- Actions -->
      <div class="flex items-center gap-2.5 w-full sm:w-auto justify-end">
        <Button
          label="Cancel"
          severity="secondary"
          text
          size="small"
          class="text-xs text-slate-400 hover:text-white"
          @click="emit('update:visible', false)"
        />
        <Button
          label="Deploy Workload"
          icon="pi pi-check"
          size="small"
          class="btn-sky text-xs font-semibold px-4 py-2 shadow-lg shadow-sky-500/20"
          :loading="isDeploying"
          @click="deployWorkload"
        />
      </div>
    </div>
  </Dialog>
</template>

<style scoped></style>

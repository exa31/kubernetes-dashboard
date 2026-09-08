import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

import { authApi } from '@/api'
import type { AuthUser, LoginCredentials } from '@/types'
import { logger } from '@/utils'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<AuthUser | null>((() => {
    try {
      const saved = localStorage.getItem('kubeenv_user')
      return saved ? JSON.parse(saved) : null
    } catch {
      return null
    }
  })())
  const isLoading = ref(false)
  const isAuthenticated = computed(() => !!user.value)
  const userRole = computed(() => user.value?.role || 'viewer')
  const isAdmin = computed(() => user.value?.role === 'admin')
  const isDevOps = computed(() => user.value?.role === 'devops' || user.value?.role === 'admin')
  const isViewer = computed(() => user.value?.role === 'viewer')

  const allowedNamespaces = computed(() => user.value?.allowed_namespaces || '*')
  const allowedNamespacesList = computed<string[]>(() => {
    const raw = (user.value?.allowed_namespaces || '*').trim()
    if (!raw || raw === '*') return ['*']
    return raw.split(',').map((s) => s.trim().toLowerCase()).filter(Boolean)
  })

  const systemNamespaces = ['kube-system', 'kube-public', 'kube-node-lease']

  function canReadNamespace(targetNamespace?: string): boolean {
    if (!targetNamespace) return true
    if (isAdmin.value) return true
    if (allowedNamespacesList.value.includes('*')) return true
    return allowedNamespacesList.value.includes(targetNamespace.toLowerCase())
  }

  function canMutateNamespace(targetNamespace?: string): boolean {
    if (!targetNamespace || !isAuthenticated.value) return false
    if (isViewer.value) return false
    if (isAdmin.value) return true

    // DevOps: block system namespaces and check allowed list
    if (systemNamespaces.includes(targetNamespace.toLowerCase())) {
      return false
    }
    if (allowedNamespacesList.value.includes('*')) return true
    return allowedNamespacesList.value.includes(targetNamespace.toLowerCase())
  }

  async function login(credentials: LoginCredentials): Promise<void> {
    isLoading.value = true
    try {
      const data = await authApi.login(credentials)
      user.value = data.user
      try {
        localStorage.setItem('kubeenv_user', JSON.stringify(data.user))
        if (data.tokens?.refresh_token) {
          localStorage.setItem('kubeenv_refresh_token', data.tokens.refresh_token)
        }
        if (data.tokens?.access_token) {
          localStorage.setItem('kubeenv_access_token', data.tokens.access_token)
        }
      } catch (err) {
        logger.warn('Failed to persist user profile and tokens', err)
      }
      logger.info(`User logged in: ${data.user.email}`)
    } finally {
      isLoading.value = false
    }
  }

  async function checkAuth(): Promise<boolean> {
    try {
      const profile = await authApi.getProfile()
      user.value = profile
      localStorage.setItem('kubeenv_user', JSON.stringify(profile))
      return true
    } catch (err) {
      logger.warn('Session verification failed, logging out', err)
      user.value = null
      localStorage.removeItem('kubeenv_user')
      localStorage.removeItem('kubeenv_refresh_token')
      localStorage.removeItem('kubeenv_access_token')
      return false
    }
  }

  async function logout(): Promise<void> {
    isLoading.value = true
    try {
      await authApi.logout()
    } catch (err) {
      logger.warn('Logout API error, clearing local state anyway', err)
    } finally {
      user.value = null
      localStorage.removeItem('kubeenv_user')
      localStorage.removeItem('kubeenv_refresh_token')
      localStorage.removeItem('kubeenv_access_token')
      isLoading.value = false
    }
  }

  return {
    user,
    isLoading,
    isAuthenticated,
    userRole,
    isAdmin,
    isDevOps,
    isViewer,
    allowedNamespaces,
    allowedNamespacesList,
    canReadNamespace,
    canMutateNamespace,
    login,
    checkAuth,
    logout,
  }
})

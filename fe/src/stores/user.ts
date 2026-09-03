import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

import { userApi } from '@/api'
import type { CreateUserPayload, ResetPasswordPayload, UpdateUserPayload, User, UserRole } from '@/types'
import { logger } from '@/utils'

export const useUserStore = defineStore('user', () => {
  const users = ref<User[]>([])
  const isLoading = ref(false)
  const isActionLoading = ref(false)
  const error = ref<string | null>(null)

  // Filter states
  const searchQuery = ref('')
  const roleFilter = ref<'all' | UserRole>('all')
  const statusFilter = ref<'all' | 'active' | 'inactive'>('all')

  const totalUsers = computed(() => users.value.length)
  const adminCount = computed(() => users.value.filter((u) => u.role === 'admin').length)
  const devopsCount = computed(() => users.value.filter((u) => u.role === 'devops').length)
  const viewerCount = computed(() => users.value.filter((u) => u.role === 'viewer').length)
  const activeCount = computed(() => users.value.filter((u) => u.is_active).length)

  const filteredUsers = computed(() => {
    return users.value.filter((u) => {
      // Search by name or email
      if (searchQuery.value) {
        const query = searchQuery.value.toLowerCase()
        const matchName = u.name.toLowerCase().includes(query)
        const matchEmail = u.email.toLowerCase().includes(query)
        if (!matchName && !matchEmail) return false
      }

      // Filter by role
      if (roleFilter.value !== 'all' && u.role !== roleFilter.value) {
        return false
      }

      // Filter by status
      if (statusFilter.value === 'active' && !u.is_active) {
        return false
      }
      if (statusFilter.value === 'inactive' && u.is_active) {
        return false
      }

      return true
    })
  })

  function getErrorMessage(err: unknown, fallback: string): string {
    if (err && typeof err === 'object' && 'response' in err) {
      const res = (err as { response?: { data?: { message?: string } } }).response
      if (res?.data?.message) return res.data.message
    }
    if (err instanceof Error) return err.message
    return fallback
  }

  async function fetchUsers() {
    isLoading.value = true
    error.value = null
    try {
      users.value = await userApi.getUsers()
    } catch (err) {
      logger.error('Failed to fetch users', err)
      error.value = getErrorMessage(err, 'Failed to fetch users')
    } finally {
      isLoading.value = false
    }
  }

  async function createUser(payload: CreateUserPayload) {
    isActionLoading.value = true
    error.value = null
    try {
      const newUser = await userApi.createUser(payload)
      users.value.unshift(newUser)
      return newUser
    } catch (err) {
      error.value = getErrorMessage(err, 'Failed to create user')
      throw err
    } finally {
      isActionLoading.value = false
    }
  }

  async function updateUser(id: string, payload: UpdateUserPayload) {
    isActionLoading.value = true
    error.value = null
    try {
      const updated = await userApi.updateUser(id, payload)
      const index = users.value.findIndex((u) => u.id === id)
      if (index !== -1) {
        users.value[index] = updated
      }
      return updated
    } catch (err) {
      error.value = getErrorMessage(err, 'Failed to update user')
      throw err
    } finally {
      isActionLoading.value = false
    }
  }

  async function resetPassword(id: string, payload: ResetPasswordPayload) {
    isActionLoading.value = true
    error.value = null
    try {
      await userApi.resetPassword(id, payload)
    } catch (err) {
      error.value = getErrorMessage(err, 'Failed to reset password')
      throw err
    } finally {
      isActionLoading.value = false
    }
  }

  async function deleteUser(id: string) {
    isActionLoading.value = true
    error.value = null
    try {
      await userApi.deleteUser(id)
      const index = users.value.findIndex((u) => u.id === id)
      if (index !== -1) {
        users.value[index].is_active = false
      }
    } catch (err) {
      error.value = getErrorMessage(err, 'Failed to deactivate user')
      throw err
    } finally {
      isActionLoading.value = false
    }
  }

  async function hardDeleteUser(id: string) {
    isActionLoading.value = true
    error.value = null
    try {
      await userApi.hardDeleteUser(id)
      users.value = users.value.filter((u) => u.id !== id)
    } catch (err) {
      error.value = getErrorMessage(err, 'Failed to permanently delete user')
      throw err
    } finally {
      isActionLoading.value = false
    }
  }

  return {
    users,
    isLoading,
    isActionLoading,
    error,
    searchQuery,
    roleFilter,
    statusFilter,
    totalUsers,
    adminCount,
    devopsCount,
    viewerCount,
    activeCount,
    filteredUsers,
    fetchUsers,
    createUser,
    updateUser,
    resetPassword,
    deleteUser,
    hardDeleteUser,
  }
})

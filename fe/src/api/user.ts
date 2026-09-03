import type { CreateUserPayload, ResetPasswordPayload, UpdateUserPayload, User } from '@/types'
import { apiClient } from './client'

export const userApi = {
  getUsers: async (): Promise<User[]> => {
    const res = await apiClient.get<{ data: User[] }>('/users')
    return res.data.data ?? []
  },

  getUser: async (id: string): Promise<User> => {
    const res = await apiClient.get<{ data: User }>(`/users/${encodeURIComponent(id)}`)
    return res.data.data
  },

  createUser: async (payload: CreateUserPayload): Promise<User> => {
    const res = await apiClient.post<{ data: User }>('/users', payload)
    return res.data.data
  },

  updateUser: async (id: string, payload: UpdateUserPayload): Promise<User> => {
    const res = await apiClient.put<{ data: User }>(`/users/${encodeURIComponent(id)}`, payload)
    return res.data.data
  },

  resetPassword: async (id: string, payload: ResetPasswordPayload): Promise<void> => {
    await apiClient.post(`/users/${encodeURIComponent(id)}/reset-password`, payload)
  },

  deleteUser: async (id: string): Promise<void> => {
    await apiClient.delete(`/users/${encodeURIComponent(id)}`)
  },

  hardDeleteUser: async (id: string): Promise<void> => {
    await apiClient.delete(`/users/admin/${encodeURIComponent(id)}`)
  },
}

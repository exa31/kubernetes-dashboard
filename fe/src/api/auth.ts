import type { AuthResponseData, AuthUser, LoginCredentials } from '@/types'

import { apiClient } from './client'

export const authApi = {
  login: async (credentials: LoginCredentials): Promise<AuthResponseData> => {
    const res = await apiClient.post<{ data: AuthResponseData }>('/auth/login', credentials)
    return res.data.data
  },

  getProfile: async (): Promise<AuthUser> => {
    const res = await apiClient.get<{ data: AuthUser }>('/auth/profile')
    return res.data.data
  },

  refreshToken: async (): Promise<void> => {
    await apiClient.post('/auth/refresh')
  },

  logout: async (): Promise<void> => {
    await apiClient.post('/auth/logout')
  },
}

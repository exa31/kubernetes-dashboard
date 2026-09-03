import type { AxiosError, InternalAxiosRequestConfig } from 'axios'
import axios from 'axios'

import { logger } from '@/utils'

// Create Axios client with cookie-only credentials
export const apiClient = axios.create({
  baseURL: '/api/v1',
  withCredentials: true, // Required for cookie-only authentication
  headers: {
    'Content-Type': 'application/json',
    Accept: 'application/json',
  },
})

// Request tracking & Refresh token state
let isRefreshing = false
let failedQueue: Array<{
  resolve: (value?: unknown) => void
  reject: (reason?: unknown) => void
}> = []

const processQueue = (error: unknown = null) => {
  failedQueue.forEach((promise) => {
    if (error) {
      promise.reject(error)
    } else {
      promise.resolve()
    }
  })
  failedQueue = []
}

// Request Interceptor
apiClient.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    logger.info(`[HTTP ${config.method?.toUpperCase()}] ${config.url}`)
    return config
  },
  (error: AxiosError) => Promise.reject(error),
)

// Response Interceptor with Cookie-Only Refresh Token Interception
apiClient.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const originalRequest = error.config as InternalAxiosRequestConfig & { _retry?: boolean }

    // Check if error is 401 Unauthorized
    if (error.response?.status === 401 && originalRequest) {
      const requestUrl = originalRequest.url || ''

      // Do not try to refresh if the failed request was login, refresh, or logout itself
      if (
        requestUrl.includes('/auth/login') ||
        requestUrl.includes('/auth/refresh') ||
        requestUrl.includes('/auth/logout')
      ) {
        return Promise.reject(error)
      }

      // If already retried once, don't loop
      if (originalRequest._retry) {
        return Promise.reject(error)
      }

      if (isRefreshing) {
        // Queue this request while refresh is in flight
        return new Promise((resolve, reject) => {
          failedQueue.push({ resolve, reject })
        })
          .then(() => apiClient(originalRequest))
          .catch((err) => Promise.reject(err))
      }

      originalRequest._retry = true
      isRefreshing = true

      try {
        logger.info('[Auth] Access token expired. Triggering token refresh...')
        const storedRefreshToken = typeof window !== 'undefined' ? localStorage.getItem('kubeenv_refresh_token') || '' : ''
        
        // Multi-channel refresh: withCredentials sends cookies automatically, plus fallback body & headers
        const res = await apiClient.post<{ data: { access_token?: string; refresh_token?: string } }>(
          '/auth/refresh',
          { refresh_token: storedRefreshToken },
          {
            headers: storedRefreshToken ? {
              'X-Refresh-Token': storedRefreshToken,
              'Authorization': `Bearer ${storedRefreshToken}`,
            } : {},
          }
        )

        if (res.data?.data?.refresh_token && typeof window !== 'undefined') {
          localStorage.setItem('kubeenv_refresh_token', res.data.data.refresh_token)
        }
        if (res.data?.data?.access_token && typeof window !== 'undefined') {
          localStorage.setItem('kubeenv_access_token', res.data.data.access_token)
        }

        logger.info('[Auth] Token refreshed successfully')
        processQueue(null)
        return apiClient(originalRequest)
      } catch (refreshError) {
        logger.warn('[Auth] Token refresh failed. Redirecting to login...', refreshError)
        processQueue(refreshError)

        // Clear local state or redirect to login if browser window exists
        if (typeof window !== 'undefined' && !window.location.pathname.includes('/auth/login')) {
          window.location.href = '/auth/login'
        }

        return Promise.reject(refreshError)
      } finally {
        isRefreshing = false
      }
    }

    return Promise.reject(error)
  },
)

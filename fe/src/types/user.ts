/**
 * User domain types & payloads for RBAC
 */
export type UserRole = 'admin' | 'devops' | 'viewer'

export interface User {
  id: string
  name: string
  email: string
  role: UserRole
  allowed_namespaces?: string
  phone?: string
  avatarUrl?: string
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface CreateUserPayload {
  name: string
  email: string
  password: string
  role: UserRole
  allowed_namespaces?: string
  phone?: string
}

export interface UpdateUserPayload {
  name?: string
  email?: string
  role?: UserRole
  allowed_namespaces?: string
  phone?: string
  is_active?: boolean
}

export interface ResetPasswordPayload {
  new_password: string
}

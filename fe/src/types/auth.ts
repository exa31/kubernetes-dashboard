export interface AuthUser {
  id: string
  name: string
  email: string
  role?: 'admin' | 'devops' | 'viewer'
  allowed_namespaces?: string
  created_at?: string
}

export interface LoginCredentials {
  email: string
  password: string
}

export interface AuthResponseData {
  user: AuthUser
  tokens?: {
    access_token: string
    refresh_token: string
  }
}

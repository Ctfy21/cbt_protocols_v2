export type UserRole = 'admin' | 'user'

export interface User {
  id: string
  username: string
  name: string
  role: UserRole
  is_active: boolean
  created_at: string
  updated_at: string
  last_login?: string
}

export interface LoginCredentials {
  username: string
  password: string
}

export interface RegisterData {
  username: string
  password: string
  name: string
}

export interface AuthResponse {
  user: User
  token: string
  refresh_token: string
  expires_in: number
}

export interface ChangePasswordData {
  current_password: string
  new_password: string
}

export interface ApiToken {
  id: string
  name: string
  type: 'personal' | 'service'
  service_name?: string
  permissions: string[]
  expires_at?: string
  created_at: string
  token: string
}
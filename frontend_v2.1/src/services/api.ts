import axios from 'axios'
import type { AxiosInstance } from 'axios'
import type { ApiResponse, Chamber, Experiment, ExperimentFormData } from '@/types'
import type { LoginCredentials, RegisterData, AuthResponse, User, ApiToken } from '@/types/auth'
import { useAuthStore } from '@/stores/auth'

class ApiService {
  public api: AxiosInstance

  constructor() {
    // Use relative URL for production, localhost for development
    const baseURL = import.meta.env.DEV ? 'http://localhost:8080' : ''
    
    this.api = axios.create({
      baseURL: baseURL + '/api',
      headers: {
        'Content-Type': 'application/json'
      }
    })

    // Request interceptor
    this.api.interceptors.request.use(
      (config) => {
        // Get token from auth store instead of localStorage
        const authStore = useAuthStore()
        if (authStore.token) {
          config.headers.Authorization = `Bearer ${authStore.token}`
        }
        return config
      },
      (error) => {
        return Promise.reject(error)
      }
    )

    // Response interceptor
    this.api.interceptors.response.use(
      (response) => {
        // Ensure response data has success property
        if (response.data && typeof response.data === 'object') {
          if (!('success' in response.data)) {
            response.data = {
              success: true,
              data: response.data
            }
          }
        }
        return response
      },
      (error) => {
        console.error('API Error:', error)
        return Promise.reject(error)
      }
    )
  }

  // Chamber endpoints
  async getChambers(): Promise<ApiResponse<Chamber[]>> {
    const response = await this.api.get('/chambers')
    return response.data
  }

  async getChamber(id: string): Promise<ApiResponse<Chamber>> {
    const response = await this.api.get(`/chambers/${id}`)
    return response.data
  }

  async registerChamber(data: Partial<Chamber>): Promise<ApiResponse<Chamber>> {
    const response = await this.api.post('/chambers', data)
    return response.data
  }

  async updateChamberHeartbeat(id: string): Promise<ApiResponse<void>> {
    const response = await this.api.post(`/chambers/${id}/heartbeat`)
    return response.data
  }

  async getChamberWateringZones(id: string): Promise<ApiResponse<any>> {
    const response = await this.api.get(`/chambers/${id}/watering-zones`)
    return response.data
  }

  // Experiment endpoints
  async getExperiments(params?: { chamber_id?: string }): Promise<ApiResponse<Experiment[]>> {
    const response = await this.api.get('/experiments', { params })
    return response.data
  }

  async getExperiment(id: string): Promise<ApiResponse<Experiment>> {
    const response = await this.api.get(`/experiments/${id}`)
    return response.data
  }

  async createExperiment(data: ExperimentFormData): Promise<ApiResponse<Experiment>>{
    const response = await this.api.post('/experiments', data)
    return response.data
  }

  async updateExperiment(id: string, data: Partial<ExperimentFormData>): Promise<ApiResponse<Experiment>> {
    const response = await this.api.put(`/experiments/${id}`, data)
    return response.data
  }

  async deleteExperiment(id: string): Promise<ApiResponse<void>> {
    const response = await this.api.delete(`/experiments/${id}`)
    return response.data
  }

  // Auth endpoints
  async login(credentials: LoginCredentials): Promise<ApiResponse<AuthResponse>> {
    const response = await this.api.post('/auth/login', credentials)
    return response.data
  }

  async register(data: RegisterData): Promise<ApiResponse<AuthResponse>> {
    const response = await this.api.post('/auth/register', data)
    return response.data
  }

  async logout(): Promise<ApiResponse<void>> {
    const response = await this.api.post('/auth/logout')
    return response.data
  }

  async refreshToken(refreshToken: string): Promise<ApiResponse<AuthResponse>> {
    const response = await this.api.post('/auth/refresh', { refresh_token: refreshToken })
    return response.data
  }

  async getCurrentUser(): Promise<ApiResponse<User>> {
    const response = await this.api.get('/auth/me')
    return response.data
  }

  async updateProfile(data: { name?: string }): Promise<ApiResponse<User>> {
    const response = await this.api.put('/auth/profile', data)
    return response.data
  }

  async changePassword(currentPassword: string, newPassword: string): Promise<ApiResponse<void>> {
    const response = await this.api.put('/auth/password', {
      current_password: currentPassword,
      new_password: newPassword
    })
    return response.data
  }

  setAuthToken(token: string | null) {
    if (token) {
      this.api.defaults.headers.common['Authorization'] = `Bearer ${token}`
    } else {
      delete this.api.defaults.headers.common['Authorization']
    }
  }

  // User Management endpoints (Admin only)
  async getAllUsers(): Promise<ApiResponse<User[]>> {
    const response = await this.api.get('/users')
    return response.data
  }

  async getUser(id: string): Promise<ApiResponse<User>> {
    const response = await this.api.get(`/users/${id}`)
    return response.data
  }

  async createUser(data: {
    username: string
    password: string
    role: 'user' | 'admin'
  }): Promise<ApiResponse<User>> {
    const response = await this.api.post('/users', data)
    return response.data
  }

  async updateUser(id: string, data: {
    username?: string
    role?: 'user' | 'admin'
    is_active?: boolean
  }): Promise<ApiResponse<User>> {
    const response = await this.api.put(`/users/${id}`, data)
    return response.data
  }

  async deactivateUser(id: string): Promise<ApiResponse<void>> {
    const response = await this.api.delete(`/users/${id}`)
    return response.data
  }

  async activateUser(id: string): Promise<ApiResponse<void>> {
    const response = await this.api.post(`/users/${id}/activate`)
    return response.data
  }

  // API Token endpoints
  async getApiTokens(): Promise<ApiResponse<ApiToken[]>> {
    const response = await this.api.get('/api-tokens')
    return response.data
  }

  async createApiToken(data: {
    name: string
    type: 'personal' | 'service'
    service_name?: string
    permissions: string[]
    expires_at?: string
  }): Promise<ApiResponse<ApiToken>> {
    console.log(data)
    
    const response = await this.api.post('/api-tokens', data)
    return response.data
  }

  async deleteApiToken(id: string): Promise<ApiResponse<void>> {
    const response = await this.api.delete(`/api-tokens/${id}`)
    return response.data
  }

  // User Chamber Access endpoints
  async getAllUsersWithChamberAccess(): Promise<ApiResponse<any[]>> {
    const response = await this.api.get('/users/chambers')
    return response.data
  }

  async getUserChamberAccess(userId: string): Promise<ApiResponse<Chamber[]>> {
    const response = await this.api.get(`/users/${userId}/chambers`)
    return response.data
  }

  async setUserChamberAccess(userId: string, chamberIds: string[]): Promise<ApiResponse<void>> {
    const response = await this.api.put(`/users/${userId}/chambers`, { chamber_ids: chamberIds })
    return response.data
  }

  async grantChamberAccess(userId: string, chamberId: string): Promise<ApiResponse<void>> {
    const response = await this.api.post(`/users/${userId}/chambers/${chamberId}`)
    return response.data
  }

  async revokeChamberAccess(userId: string, chamberId: string): Promise<ApiResponse<void>> {
    const response = await this.api.delete(`/users/${userId}/chambers/${chamberId}`)
    return response.data
  }

  async hasChamberAccess(userId: string, chamberId: string): Promise<ApiResponse<{ has_access: boolean }>> {
    const response = await this.api.get(`/users/${userId}/chambers/${chamberId}/check`)
    return response.data
  }

  async getMyChamberAccess(): Promise<ApiResponse<Chamber[]>> {
    const response = await this.api.get('/me/chambers')
    return response.data
  }

  // Helper methods
  formatError(error: any): string {
    if (error.response?.data?.error) {
      return error.response.data.error
    } else if (error.response?.data?.message) {
      return error.response.data.message
    } else if (error.message) {
      return error.message
    }
    return 'An unexpected error occurred'
  }
}

export default new ApiService()
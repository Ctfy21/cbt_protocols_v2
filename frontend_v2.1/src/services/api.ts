import axios from 'axios'
import type { AxiosInstance } from 'axios'
import type { ApiResponse, Chamber, Experiment, ExperimentFormData } from '@/types'

class ApiService {
  private api: AxiosInstance

  constructor() {
    this.api = axios.create({
      baseURL: 'http://localhost:8080',
      headers: {
        'Content-Type': 'application/json'
      }
    })

    // Request interceptor
    this.api.interceptors.request.use(
      (config) => {
        // Add any auth headers here if needed
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
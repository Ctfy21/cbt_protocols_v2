import type { 
  ApiResponse, 
  AuthResponse, 
  LoginRequest, 
  User,
  Chamber,
  ChamberConfig,
  Experiment,
  APIToken,
  CreateAPITokenRequest,
  UserChamberAccess
} from '@/types';

class ApiClient {
  private baseUrl: string;
  private token: string | null = null;

  constructor() {
    this.baseUrl = import.meta.env.VITE_API_URL || '/api';
    this.token = localStorage.getItem('token');
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<ApiResponse<T>> {
    const url = `${this.baseUrl}${endpoint}`;
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
      ...options.headers,
    };

    if (this.token) {
      (headers as Record<string, string>)['Authorization'] = `Bearer ${this.token}`;
    }

    try {
      const response = await fetch(url, {
        ...options,
        headers,
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || `HTTP error! status: ${response.status}`);
      }

      return data;
    } catch (error) {
      console.error('API request failed:', error);
      throw error;
    }
  }

  setToken(token: string | null) {
    this.token = token;
    if (token) {
      localStorage.setItem('token', token);
    } else {
      localStorage.removeItem('token');
    }
  }

  // Auth endpoints
  async login(credentials: LoginRequest): Promise<AuthResponse> {
    const response = await this.request<AuthResponse>('/auth/login', {
      method: 'POST',
      body: JSON.stringify(credentials),
    });
    
    if (response.success && response.data) {
      this.setToken(response.data.token);
      return response.data;
    }
    
    throw new Error(response.error || 'Login failed');
  }

  async logout(): Promise<void> {
    await this.request('/auth/logout', { method: 'POST' });
    this.setToken(null);
  }

  async getMe(): Promise<User> {
    const response = await this.request<User>('/auth/me');
    if (response.success && response.data) {
      return response.data;
    }
    throw new Error(response.error || 'Failed to get user info');
  }

  async changePassword(currentPassword: string, newPassword: string): Promise<void> {
    const response = await this.request('/auth/change-password', {
      method: 'POST',
      body: JSON.stringify({
        current_password: currentPassword,
        new_password: newPassword,
      }),
    });
    
    if (!response.success) {
      throw new Error(response.error || 'Failed to change password');
    }
  }

  // Chamber endpoints
  async getChambers(): Promise<Chamber[]> {
    const response = await this.request<Chamber[]>('/chambers');
    if (response.success && response.data) {
      return response.data;
    }
    throw new Error(response.error || 'Failed to get chambers');
  }

  async getChamber(id: string): Promise<Chamber> {
    const response = await this.request<Chamber>(`/chambers/${id}`);
    if (response.success && response.data) {
      return response.data;
    }
    throw new Error(response.error || 'Failed to get chamber');
  }

  async getChamberConfig(id: string): Promise<ChamberConfig> {
    const response = await this.request<ChamberConfig>(`/chambers/${id}/config`);
    if (response.success && response.data) {
      return response.data;
    }
    throw new Error(response.error || 'Failed to get chamber config');
  }

  async updateChamberConfig(id: string, config: Partial<ChamberConfig>): Promise<ChamberConfig> {
    const response = await this.request<ChamberConfig>(`/chambers/${id}/config`, {
      method: 'PUT',
      body: JSON.stringify(config),
    });
    if (response.success && response.data) {
      return response.data;
    }
    throw new Error(response.error || 'Failed to update chamber config');
  }

  async getChamberWateringZones(id: string): Promise<any[]> {
    const response = await this.request<any[]>(`/chambers/${id}/watering-zones`);
    if (response.success && response.data) {
      return response.data;
    }
    throw new Error(response.error || 'Failed to get watering zones');
  }

  // Experiment endpoints
  async getExperiments(chamberId?: string): Promise<Experiment[]> {
    const params = chamberId ? `?chamber_id=${chamberId}` : '';
    const response = await this.request<Experiment[]>(`/experiments${params}`);
    if (response.success && response.data) {
      return response.data;
    }
    throw new Error(response.error || 'Failed to get experiments');
  }

  async getExperiment(id: string): Promise<Experiment> {
    const response = await this.request<Experiment>(`/experiments/${id}`);
    if (response.success && response.data) {
      return response.data;
    }
    throw new Error(response.error || 'Failed to get experiment');
  }

  async createExperiment(experiment: Partial<Experiment>): Promise<Experiment> {
    const response = await this.request<Experiment>('/experiments', {
      method: 'POST',
      body: JSON.stringify(experiment),
    });
    if (response.success && response.data) {
      return response.data;
    }
    throw new Error(response.error || 'Failed to create experiment');
  }

  async updateExperiment(id: string, experiment: Partial<Experiment>): Promise<Experiment> {
    const response = await this.request<Experiment>(`/experiments/${id}`, {
      method: 'PUT',
      body: JSON.stringify(experiment),
    });
    if (response.success && response.data) {
      return response.data;
    }
    throw new Error(response.error || 'Failed to update experiment');
  }

  async deleteExperiment(id: string): Promise<void> {
    const response = await this.request(`/experiments/${id}`, {
      method: 'DELETE',
    });
    if (!response.success) {
      throw new Error(response.error || 'Failed to delete experiment');
    }
  }

  // User management endpoints (admin only)
  async getUsers(): Promise<User[]> {
    const response = await this.request<User[]>('/users');
    if (response.success && response.data) {
      return response.data;
    }
    throw new Error(response.error || 'Failed to get users');
  }

  async getUser(id: string): Promise<User> {
    const response = await this.request<User>(`/users/${id}`);
    if (response.success && response.data) {
      return response.data;
    }
    throw new Error(response.error || 'Failed to get user');
  }

  async createUser(user: {
    username: string;
    password: string;
    name?: string;
    role?: 'admin' | 'user';
  }): Promise<User> {
    const response = await this.request<User>('/users', {
      method: 'POST',
      body: JSON.stringify(user),
    });
    if (response.success && response.data) {
      return response.data;
    }
    throw new Error(response.error || 'Failed to create user');
  }

  async updateUser(id: string, updates: Partial<User>): Promise<User> {
    const response = await this.request<User>(`/users/${id}`, {
      method: 'PUT',
      body: JSON.stringify(updates),
    });
    if (response.success && response.data) {
      return response.data;
    }
    throw new Error(response.error || 'Failed to update user');
  }

  async deactivateUser(id: string): Promise<void> {
    const response = await this.request(`/users/${id}`, {
      method: 'DELETE',
    });
    if (!response.success) {
      throw new Error(response.error || 'Failed to deactivate user');
    }
  }

  async activateUser(id: string): Promise<void> {
    const response = await this.request(`/users/${id}/activate`, {
      method: 'POST',
    });
    if (!response.success) {
      throw new Error(response.error || 'Failed to activate user');
    }
  }

  // API Token endpoints
  async getAPITokens(): Promise<APIToken[]> {
    const response = await this.request<APIToken[]>('/api-tokens');
    if (response.success && response.data) {
      return response.data;
    }
    throw new Error(response.error || 'Failed to get API tokens');
  }

  async createAPIToken(token: CreateAPITokenRequest): Promise<APIToken> {
    const response = await this.request<APIToken>('/api-tokens', {
      method: 'POST',
      body: JSON.stringify(token),
    });
    if (response.success && response.data) {
      return response.data;
    }
    throw new Error(response.error || 'Failed to create API token');
  }

  async revokeAPIToken(id: string): Promise<void> {
    const response = await this.request(`/api-tokens/${id}`, {
      method: 'DELETE',
    });
    if (!response.success) {
      throw new Error(response.error || 'Failed to revoke API token');
    }
  }

  // User Chamber Access endpoints
  async getUserChamberAccess(userId: string): Promise<Chamber[]> {
    const response = await this.request<Chamber[]>(`/users/${userId}/chambers`);
    if (response.success && response.data) {
      return response.data;
    }
    throw new Error(response.error || 'Failed to get user chamber access');
  }

  async setUserChamberAccess(userId: string, chamberIds: string[]): Promise<void> {
    const response = await this.request(`/users/${userId}/chambers`, {
      method: 'PUT',
      body: JSON.stringify({ chamber_ids: chamberIds }),
    });
    if (!response.success) {
      throw new Error(response.error || 'Failed to set user chamber access');
    }
  }

  async getAllUsersWithChamberAccess(): Promise<UserChamberAccess[]> {
    const response = await this.request<UserChamberAccess[]>('/users/chambers');
    if (response.success && response.data) {
      return response.data;
    }
    throw new Error(response.error || 'Failed to get users with chamber access');
  }

  async getMyChambers(): Promise<Chamber[]> {
    const response = await this.request<Chamber[]>('/me/chambers');
    if (response.success && response.data) {
      return response.data;
    }
    throw new Error(response.error || 'Failed to get my chambers');
  }

  async getMyChambersAccess(): Promise<UserChamberAccess> {
    const response = await this.request<UserChamberAccess>('/me/chambers/access');
    if (response.success && response.data) {
      return response.data;
    }
    throw new Error(response.error || 'Failed to get my chambers access');
  }
  
  // Health check
  async checkHealth(): Promise<{ status: string; time: string }> {
    const response = await this.request<{ status: string; time: string }>('/health');
    if (response.success && response.data) {
      return response.data;
    }
    throw new Error('Health check failed');
  }
}

export const api = new ApiClient();
import { apiClient } from '../../../shared/api/client';
import type { LoginCredentials, RegisterData, AuthResponse, User } from '../types';

export const authApi = {
  login: (credentials: LoginCredentials) =>
    apiClient.post<AuthResponse>('/auth/login', credentials),

  register: (data: RegisterData) =>
    apiClient.post<AuthResponse>('/auth/register', data),

  getMe: () => apiClient.get<User>('/auth/me'),

  logout: () => apiClient.post('/auth/logout'),

  refreshToken: () => apiClient.post<AuthResponse>('/auth/refresh'),
};


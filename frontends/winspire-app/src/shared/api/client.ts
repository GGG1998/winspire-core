import type { ApiResponse } from '../types';
import { supabase } from './supabase';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '';

/**
 * Determine which service should handle a request based on the endpoint path.
 * This is used for header-based routing at the ALB level.
 */
function getServiceForEndpoint(endpoint: string): 'tournament' | 'matchmaking' | null {
  // Matchmaking service endpoints
  if (endpoint.startsWith('/v1/matchmaking')) {
    return 'matchmaking';
  }
  if (endpoint.startsWith('/v1/matches') || endpoint.startsWith('/v1/lobbies') || endpoint.startsWith('/v1/brackets')) {
    return 'matchmaking';
  }

  // Game bundle streaming and admin endpoints (merged into matchmaking)
  // NOTE: Public game listing (/v1/games) removed - frontend uses Supabase directly
  if (endpoint.startsWith('/v1/g/') || endpoint.startsWith('/v1/admin/games')) {
    return 'matchmaking';
  }

  // Tournament service endpoints (hosts, registrations, and dynamic host paths)
  if (endpoint.startsWith('/v1/hosts')) {
    return 'tournament';
  }
  // Matches pattern like /v1/{hostId}/tournaments or /v1/{hostId}/registrations
  const dynamicHostPattern = /^\/v1\/[0-9a-f-]+\/(tournaments|registrations)/;
  if (dynamicHostPattern.test(endpoint)) {
    return 'tournament';
  }

  return null;
}

class ApiClient {
  private async getAuthToken(): Promise<string | null> {
    const { data: { session } } = await supabase.auth.getSession();
    return session?.access_token || null;
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<ApiResponse<T>> {
    const token = await this.getAuthToken();
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      ...(options.headers as Record<string, string>),
    };

    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }

    // Add X-Service header for ALB routing
    const service = getServiceForEndpoint(endpoint);
    if (service) {
      headers['X-Service'] = service;
    }

    try {
      const response = await fetch(`${API_BASE_URL}${endpoint}`, {
        ...options,
        headers,
      });

      const data = await response.json();

      if (!response.ok) {
        return {
          error: {
            code: data.code || 'UNKNOWN_ERROR',
            message: data.message || 'An error occurred',
            details: data.details,
          },
        };
      }

      return { data };
    } catch (error) {
      return {
        error: {
          code: 'NETWORK_ERROR',
          message: error instanceof Error ? error.message : 'Network error occurred',
        },
      };
    }
  }

  async get<T>(endpoint: string): Promise<ApiResponse<T>> {
    return this.request<T>(endpoint, { method: 'GET' });
  }

  async post<T>(endpoint: string, body?: unknown): Promise<ApiResponse<T>> {
    return this.request<T>(endpoint, {
      method: 'POST',
      body: body ? JSON.stringify(body) : undefined,
    });
  }

  async put<T>(endpoint: string, body?: unknown): Promise<ApiResponse<T>> {
    return this.request<T>(endpoint, {
      method: 'PUT',
      body: body ? JSON.stringify(body) : undefined,
    });
  }

  async patch<T>(endpoint: string, body?: unknown): Promise<ApiResponse<T>> {
    return this.request<T>(endpoint, {
      method: 'PATCH',
      body: body ? JSON.stringify(body) : undefined,
    });
  }

  async delete<T>(endpoint: string): Promise<ApiResponse<T>> {
    return this.request<T>(endpoint, { method: 'DELETE' });
  }
}

export const apiClient = new ApiClient();


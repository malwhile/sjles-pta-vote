import axios, { AxiosRequestConfig } from 'axios';

/**
 * Get authorization headers with JWT token
 * Used for admin endpoints that require authentication
 */
export function getAuthHeaders(): AxiosRequestConfig {
  const token = localStorage.getItem('adminToken');
  if (!token) {
    return {};
  }
  return {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  };
}

/**
 * Make a request to a protected admin endpoint
 * Automatically includes Authorization header
 */
export function apiRequest(method: string, url: string, data?: any) {
  const config = getAuthHeaders();
  config.method = method as any;
  config.url = url;
  if (data) {
    config.data = data;
  }
  return axios(config);
}

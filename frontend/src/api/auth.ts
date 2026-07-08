import { apiClient } from './client'

export const authApi = {
  register: (data: { username: string; password: string; email: string }) =>
    apiClient.post('/auth/register', data),
  login: (data: { username: string; password: string }) =>
    apiClient.post('/auth/login', data),
  refresh: (refresh_token: string) =>
    apiClient.post('/auth/refresh', { refresh_token }),
  getProfile: () => apiClient.get('/user/profile'),
}

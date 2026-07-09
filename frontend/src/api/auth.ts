import { apiClient } from './client'

export const authApi = {
  register: (data: { username: string; password: string; email: string }) =>
    apiClient.post('/auth/register', data),
  login: (data: { username: string; password: string }) =>
    apiClient.post('/auth/login', data),
  refresh: (refresh_token: string) =>
    apiClient.post('/auth/refresh', { refresh_token }),
  getProfile: () => apiClient.get('/user/profile'),
  updateProfile: (data: { username?: string; email?: string }) =>
    apiClient.put('/user/profile', data),
  uploadAvatar: (file: File) => {
    const formData = new FormData()
    formData.append('avatar', file)
    return apiClient.post('/user/avatar', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
  },
  getFavorites: (params?: { page?: number; page_size?: number }) =>
    apiClient.get('/user/favorites', { params }),
  getLikes: (params?: { page?: number; page_size?: number }) =>
    apiClient.get('/user/likes', { params }),
}

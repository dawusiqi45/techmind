import { apiClient } from './client'

export const incidentApi = {
  list: (params?: { status?: string; page?: number; page_size?: number }) =>
    apiClient.get('/admin/incidents', { params }),
  get: (id: string) => apiClient.get(`/admin/incidents/${id}`),
  resolve: (id: string) => apiClient.post(`/admin/incidents/${id}/resolve`),
}

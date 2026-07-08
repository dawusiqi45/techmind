import { apiClient } from './client'
export const alertApi = {
  list: (params?: { status?: string; page?: number; page_size?: number }) =>
    apiClient.get('/admin/alerts', { params }),
  get: (id: string) => apiClient.get(`/admin/alerts/${id}`),
  ack: (id: string) => apiClient.post(`/admin/alerts/${id}/ack`),
  diagnose: (id: string) => apiClient.post(`/admin/alerts/${id}/diagnose`),
}

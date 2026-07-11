import { apiClient } from './client'
export const opsApi = {
  listReports: (params?: { page?: number; page_size?: number }) =>
    apiClient.get('/admin/ops/reports', { params }),
  getReport: (id: string) => apiClient.get(`/admin/ops/reports/${id}`),
  getTimeline: (id: string) => apiClient.get(`/admin/ops/reports/${id}/timeline`),
  diagnose: (data: { service?: string; alert_name?: string }) =>
    apiClient.post('/admin/ops/diagnose', data),
}

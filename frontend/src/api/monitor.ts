import { apiClient } from './client'
export const monitorApi = {
  overview: () => apiClient.get('/admin/monitor/overview'),
  slowRequests: (params?: { page?: number; page_size?: number }) =>
    apiClient.get('/admin/monitor/slow-requests', { params }),
  errors: (params?: { page?: number; page_size?: number; source?: string }) =>
    apiClient.get('/admin/monitor/errors', { params }),
  queues: () => apiClient.get('/admin/monitor/queues'),
  aiCalls: (params?: { page?: number; page_size?: number; skill?: string }) =>
    apiClient.get('/admin/monitor/ai-calls', { params }),
}

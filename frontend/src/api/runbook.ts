import { apiClient } from './client'
export const runbookApi = {
  list: () => apiClient.get('/admin/runbooks'),
  create: (data: { title: string; content: string; alert_name?: string; service?: string }) =>
    apiClient.post('/admin/runbooks', data),
}

import { apiClient } from './client'
export const deploymentApi = {
  list: () => apiClient.get('/admin/deployment-changes'),
  create: (data: {
    service: string; namespace?: string; image: string;
    old_image?: string; changed_by?: string; source?: string
  }) => apiClient.post('/admin/deployment-changes', data),
}

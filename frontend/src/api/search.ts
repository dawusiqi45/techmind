import { apiClient } from './client'
export const searchApi = {
  search: (q: string) => apiClient.get('/search', { params: { q } }),
}

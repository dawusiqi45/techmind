import { apiClient } from './client'

export const articleApi = {
  list: (params?: { page?: number; page_size?: number; tag_id?: number }) =>
    apiClient.get('/articles', { params }),
  hot: () => apiClient.get('/articles/hot'),
  get: (id: string) => apiClient.get(`/articles/${id}`),
  create: (data: { title: string; content: string; tags?: string[] }) =>
    apiClient.post('/articles', data),
  update: (id: string, data: { title: string; content: string; tags?: string[] }) =>
    apiClient.put(`/articles/${id}`, data),
  remove: (id: string) => apiClient.delete(`/articles/${id}`),
  like: (id: string) => apiClient.post(`/articles/${id}/like`),
  favorite: (id: string) => apiClient.post(`/articles/${id}/favorite`),
  listComments: (id: string) => apiClient.get(`/articles/${id}/comments`),
  createComment: (id: string, data: { content: string; parent_id?: number }) =>
    apiClient.post(`/articles/${id}/comments`, data),
  tags: () => apiClient.get('/tags'),
}

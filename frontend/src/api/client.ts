import axios from 'axios'
import { tokenUtil } from '@/utils/token'

export const apiClient = axios.create({
  baseURL: '/api/v1',
  timeout: 15000,
})

apiClient.interceptors.request.use((config) => {
  const token = tokenUtil.getAccess()
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

let refreshing = false
let queue: Array<(token: string) => void> = []

apiClient.interceptors.response.use(
  (res) => res,
  async (err) => {
    const original = err.config
    if (err.response?.status !== 401 || original._retry) {
      return Promise.reject(err)
    }
    if (refreshing) {
      return new Promise((resolve) => {
        queue.push((token) => {
          original.headers.Authorization = `Bearer ${token}`
          resolve(apiClient(original))
        })
      })
    }
    refreshing = true
    original._retry = true
    try {
      const { data } = await axios.post('/api/v1/auth/refresh', {
        refresh_token: tokenUtil.getRefresh(),
      })
      const newToken = data.data.access_token
      tokenUtil.setAccess(newToken)
      queue.forEach((cb) => cb(newToken))
      queue = []
      original.headers.Authorization = `Bearer ${newToken}`
      return apiClient(original)
    } catch {
      tokenUtil.clear()
      return Promise.reject(err)
    } finally {
      refreshing = false
    }
  }
)

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
let queue: Array<{
  resolve: (token: string) => void
  reject: (error: unknown) => void
}> = []

apiClient.interceptors.response.use(
	(res) => {
		if (typeof res.data?.code === 'number' && res.data.code !== 1000) {
			return Promise.reject({ response: res, message: res.data.msg })
		}
		return res
	},
  async (err) => {
    const original = err.config
    const isAuthRequest = typeof original?.url === 'string' && original.url.includes('/auth/')
    if (err.response?.status !== 401 || original?._retry || isAuthRequest) {
      return Promise.reject(err)
    }
    if (refreshing) {
      return new Promise((resolve, reject) => {
        queue.push({
          resolve: (token) => {
            original.headers.Authorization = `Bearer ${token}`
            resolve(apiClient(original))
          },
          reject,
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
      queue.forEach((item) => item.resolve(newToken))
      queue = []
      original.headers.Authorization = `Bearer ${newToken}`
      return apiClient(original)
    } catch (refreshError) {
      tokenUtil.clear()
      queue.forEach((item) => item.reject(refreshError))
      queue = []
      return Promise.reject(refreshError)
    } finally {
      refreshing = false
    }
  }
)

const ACCESS_KEY = 'tm_access_token'
const REFRESH_KEY = 'tm_refresh_token'

export const tokenUtil = {
  getAccess: () => localStorage.getItem(ACCESS_KEY) ?? '',
  getRefresh: () => localStorage.getItem(REFRESH_KEY) ?? '',
  setAccess: (t: string) => localStorage.setItem(ACCESS_KEY, t),
  setRefresh: (t: string) => localStorage.setItem(REFRESH_KEY, t),
  clear: () => {
    localStorage.removeItem(ACCESS_KEY)
    localStorage.removeItem(REFRESH_KEY)
  },
}

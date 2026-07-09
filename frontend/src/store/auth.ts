import { create } from 'zustand'
import { tokenUtil } from '@/utils/token'
import { authApi } from '@/api/auth'

interface User { id: number; username: string; role: number; avatar: string }
interface AuthState {
  user: User | null
  initialized: boolean
  init: () => Promise<void>
  login: (accessToken: string, refreshToken: string, user: User) => void
  logout: () => void
  setUser: (user: User) => void
  isAdmin: () => boolean
}

const USER_KEY = 'tm_user'

function loadUser(): User | null {
  try {
    const raw = localStorage.getItem(USER_KEY)
    return raw ? JSON.parse(raw) : null
  } catch {
    return null
  }
}

function saveUser(user: User | null) {
  if (user) {
    localStorage.setItem(USER_KEY, JSON.stringify(user))
  } else {
    localStorage.removeItem(USER_KEY)
  }
}

export const useAuthStore = create<AuthState>((set, get) => ({
  user: loadUser(),
  initialized: false,
  init: async () => {
    const token = tokenUtil.getAccess()
    if (!token) {
      set({ initialized: true, user: null })
      return
    }
    if (get().user) {
      set({ initialized: true })
      return
    }
    try {
      const res = await authApi.getProfile()
      const user = res.data.data as User
      saveUser(user)
      set({ user, initialized: true })
    } catch {
      tokenUtil.clear()
      saveUser(null)
      set({ user: null, initialized: true })
    }
  },
  login: (accessToken, refreshToken, user) => {
    tokenUtil.setAccess(accessToken)
    tokenUtil.setRefresh(refreshToken)
    saveUser(user)
    set({ user })
  },
  logout: () => {
    tokenUtil.clear()
    saveUser(null)
    set({ user: null })
  },
  setUser: (user) => {
    saveUser(user)
    set({ user })
  },
  isAdmin: () => get().user?.role === 1,
}))

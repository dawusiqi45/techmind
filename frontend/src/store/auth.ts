import { create } from 'zustand'
import { tokenUtil } from '@/utils/token'

interface User { id: number; username: string; role: number; avatar: string }
interface AuthState {
  user: User | null
  login: (accessToken: string, refreshToken: string, user: User) => void
  logout: () => void
  isAdmin: () => boolean
}

export const useAuthStore = create<AuthState>((set, get) => ({
  user: null,
  login: (accessToken, refreshToken, user) => {
    tokenUtil.setAccess(accessToken)
    tokenUtil.setRefresh(refreshToken)
    set({ user })
  },
  logout: () => {
    tokenUtil.clear()
    set({ user: null })
  },
  isAdmin: () => get().user?.role === 1,
}))

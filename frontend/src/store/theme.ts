import { create } from 'zustand'
type Theme = 'forum' | 'admin'
interface ThemeState { theme: Theme; setTheme: (t: Theme) => void }
export const useThemeStore = create<ThemeState>((set) => ({
  theme: 'forum',
  setTheme: (theme) => set({ theme }),
}))

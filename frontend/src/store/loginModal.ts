import { create } from 'zustand'

interface LoginModalState {
  visible: boolean
  open: () => void
  close: () => void
}

export const useLoginModal = create<LoginModalState>((set) => ({
  visible: false,
  open: () => set({ visible: true }),
  close: () => set({ visible: false }),
}))

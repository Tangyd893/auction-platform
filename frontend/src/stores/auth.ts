import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import { authApi, type User } from '../api/rest'

interface AuthState {
  token: string | null
  user: User | null
  setAuth: (token: string, user: User) => void
  logout: () => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      token: null,
      user: null,
      setAuth: (token, user) => {
        set({ token, user })
      },
      logout: () => {
        set({ token: null, user: null })
        localStorage.removeItem('token')
        localStorage.removeItem('auth-storage')
      },
    }),
    {
      name: 'auth-storage',
    }
  )
)

export async function login(username: string, password: string) {
  const res = await authApi.login(username, password)
  const { token, user } = res.data
  const { setAuth } = useAuthStore.getState()
  setAuth(token, user)
  return { token, user }
}

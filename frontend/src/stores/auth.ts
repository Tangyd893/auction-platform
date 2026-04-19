import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import { authClient, setAuthToken } from '../grpc/client'

interface User {
  id: number
  username: string
  email: string
  role: string
}

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
        setAuthToken(token)
      },
      logout: () => {
        set({ token: null, user: null })
        setAuthToken(null)
      },
    }),
    {
      name: 'auth-storage',
      onRehydrateStorage: () => (state) => {
        if (state?.token) {
          setAuthToken(state.token)
        }
      },
    }
  )
)

// Login helper
export async function login(username: string, password: string) {
  const { token, user } = await authClient.login(username, password)
  const { setAuth } = useAuthStore.getState()
  setAuth(token, {
    id: user.getId(),
    username: user.getUsername(),
    email: user.getEmail(),
    role: user.getRole().toString(),
  })
  return { token, user }
}

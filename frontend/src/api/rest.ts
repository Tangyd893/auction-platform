import axios from 'axios'

const API_BASE = '/api'

const api = axios.create({
  baseURL: API_BASE,
  headers: { 'Content-Type': 'application/json' },
})

type PersistedAuthState = {
  state?: {
    token?: string | null
  }
}

function getStoredToken() {
  const token = localStorage.getItem('token')
  if (token) return token

  const persisted = localStorage.getItem('auth-storage')
  if (!persisted) return null

  try {
    const parsed = JSON.parse(persisted) as PersistedAuthState
    return parsed.state?.token ?? null
  } catch {
    return null
  }
}

api.interceptors.request.use((config) => {
  const token = getStoredToken()
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

// 响应拦截器：统一错误处理
api.interceptors.response.use(
  (res) => res,
  (err) => {
    if (err.response?.status === 401) {
      localStorage.removeItem('token')
      window.location.href = '/login'
    }
    return Promise.reject(err)
  }
)

export default api

// ============ Types ============
export interface User {
  id: number
  username: string
  email: string
  role: string
}

export interface Item {
  id: number
  title: string
  description: string
  image_url: string
  start_price: number   // 分
  current_price: number  // 分
  reserve_price: number
  bid_increment: number
  seller_id: number
  status: 'draft' | 'listed' | 'sold' | 'unsold' | 'cancelled'
  start_time: string
  end_time: string
  created_at: string
}

export interface Bid {
  id: number
  item_id: number
  buyer_id: number
  amount: number   // 分
  status: 'active' | 'outbid' | 'winning'
  created_at: string
}

export interface Order {
  id: number
  item_id: number
  seller_id: number
  buyer_id: number
  final_price: number  // 分
  status: 'pending' | 'paid' | 'shipped' | 'completed' | 'cancelled'
  created_at: string
  paid_at?: string
}

// ============ Auth API ============
export const authApi = {
  login: (username: string, password: string) =>
    api.post<{ token: string; user: User }>('/auth/login', { username, password }),

  register: (data: { username: string; password: string; email: string; role: string }) =>
    api.post<User>('/auth/register', data),
}

// ============ Item API ============
export const itemApi = {
  list: (params?: { status?: string; keyword?: string; page?: number; pageSize?: number }) =>
    api.get<{ items: Item[]; total: number }>('/items', { params }),

  get: (id: number) =>
    api.get<Item>(`/items/${id}`).then((r) => r.data),

  create: (data: {
    title: string
    description: string
    imageUrl: string
    startPrice: number
    reservePrice: number
    bidIncrement: number
    startTime: number
    endTime: number
  }) => api.post<Item>('/items', data),

  myItems: (params?: { status?: string }) =>
    api.get<{ items: Item[]; total: number }>('/my-items', { params }),

  cancel: (id: number) =>
    api.delete<Item>(`/items/${id}`),
}

// ============ Bid API ============
export const bidApi = {
  place: (itemId: number, amount: number) =>
    api.post<{ bidId: number; currentPrice: number; isWinning: boolean }>('/bids', {
      itemId,
      amount,
    }),

  getByItem: (itemId: number) =>
    api.get<{ bids: Bid[]; highestPrice: number; highestBidderId: number }>(
      `/items/${itemId}/bids`
    ),

  myBids: () =>
    api.get<{ bids: Bid[] }>('/my-bids'),
}

// ============ Order API ============
export const orderApi = {
  create: (itemId: number) =>
    api.post<Order>('/orders', { itemId }),

  get: (id: number) =>
    api.get<Order>(`/orders/${id}`).then((r) => r.data),

  list: (params?: { page?: number; pageSize?: number }) =>
    api.get<{ orders: Order[]; total: number }>('/orders', { params }),

  updateStatus: (id: number, status: string) =>
    api.put<Order>(`/orders/${id}/status`, { status }),
}

// ============ User API ============
export const userApi = {
  get: (id: number) =>
    api.get<User>(`/users/${id}`),
}

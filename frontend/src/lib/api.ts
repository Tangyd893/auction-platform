import axios from 'axios'

const API_BASE = import.meta.env.VITE_API_BASE_URL || '/api'

const api = axios.create({
  baseURL: API_BASE,
  headers: {
    'Content-Type': 'application/json',
  },
})

// 请求拦截器：注入 token
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// 响应拦截器：统一错误处理
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export default api

// ============ Auth API ============
export const authApi = {
  login: (username: string, password: string) =>
    api.post('/auth/login', { username, password }),
  register: (data: { username: string; password: string; email: string; role: string }) =>
    api.post('/auth/register', data),
}

// ============ Item API ============
export const itemApi = {
  list: (params?: { status?: string; page?: number; pageSize?: number; keyword?: string }) =>
    api.get('/items', { params }),
  get: (id: number) => api.get(`/items/${id}`),
  create: (data: {
    title: string
    description: string
    imageUrl: string
    startPrice: number
    reservePrice: number
    bidIncrement: number
    startTime: number
    endTime: number
  }) => api.post('/items', data),
  myItems: (params?: { status?: string }) => api.get('/items/my', { params }),
}

// ============ Bid API ============
export const bidApi = {
  place: (itemId: number, amount: number) =>
    api.post('/bids', { itemId, amount }),
  getByItem: (itemId: number) => api.get(`/items/${itemId}/bids`),
  myBids: () => api.get('/bids/my'),
}

// ============ Order API ============
export const orderApi = {
  create: (itemId: number) => api.post('/orders', { itemId }),
  get: (id: number) => api.get(`/orders/${id}`),
  list: (params?: { page?: number; pageSize?: number }) =>
    api.get('/orders', { params }),
  updateStatus: (id: number, status: string) =>
    api.patch(`/orders/${id}/status`, { status }),
}

// ============ User API ============
export const userApi = {
  list: (params?: { page?: number; pageSize?: number }) =>
    api.get('/users', { params }),
  get: (id: number) => api.get(`/users/${id}`),
}

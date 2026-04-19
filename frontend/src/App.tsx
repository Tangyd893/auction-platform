import { Routes, Route, Navigate } from 'react-router-dom'
import { useAuthStore } from './stores/auth'
import Layout from './components/Layout'
import LoginPage from './pages/LoginPage'
import RegisterPage from './pages/RegisterPage'
import HomePage from './pages/HomePage'
import ItemListPage from './pages/ItemListPage'
import ItemDetailPage from './pages/ItemDetailPage'
import MyItemsPage from './pages/MyItemsPage'
import CreateItemPage from './pages/CreateItemPage'
import MyBidsPage from './pages/MyBidsPage'
import OrdersPage from './pages/OrdersPage'

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { token } = useAuthStore()
  if (!token) {
    return <Navigate to="/login" replace />
  }
  return <>{children}</>
}

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/register" element={<RegisterPage />} />
      <Route path="/" element={<Layout />}>
        <Route index element={<HomePage />} />
        <Route path="items" element={<ItemListPage />} />
        <Route path="items/:id" element={<ItemDetailPage />} />
        <Route
          path="my-items"
          element={
            <ProtectedRoute>
              <MyItemsPage />
            </ProtectedRoute>
          }
        />
        <Route
          path="create-item"
          element={
            <ProtectedRoute>
              <CreateItemPage />
            </ProtectedRoute>
          }
        />
        <Route
          path="my-bids"
          element={
            <ProtectedRoute>
              <MyBidsPage />
            </ProtectedRoute>
          }
        />
        <Route
          path="orders"
          element={
            <ProtectedRoute>
              <OrdersPage />
            </ProtectedRoute>
          }
        />
      </Route>
    </Routes>
  )
}

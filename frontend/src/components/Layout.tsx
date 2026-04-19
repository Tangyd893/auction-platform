import { Outlet, Link, useNavigate } from 'react-router-dom'
import { useAuthStore } from '../stores/auth'

export default function Layout() {
  const { token, user, logout } = useAuthStore()
  const navigate = useNavigate()

  const handleLogout = () => {
    logout()
    navigate('/login')
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* 顶部导航 */}
      <header className="bg-white shadow-sm">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between h-16">
            <div className="flex items-center">
              <Link to="/" className="text-xl font-bold text-indigo-600">
                🔨 拍卖平台
              </Link>
              <nav className="ml-10 flex space-x-4">
                <Link to="/items" className="text-gray-700 hover:text-indigo-600 px-3 py-2">
                  拍品列表
                </Link>
                {token && (
                  <>
                    <Link to="/my-items" className="text-gray-700 hover:text-indigo-600 px-3 py-2">
                      我的拍品
                    </Link>
                    <Link to="/my-bids" className="text-gray-700 hover:text-indigo-600 px-3 py-2">
                      我的出价
                    </Link>
                    <Link to="/orders" className="text-gray-700 hover:text-indigo-600 px-3 py-2">
                      订单
                    </Link>
                  </>
                )}
              </nav>
            </div>
            <div className="flex items-center">
              {token ? (
                <div className="flex items-center space-x-4">
                  <span className="text-sm text-gray-700">
                    {user?.username}
                  </span>
                  <button
                    onClick={handleLogout}
                    className="text-sm text-gray-500 hover:text-red-600"
                  >
                    退出
                  </button>
                </div>
              ) : (
                <div className="space-x-4">
                  <Link to="/login" className="text-sm text-indigo-600 hover:text-indigo-800">
                    登录
                  </Link>
                  <Link
                    to="/register"
                    className="text-sm bg-indigo-600 text-white px-4 py-2 rounded hover:bg-indigo-700"
                  >
                    注册
                  </Link>
                </div>
              )}
            </div>
          </div>
        </div>
      </header>

      {/* 主内容 */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <Outlet />
      </main>
    </div>
  )
}

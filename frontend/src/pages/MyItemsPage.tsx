import { Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { itemClient, proto } from '../grpc/client'
import { useAuthStore } from '../stores/auth'
import { Plus } from 'lucide-react'

export default function MyItemsPage() {
  const { user } = useAuthStore()

  const { data, isLoading } = useQuery({
    queryKey: ['my-items'],
    queryFn: () => itemClient.myItems(user?.id || 0),
  })

  const items = data?.items || []

  const formatPrice = (cents: number) => `¥${(cents / 100).toFixed(2)}`

  const statusText: Record<number, string> = {
    [proto.ItemStatus.DRAFT]: '草稿',
    [proto.ItemStatus.LISTED]: '正在拍卖',
    [proto.ItemStatus.SOLD]: '已售出',
    [proto.ItemStatus.UNSOLD]: '流拍',
    [proto.ItemStatus.CANCELLED]: '已取消',
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold">我的拍品</h1>
        <Link
          to="/create-item"
          className="bg-indigo-600 text-white px-4 py-2 rounded-lg hover:bg-indigo-700 flex items-center gap-2"
        >
          <Plus size={18} />
          发布新拍品
        </Link>
      </div>

      {isLoading ? (
        <div className="text-center py-12 text-gray-500">加载中...</div>
      ) : items.length === 0 ? (
        <div className="text-center py-12 bg-white rounded-xl shadow-sm">
          <p className="text-gray-500 mb-4">您还没有发布任何拍品</p>
          <Link
            to="/create-item"
            className="text-indigo-600 hover:text-indigo-800"
          >
            立即发布
          </Link>
        </div>
      ) : (
        <div className="grid grid-cols-3 gap-6">
          {items.map((item: proto.Item) => (
            <div
              key={item.getId()}
              className="bg-white rounded-xl overflow-hidden shadow-sm"
            >
              <img
                src={item.getImageUrl() || 'https://via.placeholder.com/300x200'}
                alt={item.getTitle()}
                className="w-full h-40 object-cover"
              />
              <div className="p-4">
                <h3 className="font-semibold text-lg mb-2 truncate">
                  {item.getTitle()}
                </h3>
                <div className="space-y-1 text-sm mb-3">
                  <div className="flex justify-between">
                    <span className="text-gray-500">当前价</span>
                    <span className="text-indigo-600 font-bold">
                      {formatPrice(item.getCurrentPrice())}
                    </span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-gray-500">状态</span>
                    <span
                      className={
                        item.getStatus() === proto.ItemStatus.LISTED
                          ? 'text-green-600'
                          : 'text-gray-600'
                      }
                    >
                      {statusText[item.getStatus()]}
                    </span>
                  </div>
                </div>
                <Link
                  to={`/items/${item.getId()}`}
                  className="block w-full text-center bg-indigo-50 text-indigo-600 py-2 rounded-lg hover:bg-indigo-100"
                >
                  查看详情
                </Link>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}

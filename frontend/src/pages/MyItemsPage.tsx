import { Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { itemApi } from '../api/rest'
import { Plus } from 'lucide-react'

const STATUS_LABELS: Record<string, string> = {
  draft: '草稿',
  listed: '正在拍卖',
  sold: '已售出',
  unsold: '流拍',
  cancelled: '已取消',
}

export default function MyItemsPage() {
  const { data, isLoading } = useQuery({
    queryKey: ['my-items'],
    queryFn: async () => { const r = await itemApi.myItems(); return r.data },
  })

  const items = data?.items || []
  const formatPrice = (cents: number) => `¥${(cents / 100).toFixed(2)}`

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold">我的拍品</h1>
        <Link to="/create-item" className="bg-indigo-600 text-white px-4 py-2 rounded-lg hover:bg-indigo-700 flex items-center gap-2">
          <Plus size={18} />发布新拍品
        </Link>
      </div>
      {isLoading ? (
        <div className="text-center py-12 text-gray-500">加载中...</div>
      ) : items.length === 0 ? (
        <div className="text-center py-12 bg-white rounded-xl shadow-sm">
          <p className="text-gray-500 mb-4">您还没有发布任何拍品</p>
          <Link to="/create-item" className="text-indigo-600 hover:text-indigo-800">立即发布</Link>
        </div>
      ) : (
        <div className="grid grid-cols-3 gap-6">
          {items.map((item: any) => (
            <div key={item.id} className="bg-white rounded-xl overflow-hidden shadow-sm">
              <img src={item.image_url || 'https://via.placeholder.com/300x200'} alt={item.title}
                className="w-full h-40 object-cover" />
              <div className="p-4">
                <h3 className="font-semibold text-lg mb-2 truncate">{item.title}</h3>
                <div className="space-y-1 text-sm mb-3">
                  <div className="flex justify-between">
                    <span className="text-gray-500">当前价</span>
                    <span className="text-indigo-600 font-bold">{formatPrice(item.current_price)}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-gray-500">状态</span>
                    <span className={item.status === 'listed' ? 'text-green-600' : 'text-gray-600'}>
                      {STATUS_LABELS[item.status] || item.status}
                    </span>
                  </div>
                </div>
                <Link to={`/items/${item.id}`}
                  className="block w-full text-center bg-indigo-50 text-indigo-600 py-2 rounded-lg hover:bg-indigo-100">
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

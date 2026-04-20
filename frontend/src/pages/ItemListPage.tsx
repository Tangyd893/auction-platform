import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { itemApi } from '../api/rest'
import { Search } from 'lucide-react'

export default function ItemListPage() {
  const [status, setStatus] = useState('listed')
  const [keyword, setKeyword] = useState('')
  const [page, setPage] = useState(1)

  const { data, isLoading } = useQuery({
    queryKey: ['items', status, keyword, page],
    queryFn: () => itemApi.list({ status: status === 'all' ? undefined : status, keyword: keyword || undefined, page, pageSize: 12 }),
  })

  const items = data?.data?.items || []
  const total = data?.data?.total || 0
  const totalPages = Math.ceil(total / 12)

  const formatPrice = (cents: number) => `¥${(cents / 100).toFixed(2)}`

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold">拍品列表</h1>
        <Link to="/create-item" className="bg-indigo-600 text-white px-4 py-2 rounded-lg hover:bg-indigo-700">
          发布拍品
        </Link>
      </div>

      <div className="bg-white rounded-xl p-4 mb-6 shadow-sm">
        <div className="flex gap-4">
          <div className="flex-1 relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" size={20} />
            <input type="text" placeholder="搜索拍品..." value={keyword}
              onChange={(e) => { setKeyword(e.target.value); setPage(1) }}
              className="w-full pl-10 pr-4 py-2 border rounded-lg focus:ring-2 focus:ring-indigo-500" />
          </div>
          <select value={status} onChange={(e) => { setStatus(e.target.value); setPage(1) }}
            className="px-4 py-2 border rounded-lg focus:ring-2 focus:ring-indigo-500">
            <option value="listed">正在拍卖</option>
            <option value="draft">草稿</option>
            <option value="sold">已售出</option>
            <option value="unsold">流拍</option>
          </select>
        </div>
      </div>

      {isLoading ? (
        <div className="text-center py-12 text-gray-700">加载中...</div>
      ) : items.length === 0 ? (
        <div className="text-center py-12 text-gray-700">暂无拍品</div>
      ) : (
        <>
          <div className="grid grid-cols-4 gap-6">
            {items.map((item) => (
              <Link key={item.id} to={`/items/${item.id}`}
                className="bg-white rounded-xl overflow-hidden shadow-sm hover:shadow-md transition">
                <img src={item.image_url || 'https://via.placeholder.com/300x200'} alt={item.title}
                  className="w-full h-40 object-cover" />
                <div className="p-4">
                  <h3 className="font-semibold text-lg mb-2 truncate">{item.title}</h3>
                  <div className="space-y-1 text-sm">
                    <div className="flex justify-between">
                      <span className="text-gray-700">当前价</span>
                      <span className="text-indigo-600 font-bold">{formatPrice(item.current_price)}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-gray-600">起拍价</span>
                      <span className="text-gray-900 font-medium">{formatPrice(item.start_price)}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-gray-600">加价幅度</span>
                      <span className="text-gray-900 font-medium">{formatPrice(item.bid_increment)}</span>
                    </div>
                  </div>
                </div>
              </Link>
            ))}
          </div>

          {totalPages > 1 && (
            <div className="flex justify-center gap-2 mt-8">
              <button onClick={() => setPage((p) => Math.max(1, p - 1))} disabled={page === 1}
                className="px-4 py-2 border rounded-lg disabled:opacity-50 hover:bg-gray-50">
                上一页
              </button>
              <span className="px-4 py-2">第 {page} / {totalPages} 页，共 {total} 条</span>
              <button onClick={() => setPage((p) => Math.min(totalPages, p + 1))} disabled={page === totalPages}
                className="px-4 py-2 border rounded-lg disabled:opacity-50 hover:bg-gray-50">
                下一页
              </button>
            </div>
          )}
        </>
      )}
    </div>
  )
}

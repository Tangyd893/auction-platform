import { Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { bidApi } from '../api/rest'

export default function MyBidsPage() {
  const { data, isLoading } = useQuery({
    queryKey: ['my-bids'],
    queryFn: async () => { const r = await bidApi.myBids(); return r.data },
  })

  const bids = data?.bids || []
  const formatPrice = (cents: number) => `¥${(cents / 100).toFixed(2)}`

  const statusText: Record<string, { text: string; color: string }> = {
    active: { text: '领先', color: 'text-green-600 bg-green-50' },
    outbid: { text: '被超越', color: 'text-red-600 bg-red-50' },
    winning: { text: '最终赢家', color: 'text-indigo-600 bg-indigo-50' },
  }

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">我的出价</h1>
      {isLoading ? (
        <div className="text-center py-12 text-gray-700">加载中...</div>
      ) : bids.length === 0 ? (
        <div className="text-center py-12 bg-white rounded-xl shadow-sm">
          <p className="text-gray-700">您还没有出价记录</p>
          <Link to="/items" className="text-indigo-600 hover:text-indigo-800 mt-4 inline-block">去竞拍</Link>
        </div>
      ) : (
        <div className="bg-white rounded-xl shadow-sm overflow-hidden">
          <table className="w-full">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-700 uppercase">拍品ID</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-700 uppercase">出价金额</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-700 uppercase">状态</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-700 uppercase">出价时间</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {bids.map((bid: any) => (
                <tr key={bid.id} className="hover:bg-gray-50">
                  <td className="px-6 py-4">
                    <Link to={`/items/${bid.item_id}`} className="text-indigo-600 hover:text-indigo-800">#{bid.item_id}</Link>
                  </td>
                  <td className="px-6 py-4 font-semibold">{formatPrice(bid.amount)}</td>
                  <td className="px-6 py-4">
                    <span className={`px-2 py-1 rounded-full text-xs ${statusText[bid.status]?.color || 'text-gray-600 bg-gray-50'}`}>
                      {statusText[bid.status]?.text || bid.status}
                    </span>
                  </td>
                  <td className="px-6 py-4 text-gray-700">{new Date(bid.created_at).toLocaleString()}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}

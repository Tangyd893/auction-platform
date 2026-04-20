import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { orderApi } from '../api/rest'
import { useAuthStore } from '../stores/auth'

const STATUS_LABELS: Record<string, { text: string; color: string }> = {
  pending: { text: '待支付', color: 'text-yellow-600 bg-yellow-50' },
  paid: { text: '已支付', color: 'text-blue-600 bg-blue-50' },
  shipped: { text: '已发货', color: 'text-indigo-600 bg-indigo-50' },
  completed: { text: '已完成', color: 'text-green-600 bg-green-50' },
  cancelled: { text: '已取消', color: 'text-gray-600 bg-gray-50' },
}

export default function OrdersPage() {
  const { user } = useAuthStore()
  const queryClient = useQueryClient()

  const { data, isLoading } = useQuery({
    queryKey: ['orders'],
    queryFn: async () => { const r = await orderApi.list(); return r.data },
  })

  const orders = data?.orders || []
  const formatPrice = (cents: number) => `¥${(cents / 100).toFixed(2)}`

  const updateStatusMutation = useMutation({
    mutationFn: ({ id, status }: { id: number; status: string }) => orderApi.updateStatus(id, status),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['orders'] }),
  })

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">我的订单</h1>
      {isLoading ? (
        <div className="text-center py-12 text-gray-700">加载中...</div>
      ) : orders.length === 0 ? (
        <div className="text-center py-12 bg-white rounded-xl shadow-sm">
          <p className="text-gray-700">暂无订单</p>
        </div>
      ) : (
        <div className="bg-white rounded-xl shadow-sm overflow-hidden">
          <table className="w-full">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-700 uppercase">订单ID</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-700 uppercase">拍品ID</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-700 uppercase">成交价</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-700 uppercase">状态</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-700 uppercase">创建时间</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-700 uppercase">操作</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {orders.map((order: any) => (
                <tr key={order.id} className="hover:bg-gray-50">
                  <td className="px-6 py-4 font-medium">#{order.id}</td>
                  <td className="px-6 py-4"><span className="text-indigo-600">#{order.item_id}</span></td>
                  <td className="px-6 py-4 font-semibold">{formatPrice(order.final_price)}</td>
                  <td className="px-6 py-4">
                    <span className={`px-2 py-1 rounded-full text-xs ${STATUS_LABELS[order.status]?.color || 'text-gray-600 bg-gray-50'}`}>
                      {STATUS_LABELS[order.status]?.text || order.status}
                    </span>
                  </td>
                  <td className="px-6 py-4 text-gray-700">{new Date(order.created_at).toLocaleString()}</td>
                  <td className="px-6 py-4">
                    {order.status === 'pending' && user?.role === 'buyer' && (
                      <button onClick={() => updateStatusMutation.mutate({ id: order.id, status: 'paid' })}
                        disabled={updateStatusMutation.isPending}
                        className="text-sm bg-indigo-600 text-white px-3 py-1 rounded hover:bg-indigo-700 disabled:opacity-50">
                        支付
                      </button>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { orderClient, proto } from '../grpc/client'
import { useAuthStore } from '../stores/auth'

export default function OrdersPage() {
  const { user } = useAuthStore()
  const queryClient = useQueryClient()

  const { data, isLoading } = useQuery({
    queryKey: ['orders'],
    queryFn: () => orderClient.list(user?.id || 0),
  })

  const orders = data?.orders || []

  const formatPrice = (cents: number) => `¥${(cents / 100).toFixed(2)}`

  const statusText: Record<number, { text: string; color: string }> = {
    0: { text: '待支付', color: 'text-yellow-600 bg-yellow-50' },
    1: { text: '已支付', color: 'text-blue-600 bg-blue-50' },
    2: { text: '已发货', color: 'text-indigo-600 bg-indigo-50' },
    3: { text: '已完成', color: 'text-green-600 bg-green-50' },
    4: { text: '已取消', color: 'text-gray-600 bg-gray-50' },
  }

  const updateStatusMutation = useMutation({
    mutationFn: ({ id, status }: { id: number; status: number }) =>
      orderClient.updateStatus(id, status),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['orders'] })
    },
  })

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">我的订单</h1>

      {isLoading ? (
        <div className="text-center py-12 text-gray-500">加载中...</div>
      ) : orders.length === 0 ? (
        <div className="text-center py-12 bg-white rounded-xl shadow-sm">
          <p className="text-gray-500">暂无订单</p>
        </div>
      ) : (
        <div className="bg-white rounded-xl shadow-sm overflow-hidden">
          <table className="w-full">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                  订单ID
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                  拍品ID
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                  成交价
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                  状态
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                  创建时间
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                  操作
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {orders.map((order: proto.Order) => (
                <tr key={order.getId()} className="hover:bg-gray-50">
                  <td className="px-6 py-4 font-medium">#{order.getId()}</td>
                  <td className="px-6 py-4">
                    <span className="text-indigo-600">#{order.getItemId()}</span>
                  </td>
                  <td className="px-6 py-4 font-semibold">
                    {formatPrice(order.getFinalPrice())}
                  </td>
                  <td className="px-6 py-4">
                    <span
                      className={`px-2 py-1 rounded-full text-xs ${
                        statusText[order.getStatus()]?.color || 'text-gray-600 bg-gray-50'
                      }`}
                    >
                      {statusText[order.getStatus()]?.text || order.getStatus()}
                    </span>
                  </td>
                  <td className="px-6 py-4 text-gray-500">
                    {new Date(order.getCreatedAt() * 1000).toLocaleString()}
                  </td>
                  <td className="px-6 py-4">
                    {order.getStatus() === 0 && user?.role === 'buyer' && (
                      <button
                        onClick={() =>
                          updateStatusMutation.mutate({
                            id: order.getId(),
                            status: 1, // paid
                          })
                        }
                        disabled={updateStatusMutation.isPending}
                        className="text-sm bg-indigo-600 text-white px-3 py-1 rounded hover:bg-indigo-700 disabled:opacity-50"
                      >
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

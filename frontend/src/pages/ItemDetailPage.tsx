import { useState, useEffect } from 'react'
import { useParams } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { itemClient, bidClient, proto } from '../grpc/client'
import { useAuthStore } from '../stores/auth'
import { Clock, User, Gavel } from 'lucide-react'

export default function ItemDetailPage() {
  const { id } = useParams()
  const { token } = useAuthStore()
  const queryClient = useQueryClient()
  const [bidAmount, setBidAmount] = useState('')
  const [streamBids, setStreamBids] = useState<any>(null)

  const itemId = Number(id)

  const { data: itemData, isLoading } = useQuery({
    queryKey: ['item', itemId],
    queryFn: () => itemClient.get(itemId),
    refetchInterval: 5000,
  })

  const { data: bidsData } = useQuery({
    queryKey: ['bids', itemId],
    queryFn: () => bidClient.getByItem(itemId),
    refetchInterval: 3000,
  })

  const item = itemData
  const bids = bidsData?.bids || []

  const bidMutation = useMutation({
    mutationFn: (amount: number) => bidClient.place(itemId, amount),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['item', itemId] })
      queryClient.invalidateQueries({ queryKey: ['bids', itemId] })
      setBidAmount('')
    },
  })

  // 启动 Server Streaming 订阅出价更新
  useEffect(() => {
    if (!itemId) return

    const stream = bidClient.streamBids(itemId, (bid, currentPrice, totalBids) => {
      queryClient.setQueryData(['item', itemId], (old: any) => {
        if (!old) return old
        // 更新当前价
        const updated = old.clone()
        updated.setCurrentPrice(currentPrice)
        return updated
      })
      console.log(`出价更新: ¥${currentPrice / 100}, 共 ${totalBids} 次出价`)
    })

    setStreamBids(stream)

    return () => {
      if (stream) {
        stream.cancel()
      }
    }
  }, [itemId])

  const handleBid = () => {
    if (!token) {
      window.location.href = '/login'
      return
    }
    const amount = Math.floor(parseFloat(bidAmount) * 100)
    if (isNaN(amount) || amount <= 0) return
    bidMutation.mutate(amount)
  }

  const formatPrice = (cents: number) => `¥${(cents / 100).toFixed(2)}`
  const minBid = item ? item.getCurrentPrice() + item.getBidIncrement() : 0

  if (isLoading) return <div className="text-center py-12">加载中...</div>
  if (!item) return <div className="text-center py-12">拍品不存在</div>

  const statusText: Record<number, string> = {
    [proto.ItemStatus.DRAFT]: '草稿',
    [proto.ItemStatus.LISTED]: '正在拍卖',
    [proto.ItemStatus.SOLD]: '已售出',
    [proto.ItemStatus.UNSOLD]: '流拍',
    [proto.ItemStatus.CANCELLED]: '已取消',
  }

  return (
    <div className="grid grid-cols-2 gap-8">
      {/* 左侧：图片 */}
      <div className="bg-white rounded-xl overflow-hidden shadow-sm">
        <img
          src={item.getImageUrl() || 'https://via.placeholder.com/600x400'}
          alt={item.getTitle()}
          className="w-full h-96 object-cover"
        />
      </div>

      {/* 右侧：信息 */}
      <div className="space-y-6">
        <div>
          <h1 className="text-3xl font-bold mb-2">{item.getTitle()}</h1>
          <span
            className={`inline-block px-3 py-1 rounded-full text-sm ${
              item.getStatus() === proto.ItemStatus.LISTED
                ? 'bg-green-100 text-green-800'
                : 'bg-gray-100 text-gray-800'
            }`}
          >
            {statusText[item.getStatus()]}
          </span>
        </div>

        <div className="bg-indigo-50 rounded-xl p-6">
          <div className="grid grid-cols-2 gap-4">
            <div>
              <p className="text-sm text-gray-600">当前价</p>
              <p className="text-3xl font-bold text-indigo-600">
                {formatPrice(item.getCurrentPrice())}
              </p>
            </div>
            <div>
              <p className="text-sm text-gray-600">起拍价</p>
              <p className="text-2xl text-gray-700">
                {formatPrice(item.getStartPrice())}
              </p>
            </div>
            <div>
              <p className="text-sm text-gray-600">加价幅度</p>
              <p className="text-xl text-gray-700">
                {formatPrice(item.getBidIncrement())}
              </p>
            </div>
            <div>
              <p className="text-sm text-gray-600">出价次数</p>
              <p className="text-xl text-gray-700">{bids.length}</p>
            </div>
          </div>
        </div>

        <div className="space-y-2">
          <div className="flex items-center gap-2 text-gray-600">
            <Clock size={18} />
            <span>
              截止时间：{new Date(item.getEndTime() * 1000).toLocaleString()}
            </span>
          </div>
          <div className="flex items-center gap-2 text-gray-600">
            <User size={18} />
            <span>卖家ID：{item.getSellerId()}</span>
          </div>
        </div>

        <p className="text-gray-700">{item.getDescription()}</p>

        {/* 出价区域 */}
        {item.getStatus() === proto.ItemStatus.LISTED && (
          <div className="bg-white rounded-xl p-6 shadow-sm">
            <div className="flex gap-4">
              <input
                type="number"
                value={bidAmount}
                onChange={(e) => setBidAmount(e.target.value)}
                placeholder={`最低出价 ${formatPrice(minBid)}`}
                className="flex-1 px-4 py-2 border rounded-lg focus:ring-2 focus:ring-indigo-500"
              />
              <button
                onClick={handleBid}
                disabled={bidMutation.isPending}
                className="bg-indigo-600 text-white px-6 py-2 rounded-lg hover:bg-indigo-700 disabled:opacity-50 flex items-center gap-2"
              >
                <Gavel size={18} />
                {bidMutation.isPending ? '出价中...' : '出价'}
              </button>
            </div>
            {bidMutation.isError && (
              <p className="text-red-600 text-sm mt-2">出价失败，请重试</p>
            )}
          </div>
        )}

        {/* 出价记录 */}
        <div className="bg-white rounded-xl p-6 shadow-sm">
          <h3 className="text-lg font-semibold mb-4">出价记录</h3>
          {bids.length === 0 ? (
            <p className="text-gray-500 text-center py-4">暂无出价</p>
          ) : (
            <div className="space-y-2">
              {bids.slice(0, 5).map((bid: proto.Bid, index: number) => (
                <div
                  key={bid.getId()}
                  className={`flex justify-between items-center p-3 rounded-lg ${
                    index === 0 ? 'bg-green-50' : 'bg-gray-50'
                  }`}
                >
                  <span className="text-gray-700">买家 {bid.getBuyerId()}</span>
                  <span
                    className={`font-semibold ${
                      index === 0 ? 'text-green-600' : 'text-gray-700'
                    }`}
                  >
                    {formatPrice(bid.getAmount())}
                  </span>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

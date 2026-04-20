import { Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { itemApi } from '../api/rest'
import { ArrowRight, TrendingUp, Shield, Clock } from 'lucide-react'

export default function HomePage() {
  const { data } = useQuery({
    queryKey: ['items', 'listed', 1],
    queryFn: () => itemApi.list({ status: 'listed', pageSize: 6 }),
  })

  const items = data?.data?.items || []

  const formatPrice = (cents: number) => `¥${(cents / 100).toFixed(2)}`

  return (
    <div>
      <section className="bg-gradient-to-r from-indigo-600 to-purple-600 text-white rounded-2xl p-12 mb-8">
        <h1 className="text-4xl font-bold mb-4">在线拍卖平台</h1>
        <p className="text-xl text-indigo-100 mb-6">
          发现独特商品，参与实时竞价，透明公正的拍卖体验
        </p>
        <div className="flex gap-4">
          <Link to="/items"
            className="bg-white text-indigo-600 px-6 py-3 rounded-lg font-semibold hover:bg-indigo-50 flex items-center gap-2">
            浏览拍品 <ArrowRight size={20} />
          </Link>
          <Link to="/register"
            className="border-2 border-white text-white px-6 py-3 rounded-lg font-semibold hover:bg-white/10">
            立即注册
          </Link>
        </div>
      </section>

      <section className="grid grid-cols-3 gap-6 mb-8">
        <div className="bg-white rounded-xl p-6 shadow-sm">
          <TrendingUp className="w-10 h-10 text-indigo-600 mb-4" />
          <h3 className="text-lg font-semibold mb-2">实时竞价</h3>
          <p className="text-gray-600">毫秒级更新，公平透明的竞价环境</p>
        </div>
        <div className="bg-white rounded-xl p-6 shadow-sm">
          <Clock className="w-10 h-10 text-indigo-600 mb-4" />
          <h3 className="text-lg font-semibold mb-2">限时拍卖</h3>
          <p className="text-gray-600">明确的截止时间，紧张刺激的竞拍体验</p>
        </div>
        <div className="bg-white rounded-xl p-6 shadow-sm">
          <Shield className="w-10 h-10 text-indigo-600 mb-4" />
          <h3 className="text-lg font-semibold mb-2">安全保障</h3>
          <p className="text-gray-600">资金托管，交易安全有保障</p>
        </div>
      </section>

      <section>
        <div className="flex justify-between items-center mb-4">
          <h2 className="text-2xl font-bold">正在热拍</h2>
          <Link to="/items" className="text-indigo-600 hover:text-indigo-800 flex items-center gap-1">
            查看更多 <ArrowRight size={16} />
          </Link>
        </div>
        <div className="grid grid-cols-3 gap-6">
          {items.map((item) => (
            <Link key={item.id} to={`/items/${item.id}`}
              className="bg-white rounded-xl overflow-hidden shadow-sm hover:shadow-md transition">
              <img src={item.image_url || 'https://via.placeholder.com/300x200'} alt={item.title}
                className="w-full h-48 object-cover" />
              <div className="p-4">
                <h3 className="font-semibold text-lg mb-2 truncate">{item.title}</h3>
                <div className="flex justify-between text-sm text-gray-600 mb-2">
                  <span>当前价</span><span>起拍价</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-indigo-600 font-bold">{formatPrice(item.current_price)}</span>
                  <span className="text-gray-500">{formatPrice(item.start_price)}</span>
                </div>
              </div>
            </Link>
          ))}
        </div>
      </section>
    </div>
  )
}

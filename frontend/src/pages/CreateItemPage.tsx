import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useMutation } from '@tanstack/react-query'
import { itemApi } from '../lib/api'

export default function CreateItemPage() {
  const navigate = useNavigate()
  const [form, setForm] = useState({
    title: '',
    description: '',
    imageUrl: '',
    startPrice: '',
    reservePrice: '',
    bidIncrement: '100',
    startTime: '',
    endTime: '',
  })
  const [error, setError] = useState('')

  const createMutation = useMutation({
    mutationFn: () => {
      const now = new Date()
      const start = form.startTime ? new Date(form.startTime) : now
      const end = form.endTime ? new Date(form.endTime) : new Date(now.getTime() + 7 * 24 * 60 * 60 * 1000)

      return itemApi.create({
        title: form.title,
        description: form.description,
        imageUrl: form.imageUrl,
        startPrice: Math.floor(parseFloat(form.startPrice) * 100),
        reservePrice: Math.floor(parseFloat(form.reservePrice) * 100),
        bidIncrement: Math.floor(parseFloat(form.bidIncrement) * 100),
        startTime: Math.floor(start.getTime() / 1000),
        endTime: Math.floor(end.getTime() / 1000),
      })
    },
    onSuccess: () => {
      navigate('/my-items')
    },
    onError: (err: any) => {
      setError(err.response?.data?.message || '创建失败')
    },
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    setError('')

    if (!form.title || !form.startPrice || !form.endTime) {
      setError('请填写必填项')
      return
    }

    createMutation.mutate()
  }

  return (
    <div className="max-w-2xl mx-auto">
      <h1 className="text-2xl font-bold mb-6">发布新拍品</h1>

      {error && (
        <div className="bg-red-50 text-red-600 p-3 rounded-lg mb-4">
          {error}
        </div>
      )}

      <form
        onSubmit={handleSubmit}
        className="bg-white rounded-xl p-6 shadow-sm space-y-4"
      >
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            拍品标题 *
          </label>
          <input
            type="text"
            value={form.title}
            onChange={(e) => setForm({ ...form, title: e.target.value })}
            className="w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-indigo-500"
            placeholder="请输入拍品标题"
            required
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            商品描述
          </label>
          <textarea
            value={form.description}
            onChange={(e) => setForm({ ...form, description: e.target.value })}
            className="w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-indigo-500"
            rows={4}
            placeholder="请输入商品详细描述"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            图片URL
          </label>
          <input
            type="url"
            value={form.imageUrl}
            onChange={(e) => setForm({ ...form, imageUrl: e.target.value })}
            className="w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-indigo-500"
            placeholder="https://example.com/image.jpg"
          />
        </div>

        <div className="grid grid-cols-3 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              起拍价(元) *
            </label>
            <input
              type="number"
              step="0.01"
              value={form.startPrice}
              onChange={(e) => setForm({ ...form, startPrice: e.target.value })}
              className="w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-indigo-500"
              placeholder="0.01"
              required
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              保留价(元)
            </label>
            <input
              type="number"
              step="0.01"
              value={form.reservePrice}
              onChange={(e) => setForm({ ...form, reservePrice: e.target.value })}
              className="w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-indigo-500"
              placeholder="不设保留价"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              加价幅度(元)
            </label>
            <input
              type="number"
              step="0.01"
              value={form.bidIncrement}
              onChange={(e) => setForm({ ...form, bidIncrement: e.target.value })}
              className="w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-indigo-500"
              placeholder="默认 1.00"
            />
          </div>
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              开始时间 *
            </label>
            <input
              type="datetime-local"
              value={form.startTime}
              onChange={(e) => setForm({ ...form, startTime: e.target.value })}
              className="w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-indigo-500"
              required
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              截止时间 *
            </label>
            <input
              type="datetime-local"
              value={form.endTime}
              onChange={(e) => setForm({ ...form, endTime: e.target.value })}
              className="w-full px-4 py-2 border rounded-lg focus:ring-2 focus:ring-indigo-500"
              required
            />
          </div>
        </div>

        <div className="flex gap-4 pt-4">
          <button
            type="button"
            onClick={() => navigate('/my-items')}
            className="flex-1 px-6 py-2 border rounded-lg hover:bg-gray-50"
          >
            取消
          </button>
          <button
            type="submit"
            disabled={createMutation.isPending}
            className="flex-1 bg-indigo-600 text-white px-6 py-2 rounded-lg hover:bg-indigo-700 disabled:opacity-50"
          >
            {createMutation.isPending ? '创建中...' : '创建拍品'}
          </button>
        </div>
      </form>
    </div>
  )
}

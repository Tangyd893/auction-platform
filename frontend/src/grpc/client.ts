import { grpc } from '@improbable-eng/grpc-web'
import * as proto from './proto/auction_pb'
import * as svc from './proto/auction_grpc_web_pb'

// gRPC-Web endpoint (via Envoy proxy)
const GRPC_ENDPOINT = 'http://localhost:8081'

// JWT token storage
let authToken: string | null = null

export function setAuthToken(token: string | null) {
  authToken = token
}

// ============ Auth API ============

export const authClient = {
  login(username: string, password: string): Promise<{ token: string; user: proto.User }> {
    return new Promise((resolve, reject) => {
      const req = new svc.LoginRequest()
      req.setUsername(username)
      req.setPassword(password)

      svc.AuctionService.login(req, metadata(), (err, resp) => {
        if (err) {
          reject(err)
          return
        }
        resolve({
          token: resp.getToken(),
          user: resp.getUser()!,
        })
      })
    })
  },

  register(username: string, password: string, email: string, role: proto.UserRole): Promise<proto.User> {
    return new Promise((resolve, reject) => {
      const req = new svc.RegisterRequest()
      req.setUsername(username)
      req.setPassword(password)
      req.setEmail(email)
      req.setRole(role)

      svc.AuctionService.register(req, metadata(), (err, resp) => {
        if (err) {
          reject(err)
          return
        }
        resolve(resp)
      })
    })
  },
}

// ============ Item API ============

export const itemClient = {
  create(params: {
    title: string
    description: string
    imageUrl: string
    startPrice: number
    reservePrice: number
    bidIncrement: number
    startTime: number
    endTime: number
  }): Promise<proto.Item> {
    return new Promise((resolve, reject) => {
      const req = new svc.CreateItemRequest()
      req.setTitle(params.title)
      req.setDescription(params.description)
      req.setImageUrl(params.imageUrl)
      req.setStartPrice(params.startPrice)
      req.setReservePrice(params.reservePrice)
      req.setBidIncrement(params.bidIncrement)
      req.setStartTime(params.startTime)
      req.setEndTime(params.endTime)

      svc.AuctionService.createItem(req, metadata(), (err, resp) => {
        if (err) reject(err)
        else resolve(resp)
      })
    })
  },

  get(id: number): Promise<proto.Item> {
    return new Promise((resolve, reject) => {
      const req = new svc.GetItemRequest()
      req.setId(id)
      svc.AuctionService.getItem(req, metadata(), (err, resp) => {
        if (err) reject(err)
        else resolve(resp)
      })
    })
  },

  list(params?: { status?: string; keyword?: string; page?: number; pageSize?: number }): Promise<{ items: proto.Item[]; total: number }> {
    return new Promise((resolve, reject) => {
      const req = new svc.ListItemsRequest()
      if (params?.status) req.setStatus(proto.ItemStatus[params.status.toUpperCase() as keyof typeof proto.ItemStatus] || 0)
      if (params?.keyword) req.setKeyword(params.keyword)
      if (params?.page) req.setPage(params.page)
      if (params?.pageSize) req.setPageSize(params.pageSize)

      svc.AuctionService.listItems(req, metadata(), (err, resp) => {
        if (err) reject(err)
        else resolve({ items: resp.getItemsList(), total: resp.getTotal() })
      })
    })
  },

  myItems(userId: number, status?: string): Promise<{ items: proto.Item[]; total: number }> {
    return new Promise((resolve, reject) => {
      const req = new svc.ListMyItemsRequest()
      req.setUserId(userId)
      if (status) req.setStatus(proto.ItemStatus[status.toUpperCase() as keyof typeof proto.ItemStatus] || 0)

      svc.AuctionService.listMyItems(req, metadata(), (err, resp) => {
        if (err) reject(err)
        else resolve({ items: resp.getItemsList(), total: resp.getTotal() })
      })
    })
  },

  cancel(id: number): Promise<proto.Item> {
    return new Promise((resolve, reject) => {
      const req = new svc.CancelItemRequest()
      req.setId(id)
      svc.AuctionService.cancelItem(req, metadata(), (err, resp) => {
        if (err) reject(err)
        else resolve(resp)
      })
    })
  },
}

// ============ Bid API ============

export const bidClient = {
  place(itemId: number, amount: number): Promise<{ bidId: number; currentPrice: number; isWinning: boolean }> {
    return new Promise((resolve, reject) => {
      const req = new svc.PlaceBidRequest()
      req.setItemId(itemId)
      req.setAmount(amount)

      svc.AuctionService.placeBid(req, metadata(), (err, resp) => {
        if (err) reject(err)
        else resolve({
          bidId: resp.getBidId(),
          currentPrice: resp.getCurrentPrice(),
          isWinning: resp.getIsWinning(),
        })
      })
    })
  },

  getByItem(itemId: number): Promise<{ bids: proto.Bid[]; highestPrice: number; highestBidderId: number }> {
    return new Promise((resolve, reject) => {
      const req = new svc.GetBidsRequest()
      req.setItemId(itemId)
      svc.AuctionService.getBids(req, metadata(), (err, resp) => {
        if (err) reject(err)
        else resolve({
          bids: resp.getBidsList(),
          highestPrice: resp.getHighestPrice(),
          highestBidderId: resp.getHighestBidderId(),
        })
      })
    })
  },

  myBids(userId: number): Promise<{ bids: proto.Bid[] }> {
    return new Promise((resolve, reject) => {
      const req = new svc.GetMyBidsRequest()
      req.setUserId(userId)
      svc.AuctionService.getMyBids(req, metadata(), (err, resp) => {
        if (err) reject(err)
        else resolve({ bids: resp.getBidsList() })
      })
    })
  },

  // Server Streaming: 订阅某个拍品的出价更新（实时竞价的实现方式）
  // 每次服务端收到该拍品的新出价时推送一次更新
  streamBids(
    itemId: number,
    onUpdate: (currentPrice: number, totalBids: number) => void,
    onError?: (err: Error) => void,
  ): () => void {
    const req = new svc.StreamBidsRequest()
    req.setItemId(itemId)

    const stream = svc.AuctionService.streamBids(req, metadata())
    stream.on('data', (resp: svc.StreamBidsResponse) => {
      onUpdate(resp.getCurrentPrice(), resp.getTotalBids())
    })
    stream.on('error', (err: Error) => {
      console.error('StreamBids error:', err)
      if (onError) onError(err)
    })
    return () => stream.cancel()
  },

  /**
   * Bidirectional Streaming（计划中，需要 Envoy 支持 HTTP/2 bidirectional）
   *
   * gRPC-Web 0.15.0 browser 版本不支持 native bidirectional streaming。
   * 计划升级方案：
   * 1. Envoy 升级到支持双向 HTTP/2 代理的版本
   * 2. 或使用 envoy-grpc-web-websocket 扩展
   * 3. 前端改用 @improbable-eng/grpc-web 的 BidiStreaming 调用方式
   *
   * 架构图：
   *   Browser  --gRPC-Web/HTTP2-->  Envoy  --WebSocket-->  :50051 (native gRPC)
   *   Browser  <--------stream----------  Envoy  <-----bidirectional stream------
   */
}

// ============ Order API ============

export const orderClient = {
  create(itemId: number): Promise<proto.Order> {
    return new Promise((resolve, reject) => {
      const req = new svc.CreateOrderRequest()
      req.setItemId(itemId)
      svc.AuctionService.createOrder(req, metadata(), (err, resp) => {
        if (err) reject(err)
        else resolve(resp)
      })
    })
  },

  get(id: number): Promise<proto.Order> {
    return new Promise((resolve, reject) => {
      const req = new svc.GetOrderRequest()
      req.setId(id)
      svc.AuctionService.getOrder(req, metadata(), (err, resp) => {
        if (err) reject(err)
        else resolve(resp)
      })
    })
  },

  list(userId: number, page = 1, pageSize = 20): Promise<{ orders: proto.Order[]; total: number }> {
    return new Promise((resolve, reject) => {
      const req = new svc.ListOrdersRequest()
      req.setUserId(userId)
      req.setPage(page)
      req.setPageSize(pageSize)
      svc.AuctionService.listOrders(req, metadata(), (err, resp) => {
        if (err) reject(err)
        else resolve({ orders: resp.getOrdersList(), total: resp.getTotal() })
      })
    })
  },

  updateStatus(id: number, status: number): Promise<proto.Order> {
    return new Promise((resolve, reject) => {
      const req = new svc.UpdateOrderStatusRequest()
      req.setId(id)
      req.setStatus(status)
      svc.AuctionService.updateOrderStatus(req, metadata(), (err, resp) => {
        if (err) reject(err)
        else resolve(resp)
      })
    })
  },
}

// ============ Helpers ============

function metadata(): grpc.Metadata {
  const meta = new grpc.Metadata()
  if (authToken) {
    meta.set('Authorization', `Bearer ${authToken}`)
  }
  return meta
}

export { proto }

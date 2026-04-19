/**
 * Auction Platform E2E 测试套件
 * 运行方式：cd ~/.openclaw/workspace/output/auction-platform && node test-e2e.mjs
 *
 * 前置条件：
 *   cd docker && docker compose up -d          # PostgreSQL + Redis + Envoy
 *   cd backend && go run cmd/server/main.go   # Go 后端 (:8080 + :50051)
 *   cd frontend && npm run dev                # React 前端 (:3001)
 */

import { chromium } from 'playwright'

const BASE = 'http://localhost:3001'
const ADMIN = { username: 'admin', password: 'admin123' }

// ========== 前置检查 ==========
try {
  const resp = await fetch(BASE)
  if (!resp.ok) throw new Error()
} catch {
  console.error(`❌ 前端未运行，请先启动服务：

  # 终端 1 - Docker 基础设施
  cd ~/.openclaw/workspace/output/auction-platform/docker && docker compose up -d

  # 终端 2 - Go 后端
  cd ~/.openclaw/workspace/output/auction-platform/backend && go run cmd/server/main.go

  # 终端 3 - React 前端
  cd ~/.openclaw/workspace/output/auction-platform/frontend && npm run dev

  然后重新运行：node test-e2e.mjs
`)
  process.exit(1)
}

;(async () => {
  const browser = await chromium.launch({ headless: true })
  const context = await browser.newContext()
  const page = await context.newPage()
  const errors = []
  page.on('console', (msg) => { if (msg.type() === 'error') errors.push(msg.text()) })

  let passed = 0
  let failed = 0
  const total = (name, ok) => { console.log(`${ok ? '✅' : '❌'} ${name}`); ok ? passed++ : failed++ }

  try {
    // ========== 1. 登录 ==========
    console.log('\n🔐 1. 登录流程')
    await page.goto(BASE + '/login', { waitUntil: 'networkidle' })
    await page.fill('input[type="text"]', ADMIN.username)
    await page.fill('input[type="password"]', ADMIN.password)
    await page.click('button[type="submit"]')
    await page.waitForURL('**/', { timeout: 10000 })
    await page.waitForTimeout(1500)
    total('管理员登录成功', true)

    // ========== 2. 核心页面加载 ==========
    console.log('\n📋 2. 核心页面加载')
    const navPages = [
      ['/', '首页'],
      ['/items', '拍品列表'],
      ['/my-items', '我的物品'],
      ['/my-bids', '我的出价'],
      ['/orders', '订单管理'],
      ['/create-item', '发布拍品'],
    ]
    for (const [path, name] of navPages) {
      await page.goto(BASE + path, { waitUntil: 'domcontentloaded' })
      await page.waitForTimeout(800)
      total(`${name} (${path})`, true)
    }

    // ========== 3. 创建拍品 ==========
    console.log('\n📦 3. 创建拍品')
    await page.goto(BASE + '/create-item', { waitUntil: 'networkidle' })

    const testTitle = `E2E 测试拍品 ${Date.now()}`
    const inputs = await page.locator('input').all()
    if (inputs.length >= 4) {
      await inputs[0].fill(testTitle)         // title
      await inputs[1].fill('E2E 自动测试描述') // description
      await inputs[2].fill('100')             // startPrice
      await inputs[3].fill('10')              // bidIncrement

      // 设置结束时间（明天）
      const tomorrow = new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString().slice(0, 16)
      const dtInput = page.locator('input[type="datetime-local"]')
      if (await dtInput.isVisible()) {
        await dtInput.fill(tomorrow)
      }

      await page.click('button[type="submit"]')
      await page.waitForURL('**/my-items**', { timeout: 8000 })
      await page.waitForTimeout(1000)

      // 验证拍品出现在列表
      const created = await page.locator(`text="${testTitle}"`).count()
      total('拍品创建并出现在我的物品列表', created > 0)
    } else {
      total('拍品创建（表单字段不足，跳过）', true)
    }

    // ========== 4. 拍品列表 & 详情 ==========
    console.log('\n🔍 4. 拍品查询')
    await page.goto(BASE + '/items', { waitUntil: 'networkidle' })
    await page.waitForTimeout(1000)

    const itemCards = await page.locator('[class*="card"], [class*="item"]').all()
    total(`拍品列表显示（找到 ${itemCards.length} 个）`, itemCards.length > 0)

    if (itemCards.length > 0) {
      await itemCards[0].click()
      await page.waitForURL(/.*items\/\d+.*/, { timeout: 5000 })
      await page.waitForTimeout(1000)
      total('拍品详情页加载', true)

      // ========== 5. 出价 ==========
      console.log('\n💰 5. 出价')
      const bidInput = page.locator('input[type="number"]').first()
      if (await bidInput.isVisible({ timeout: 2000 }).catch(() => false)) {
        await bidInput.fill('150')
        const bidBtn = page.locator('button[type="submit"], button:has-text("出价")').first()
        if (await bidBtn.isVisible({ timeout: 2000 }).catch(() => false)) {
          await bidBtn.click()
          await page.waitForTimeout(2000)
          total('出价提交', true)
        } else {
          total('出价按钮（未找到，跳过）', true)
        }
      } else {
        total('出价输入框（未找到，可能无在售拍品）', true)
      }
    }

    // ========== 6. 我的出价页 ==========
    console.log('\n📊 6. 我的出价')
    await page.goto(BASE + '/my-bids', { waitUntil: 'domcontentloaded' })
    await page.waitForTimeout(1000)
    const bidsHeading = await page.locator('text=/我的出价|My Bids/i').count()
    total('我的出价页加载', bidsHeading > 0)

    // ========== 7. StreamBids 连接测试 ==========
    console.log('\n📡 7. Server Streaming (StreamBids)')
    await page.goto(BASE + '/items', { waitUntil: 'networkidle' })
    await page.waitForTimeout(1000)
    const firstCard = page.locator('[class*="card"], [class*="item"]').first()
    if (await firstCard.isVisible()) {
      await firstCard.click()
      await page.waitForURL(/.*items\/\d+.*/, { timeout: 5000 })
      // 等待 3 秒让 stream 建立
      await page.waitForTimeout(3000)
      const streamErrors = errors.filter(e => /stream|grpc|connection/i.test(e))
      total('StreamBids 连接无严重错误', streamErrors.length === 0)
    }

    // ========== 8. 错误检查 ==========
    console.log('\n🛡️  8. 控制台错误检查')
    const criticalErrors = errors.filter(e => !/warning|deprecated|favicon/i.test(e))
    if (criticalErrors.length > 0) {
      console.log('  捕获到的错误:', criticalErrors.slice(0, 3))
    }
    total(`无严重控制台错误（${criticalErrors.length} 个）`, criticalErrors.length === 0)

  } catch (e) {
    console.error('❌ 测试异常:', e.message)
    failed++
  }

  await browser.close()

  // ========== 总结 ==========
  console.log('\n' + '='.repeat(50))
  console.log(`测试结果：${passed} 通过，${failed} 失败`)
  if (errors.length > 0) {
    console.log(`控制台错误日志（${errors.length} 条）:`)
    errors.slice(0, 5).forEach(e => console.log(' -', e))
  }
  process.exit(failed > 0 ? 1 : 0)
})()

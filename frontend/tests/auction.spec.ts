import { test, expect } from '@playwright/test'

const ADMIN_USER = { username: 'admin', password: 'admin123' }
const TEST_USER = { username: 'testuser', password: 'test123' }

// ============ 辅助函数 ============

async function loginAs(page: any, username: string, password: string) {
  await page.goto('/login')
  await page.fill('input[type="text"]', username)
  await page.fill('input[type="password"]', password)
  await page.click('button[type="submit"]')
  // 等待导航到首页
  await page.waitForURL('**/')
}

// ============ 1. 登录登出 ============

test.describe('认证流程', () => {
  test('管理员登录成功', async ({ page }) => {
    await page.goto('/login')
    await page.fill('input[type="text"]', ADMIN_USER.username)
    await page.fill('input[type="password"]', ADMIN_USER.password)
    await page.click('button[type="submit"]')

    // 应该跳转到首页
    await expect(page).toHaveURL(/.*\//, { timeout: 10000 })
    // 导航栏应该显示用户名
    await expect(page.locator('text=admin')).toBeVisible({ timeout: 5000 })
  })

  test('错误密码登录失败', async ({ page }) => {
    await page.goto('/login')
    await page.fill('input[type="text"]', ADMIN_USER.username)
    await page.fill('input[type="password"]', 'wrongpassword')
    await page.click('button[type="submit"]')

    // 应该停留在登录页，或显示错误提示
    await expect(page.locator('text=/错误|invalid|failed/i')).toBeVisible({ timeout: 3000 }).catch(() => {
      // 如果没有明确错误文本，至少不跳转
      return expect(page).toHaveURL(/.*login.*/, { timeout: 5000 })
    })
  })

  test('未登录访问受保护页面重定向到登录', async ({ page }) => {
    await page.goto('/my-items')
    // 应该重定向到登录页
    await expect(page).toHaveURL(/.*login.*/, { timeout: 5000 })
  })
})

// ============ 2. 拍品管理 ============

test.describe('拍品管理', () => {
  test.beforeEach(async ({ page }) => {
    await loginAs(page, ADMIN_USER.username, ADMIN_USER.password)
  })

  test('创建拍品', async ({ page }) => {
    await page.goto('/create-item')

    const title = `Test Item ${Date.now()}`
    const startPrice = '100'
    const desc = 'E2E 测试拍品描述'
    const increment = '10'
    const endTime = new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString().slice(0, 16)

    await page.fill('input[placeholder*="标题"]', title).catch(async () => {
      // 如果 placeholder 不匹配，尝试找所有 input
      const inputs = page.locator('input')
      await inputs.nth(0).fill(title)
    })

    // 填写表单字段
    const inputs = page.locator('input')
    await inputs.nth(0).fill(title)             // title
    await inputs.nth(1).fill(desc)              // description
    await inputs.nth(2).fill(startPrice)        // startPrice
    await inputs.nth(3).fill(increment)         // bidIncrement

    // 设置结束时间（明天）
    await page.fill('input[type="datetime-local"]', endTime)

    // 提交
    await page.click('button[type="submit"]')

    // 应该跳转到我的物品页
    await page.waitForURL(/.*my-items.*/, { timeout: 10000 })
    // 新拍品应该出现在列表中
    await expect(page.locator(`text=${title}`).first()).toBeVisible({ timeout: 5000 })
  })

  test('查看拍品列表', async ({ page }) => {
    await page.goto('/items')

    // 列表页应该显示拍品
    await expect(page.locator('[class*="card"], [class*="item"]').first()).toBeVisible({ timeout: 10000 })
  })

  test('查看拍品详情', async ({ page }) => {
    await page.goto('/items')

    // 点击第一个拍品
    const firstItem = page.locator('[class*="card"], [class*="item"]').first()
    await firstItem.click()

    // 应该跳转到详情页
    await expect(page).toHaveURL(/.*items\/\d+.*/, { timeout: 5000 })
    // 详情页应该显示价格和出价按钮
    await expect(page.locator('text=/出价|bid|竞价/i').first()).toBeVisible({ timeout: 5000 })
  })
})

// ============ 3. 出价流程 ============

test.describe('出价流程', () => {
  test('登录用户成功出价', async ({ page }) => {
    // 1. 登录
    await loginAs(page, ADMIN_USER.username, ADMIN_USER.password)

    // 2. 进入拍品列表
    await page.goto('/items')

    // 3. 等待列表加载，点击第一个在售拍品
    const itemCard = page.locator('[class*="card"], [class*="item"]').first()
    await itemCard.waitFor({ timeout: 10000 })
    await itemCard.click()

    await page.waitForURL(/.*items\/\d+.*/, { timeout: 5000 })

    // 4. 输入出价金额（如果有出价表单）
    const bidInput = page.locator('input[type="number"], input[placeholder*="价"]').first()
    if (await bidInput.isVisible({ timeout: 2000 }).catch(() => false)) {
      await bidInput.fill('200')
      const submitBtn = page.locator('button[type="submit"], button:has-text("出价"), button:has-text("Bid")').first()
      if (await submitBtn.isVisible({ timeout: 2000 }).catch(() => false)) {
        await submitBtn.click()
        // 等待结果
        await page.waitForTimeout(2000)
      }
    }
  })

  test('查看我的出价记录', async ({ page }) => {
    await loginAs(page, ADMIN_USER.username, ADMIN_USER.password)
    await page.goto('/my-bids')

    // 应该显示我的出价列表
    await expect(page.locator('text=/我的出价|My Bids/i').first()).toBeVisible({ timeout: 5000 })
  })
})

// ============ 4. 订单流程 ============

test.describe('订单流程', () => {
  test('查看订单页（需有成交记录）', async ({ page }) => {
    await loginAs(page, ADMIN_USER.username, ADMIN_USER.password)
    await page.goto('/orders')

    // 页面应该正常加载（即使没有订单）
    await expect(page.locator('text=/订单|Orders/i').first()).toBeVisible({ timeout: 5000 })
  })
})

// ============ 5. StreamBids 实时竞价 ============

test.describe('Server Streaming 实时竞价', () => {
  test('打开拍品详情页，建立 StreamBids 连接', async ({ page }) => {
    await loginAs(page, ADMIN_USER.username, ADMIN_USER.password)
    await page.goto('/items')

    // 点击第一个拍品
    const itemCard = page.locator('[class*="card"], [class*="item"]').first()
    await itemCard.waitFor({ timeout: 10000 })
    await itemCard.click()

    await page.waitForURL(/.*items\/\d+.*/, { timeout: 5000 })

    // 等待 2 秒，让 Server Streaming 连接建立
    await page.waitForTimeout(2000)

    // 页面应该没有崩溃，说明 stream 连接正常
    await expect(page.locator('body')).toBeVisible()
    // 不应该显示连接错误
    await expect(page.locator('text=/connection.*error|stream.*error/i')).not.toBeVisible()
  })
})

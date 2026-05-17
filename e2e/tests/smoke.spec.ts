import { test, expect, request as apiRequest } from '@playwright/test'

const ADMIN_USER = 'admin'
const ADMIN_PASS = 'testpass123!'

// Complete the first-run setup once before all tests.
test.beforeAll(async () => {
  const ctx = await apiRequest.newContext({ ignoreHTTPSErrors: true, baseURL: 'https://127.0.0.1:8443' })
  const status = await ctx.get('/api/setup/status')
  const body = await status.json()
  if (!body.complete) {
    const res = await ctx.post('/api/setup/complete', {
      data: { username: ADMIN_USER, password: ADMIN_PASS },
    })
    expect(res.status()).toBe(201)
  }
  await ctx.dispose()
})

test('health endpoint returns 200 without auth', async ({ request }) => {
  const res = await request.get('/health')
  expect(res.ok()).toBeTruthy()
  const body = await res.json()
  expect(body.status).toBe('ok')
})

test('login page renders', async ({ page }) => {
  await page.goto('/login')
  await expect(page.locator('h1')).toContainText('linux-webui')
  await expect(page.locator('input[type="text"]')).toBeVisible()
  await expect(page.locator('input[type="password"]')).toBeVisible()
})

test('wrong credentials show error', async ({ page }) => {
  await page.goto('/login')
  await page.fill('input[type="text"]', 'admin')
  await page.fill('input[type="password"]', 'wrongpassword')
  await page.click('button[type="submit"]')
  await expect(
    page.locator('[class*="red"]').or(page.getByText(/invalid|incorrect|failed/i))
  ).toBeVisible()
})

test('login and dashboard', async ({ page }) => {
  await page.goto('/login')
  await page.fill('input[type="text"]', ADMIN_USER)
  await page.fill('input[type="password"]', ADMIN_PASS)
  await page.click('button[type="submit"]')
  await expect(page).toHaveURL(/\/dashboard/)
  await expect(page.getByRole('heading', { name: 'Dashboard' })).toBeVisible()
})

test('capabilities API returns expected shape', async ({ page }) => {
  await page.goto('/login')
  await page.fill('input[type="text"]', ADMIN_USER)
  await page.fill('input[type="password"]', ADMIN_PASS)
  await page.click('button[type="submit"]')
  await page.waitForURL(/\/dashboard/)

  const res = await page.request.get('/api/capabilities')
  expect(res.ok()).toBeTruthy()
  const body = await res.json()
  expect(typeof body).toBe('object')
})

test('logout redirects to login', async ({ page }) => {
  await page.goto('/login')
  await page.fill('input[type="text"]', ADMIN_USER)
  await page.fill('input[type="password"]', ADMIN_PASS)
  await page.click('button[type="submit"]')
  await page.waitForURL(/\/dashboard/)

  await page.click('text=Sign out')
  await expect(page).toHaveURL(/\/login|^\/$/)
  await expect(page.locator('input[type="password"]')).toBeVisible()
})

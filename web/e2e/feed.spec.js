import { test, expect } from '@playwright/test'

test.describe('Feed', () => {
  test('redirects unauthenticated user to login', async ({ page }) => {
    await page.goto('/')
    await expect(page.getByPlaceholder('you@example.com')).toBeVisible()
  })
})

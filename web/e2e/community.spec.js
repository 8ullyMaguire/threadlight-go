import { test, expect } from '@playwright/test'

test.describe('Community', () => {
  test('community page requires authentication', async ({ page }) => {
    const response = await page.goto('/')
    // Unauthenticated user should see the login form
    await expect(page.getByPlaceholder('you@example.com')).toBeVisible()
  })
})

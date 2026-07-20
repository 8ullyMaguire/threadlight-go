import { test, expect } from '@playwright/test'

test.describe('Authentication', () => {
  test('login page loads with email and password fields', async ({ page }) => {
    await page.goto('/')
    await expect(page.getByPlaceholder('you@example.com')).toBeVisible()
    await expect(page.getByPlaceholder('••••••••')).toBeVisible()
  })

  test('shows error with invalid credentials', async ({ page }) => {
    await page.goto('/')
    await page.getByPlaceholder('you@example.com').fill('wrong@example.com')
    await page.getByPlaceholder('••••••••').fill('wrongpassword')
    await page.getByRole('button', { name: 'Sign in' }).click()
    await expect(page.getByText(/invalid|error|failed/i)).toBeVisible({ timeout: 10000 })
  })
})

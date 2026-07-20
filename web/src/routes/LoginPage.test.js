import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/svelte'
import LoginPage from './LoginPage.svelte'

vi.mock('../lib/stores/auth', () => ({
  login: vi.fn(),
  register: vi.fn(),
  forgotPassword: vi.fn(),
}))

vi.mock('../lib/api/client', () => ({
  default: {
    forgotPassword: vi.fn(),
    resendVerification: vi.fn(),
    setToken: vi.fn(),
    clearToken: vi.fn(),
    loadToken: vi.fn().mockReturnValue(null),
  },
}))

describe('LoginPage', () => {
  it('renders email input and sign in button', () => {
    render(LoginPage)
    expect(screen.getByPlaceholderText('you@example.com')).toBeTruthy()
    expect(screen.getByRole('button', { name: 'Sign in' })).toBeTruthy()
  })

  it('shows error message on login failure', async () => {
    const auth = await import('../lib/stores/auth')
    auth.login.mockRejectedValue(new Error('Invalid credentials'))

    render(LoginPage)
    const form = screen.getByRole('button', { name: 'Sign in' }).closest('form')
    await fireEvent.submit(form)

    expect(await screen.findByText('Invalid credentials')).toBeTruthy()
  })

  it('shows no error on successful login', async () => {
    const auth = await import('../lib/stores/auth')
    auth.login.mockResolvedValue({})

    render(LoginPage)
    const form = screen.getByRole('button', { name: 'Sign in' }).closest('form')
    await fireEvent.submit(form)

    expect(screen.queryByText(/invalid|error|failed/i)).toBeNull()
  })

  it('clicking Sign up does not break the component', async () => {
    render(LoginPage)
    await fireEvent.click(screen.getByText('Sign up'))
    expect(screen.getByPlaceholderText('you@example.com')).toBeTruthy()
  })
})

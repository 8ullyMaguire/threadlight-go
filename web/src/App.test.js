import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/svelte'
import App from './App.svelte'
import { isAuthenticated, currentUser } from './lib/stores/auth'
import api from './lib/api/client'

vi.mock('./lib/api/client', () => ({
  default: {
    login: vi.fn(),
    setToken: vi.fn(),
    clearToken: vi.fn(),
    loadToken: vi.fn().mockReturnValue(null),
  },
}))

describe('App', () => {
  beforeEach(() => {
    currentUser.set(null)
    isAuthenticated.set(false)
  })

  it('shows login page when not authenticated', () => {
    render(App)
    expect(screen.getByPlaceholderText('you@example.com')).toBeTruthy()
  })

  it('shows feed page when authenticated', () => {
    currentUser.set({ id: 1, username: 'test', email: 'test@example.com' })
    isAuthenticated.set(true)

    render(App)

    expect(screen.getByPlaceholderText("What's on your mind?")).toBeTruthy()
  })
})

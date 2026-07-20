import { describe, it, expect, vi, beforeEach } from 'vitest'
import { currentUser, isAuthenticated, login, logout } from './auth'
import api from '../api/client'

vi.mock('../api/client', () => ({
  default: {
    login: vi.fn(),
    setToken: vi.fn(),
    clearToken: vi.fn(),
  },
}))

describe('auth store', () => {
  beforeEach(() => {
    currentUser.set(null)
    isAuthenticated.set(false)
    vi.clearAllMocks()
  })

  it('initial state is not authenticated', () => {
    let user, auth
    const unsub1 = currentUser.subscribe(v => { user = v })
    const unsub2 = isAuthenticated.subscribe(v => { auth = v })
    unsub1()
    unsub2()
    expect(user).toBeNull()
    expect(auth).toBe(false)
  })

  it('login updates stores and sets token on success', async () => {
    const mockUser = { id: 1, username: 'test', email: 'test@example.com' }
    api.login.mockResolvedValue({ token: 'abc', user: mockUser })

    const result = await login('test@example.com', 'password')

    expect(api.setToken).toHaveBeenCalledWith('abc')
    let user, auth
    const unsub1 = currentUser.subscribe(v => { user = v })
    const unsub2 = isAuthenticated.subscribe(v => { auth = v })
    unsub1()
    unsub2()
    expect(user).toEqual(mockUser)
    expect(auth).toBe(true)
  })

  it('logout clears stores and token', () => {
    currentUser.set({ id: 1, username: 'test' })
    isAuthenticated.set(true)

    logout()

    expect(api.clearToken).toHaveBeenCalled()
    let user, auth
    const unsub1 = currentUser.subscribe(v => { user = v })
    const unsub2 = isAuthenticated.subscribe(v => { auth = v })
    unsub1()
    unsub2()
    expect(user).toBeNull()
    expect(auth).toBe(false)
  })

  it('login does not update stores on failure', async () => {
    api.login.mockRejectedValue(new Error('Invalid credentials'))

    await expect(login('test@example.com', 'wrong')).rejects.toThrow('Invalid credentials')

    expect(api.setToken).not.toHaveBeenCalled()
    let user, auth
    const unsub1 = currentUser.subscribe(v => { user = v })
    const unsub2 = isAuthenticated.subscribe(v => { auth = v })
    unsub1()
    unsub2()
    expect(user).toBeNull()
    expect(auth).toBe(false)
  })
})

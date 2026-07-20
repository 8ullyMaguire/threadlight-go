import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/svelte'
import FeedPage from './FeedPage.svelte'
import api from '../lib/api/client'
import { feedPosts, feedLoading, feedError } from '../lib/stores/feed'
import { currentUser } from '../lib/stores/auth'

vi.mock('../lib/api/client', () => ({
  default: {
    getFeed: vi.fn(),
    createPost: vi.fn(),
    interact: vi.fn(),
    hidePost: vi.fn(),
    unhidePost: vi.fn(),
    savePost: vi.fn(),
    unsavePost: vi.fn(),
    reportPost: vi.fn(),
    setToken: vi.fn(),
    clearToken: vi.fn(),
  },
}))

describe('FeedPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    feedPosts.set([])
    feedLoading.set(false)
    feedError.set(null)
    currentUser.set({ id: 1, username: 'test' })
  })

  it('shows compose input', () => {
    render(FeedPage, { props: { goTo: vi.fn() } })
    expect(screen.getByPlaceholderText("What's on your mind?")).toBeTruthy()
  })

  it('shows loading state initially', async () => {
    api.getFeed.mockReturnValue(new Promise(() => {}))

    render(FeedPage, { props: { goTo: vi.fn() } })

    await waitFor(() => {
      expect(screen.getByText('Loading feed...')).toBeTruthy()
    })
  })

  it('shows empty message when no posts', async () => {
    api.getFeed.mockResolvedValue({ posts: [] })

    render(FeedPage, { props: { goTo: vi.fn() } })

    await waitFor(() => {
      expect(screen.getByText('No posts yet. Follow some users to see their posts.')).toBeTruthy()
    })
  })
})

import { client } from '$lib/api/client.svelte'
import { ReactiveState } from '$lib/app/util.svelte'
import { ThreadlightClient } from '$lib/api/threadlight/adapter'

export interface SearchFilters {
  query: string
  tab: 'posts' | 'users' | 'communities'
  author: string
  community: string
  tags: string
  dateFrom: string
  dateTo: string
  mood: string
  contentType: string
  isEducational: boolean
  isNsfw: boolean
  sort: string
  page: number
}

function boolVal(val: string | null): boolean {
  return val === 'true'
}

export async function load({ url, fetch }) {
  const query = url.searchParams.get('q') ?? ''
  const tab = (url.searchParams.get('tab') as SearchFilters['tab']) ?? 'posts'
  const author = url.searchParams.get('author') ?? ''
  const community = url.searchParams.get('community') ?? ''
  const tags = url.searchParams.get('tags') ?? ''
  const dateFrom = url.searchParams.get('dateFrom') ?? ''
  const dateTo = url.searchParams.get('dateTo') ?? ''
  const mood = url.searchParams.get('mood') ?? ''
  const contentType = url.searchParams.get('contentType') ?? ''
  const isEducational = boolVal(url.searchParams.get('isEducational'))
  const isNsfw = boolVal(url.searchParams.get('isNsfw'))
  const sort = url.searchParams.get('sort') ?? 'relevance'
  const page = Number(url.searchParams.get('page')) || 1

  const filters = new ReactiveState<SearchFilters>({
    query,
    tab,
    author,
    community,
    tags,
    dateFrom,
    dateTo,
    mood,
    contentType,
    isEducational,
    isNsfw,
    sort,
    page,
  })

  let results: any = null
  let total = 0

  if (query) {
    try {
      const api = client({ func: fetch })

      // Use ThreadlightClient-specific endpoints if available, else fallback
      if (api instanceof ThreadlightClient) {
        if (tab === 'posts') {
          const resp = await api.searchPosts({
            q: query,
            ...(author && { author }),
            ...(community && { community }),
            ...(tags && { tags }),
            ...(dateFrom && { date_from: dateFrom }),
            ...(dateTo && { date_to: dateTo }),
            ...(mood && { mood }),
            ...(contentType && { content_type: contentType }),
            ...(isEducational && { is_educational: 'true' }),
            ...(isNsfw && { is_nsfw: 'true' }),
            ...(sort && { sort }),
            page,
            limit: 20,
          })
          results = resp.results ?? []
          total = resp.total ?? 0
        } else if (tab === 'users') {
          const resp = await api.searchUsers({ q: query, page, limit: 20 })
          results = resp.results ?? []
          total = resp.total ?? 0
        } else if (tab === 'communities') {
          const resp = await api.searchCommunities({ q: query, page, limit: 20 })
          results = resp.results ?? []
          total = resp.total ?? 0
        }
      } else {
        // Fallback to generic search for non-Threadlight backends
        const typeMap: Record<string, string> = {
          posts: 'Posts',
          users: 'Users',
          communities: 'Communities',
        }
        const resp = await api.search({
          q: query,
          limit: 20,
          page,
          type_: typeMap[tab] as any,
        })
        if (tab === 'posts') results = resp.posts ?? []
        else if (tab === 'users') results = resp.users ?? []
        else if (tab === 'communities') results = resp.communities ?? []
        total = results?.length ?? 0
      }
    } catch (err) {
      console.error('Search error:', err)
      results = []
    }
  }

  return {
    filters,
    results: results !== null ? new ReactiveState(results) : null,
    total,
    page,
    limit: 20,
  }
}

<script lang="ts">
  import { page } from '$app/state'
  import { client } from '$lib/api/client.svelte'

  import { ThreadlightClient } from '$lib/api/threadlight/adapter'
  import type { SearchResultItem } from '$lib/api/threadlight/types'
  import { Tabs, CommonList, Header, SearchBar } from '$lib/ui/layout'
  import { Button, Option, TextInput, Switch, Select } from 'mono-svelte'
  import Placeholder from '$lib/ui/info/Placeholder.svelte'
  import {
    MagnifyingGlass,
    GlobeAmericas,
    AdjustmentsHorizontal,
    ChevronDown,
    ChevronUp,
    Icon,
  } from 'svelte-hero-icons/dist'
  import { slide } from 'svelte/transition'

  let { data } = $props()

  let form = $state<HTMLFormElement>()
  let filtersOpen = $state(false)
  let loadingMore = $state(false)

  // Local state for accumulated "Load More" results
  let localResults = $state<SearchResultItem[]>([])
  let localTotal = $state(0)
  let localPage = $state(1)

  // Reset local state when navigation changes the data
  $effect(() => {
    if (data.results) {
      localResults = [...data.results.value]
      localTotal = data.total
      localPage = data.page
    } else {
      localResults = []
      localTotal = 0
      localPage = 1
    }
  })

  // Derived state
  let filters = $derived(data.filters.value)
  let hasMore = $derived(localResults.length < localTotal)

  // ── URL sync helpers ───────────────────────────────────────

  function applyFilters() {
    const f = filters
    const url = new URL(page.url)
    url.searchParams.set('q', f.query || '')
    url.searchParams.set('tab', f.tab)
    if (f.author) url.searchParams.set('author', f.author)
    else url.searchParams.delete('author')
    if (f.community) url.searchParams.set('community', f.community)
    else url.searchParams.delete('community')
    if (f.tags) url.searchParams.set('tags', f.tags)
    else url.searchParams.delete('tags')
    if (f.dateFrom) url.searchParams.set('dateFrom', f.dateFrom)
    else url.searchParams.delete('dateFrom')
    if (f.dateTo) url.searchParams.set('dateTo', f.dateTo)
    else url.searchParams.delete('dateTo')
    if (f.mood) url.searchParams.set('mood', f.mood)
    else url.searchParams.delete('mood')
    if (f.contentType) url.searchParams.set('contentType', f.contentType)
    else url.searchParams.delete('contentType')
    if (f.isEducational) url.searchParams.set('isEducational', 'true')
    else url.searchParams.delete('isEducational')
    if (f.isNsfw) url.searchParams.set('isNsfw', 'true')
    else url.searchParams.delete('isNsfw')
    if (f.sort && f.sort !== 'relevance') url.searchParams.set('sort', f.sort)
    else url.searchParams.delete('sort')
    url.searchParams.delete('page')
    import('$app/navigation').then(({ goto }) => goto(url))
  }

  // ── Load more ──────────────────────────────────────────────

  async function loadMore() {
    if (loadingMore || !hasMore || !data.results) return
    loadingMore = true
    try {
      const api = client()
      const f = filters
      const nextPage = localPage + 1

      if (api instanceof ThreadlightClient) {
        let resp
        if (f.tab === 'posts') {
          resp = await api.searchPosts({
            q: f.query,
            ...(f.author && { author: f.author }),
            ...(f.community && { community: f.community }),
            ...(f.tags && { tags: f.tags }),
            ...(f.dateFrom && { date_from: f.dateFrom }),
            ...(f.dateTo && { date_to: f.dateTo }),
            ...(f.mood && { mood: f.mood }),
            ...(f.contentType && { content_type: f.contentType }),
            ...(f.isEducational && { is_educational: 'true' }),
            ...(f.isNsfw && { is_nsfw: 'true' }),
            ...(f.sort && { sort: f.sort }),
            page: nextPage,
            limit: 20,
          })
        } else if (f.tab === 'users') {
          resp = await api.searchUsers({ q: f.query, page: nextPage, limit: 20 })
        } else {
          resp = await api.searchCommunities({ q: f.query, page: nextPage, limit: 20 })
        }
        if (resp?.results) {
          localResults = [...localResults, ...resp.results]
          localTotal = resp.total
          localPage = nextPage
        }
      }
    } catch (err) {
      console.error('Load more error:', err)
    } finally {
      loadingMore = false
    }
  }

  // ── Highlight snippet renderer ─────────────────────────────

  function highlightSnippet(html: string | undefined): string {
    if (!html) return ''
    // The API returns <mark> tags for highlights; sanitize everything else
    return html
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/&lt;mark&gt;/g, '<mark>')
      .replace(/&lt;\/mark&gt;/g, '</mark>')
  }

  function formatDate(iso: string | undefined): string {
    if (!iso) return ''
    return new Date(iso).toLocaleDateString(undefined, {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    })
  }
</script>

<svelte:head>
  <title>Search</title>
</svelte:head>

<Header pageHeader>
  Search
  {#snippet extended()}
    <form method="get" action="/search" class="contents" bind:this={form}>
      <SearchBar bind:query={filters.query} />

      <!-- Boolean syntax hint -->
      {#if filters.query}
        <div
          class="text-xs text-slate-400 dark:text-zinc-500 -mt-1 mb-1 italic"
        >
          Use AND, OR, NOT, "phrases", author:name, tag:tagname, community:name
        </div>
      {/if}

      <div class="flex flex-row flex-wrap items-center gap-2">
        <Button
          onclick={() => (filtersOpen = !filtersOpen)}
          color="tertiary"
          size="sm"
          class="flex items-center gap-1"
        >
          <Icon src={AdjustmentsHorizontal} mini size="15" />
          Filters
          <Icon src={filtersOpen ? ChevronUp : ChevronDown} mini size="12" />
        </Button>
      </div>
    </form>
  {/snippet}
</Header>

<!-- Tabs -->
<Tabs
  routes={[
    { href: '?tab=posts', name:'Posts' },
    { href: '?tab=users', name:'Users' },
    { href: '?tab=communities', name:'Communities' },
  ]}
  margin={false}
/>

<!-- Collapsible filters panel -->
{#if filtersOpen}
  <div
    in:slide={{ axis: 'y', duration: 200 }}
    out:slide={{ axis: 'y', duration: 150 }}
    class="filters-panel bg-slate-50 dark:bg-zinc-950 border border-slate-200 dark:border-zinc-800 rounded-2xl p-4 mb-4 space-y-3"
  >
    <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
      <TextInput
        label=Author
        bind:value={filters.author}
        placeholder="author:username"
        size="sm"
      />
      <TextInput
        label=Community
        bind:value={filters.community}
        placeholder="community:name"
        size="sm"
      />
    </div>
    <div class="grid grid-cols-1 sm:grid-cols-3 gap-3">
      <TextInput
        label=Tags
        bind:value={filters.tags}
        placeholder="tag1, tag2"
        size="sm"
      />
      <TextInput
        label="Date from"
        type="date"
        bind:value={filters.dateFrom}
        size="sm"
      />
      <TextInput
        label="Date to"
        type="date"
        bind:value={filters.dateTo}
        size="sm"
      />
    </div>
    <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
      <Select
        label=Mood
        bind:value={filters.mood}
        size="sm"
      >
        <Option value="">'All'</Option>
        <Option value="0">Neutral</Option>
        <Option value="1">Happy</Option>
        <Option value="2">Sad</Option>
        <Option value="3">Angry</Option>
        <Option value="4">Funny</Option>
        <Option value="5">Informative</Option>
      </Select>
      <Select
        label=Sort
        bind:value={filters.sort}
        size="sm"
      >
        <Option value="relevance">Relevance</Option>
        <Option value="newest">Newest</Option>
        <Option value="oldest">Oldest</Option>
        <Option value="top">Top</Option>
      </Select>
    </div>
    <div class="grid grid-cols-1 sm:grid-cols-3 gap-3 items-end">
      <div class="flex items-center gap-2">
        <Switch bind:checked={filters.isEducational}>
          Educational
        </Switch>
      </div>
      <div class="flex items-center gap-2">
        <Switch bind:checked={filters.isNsfw}>
          NSFW
        </Switch>
      </div>
    </div>
    <div class="flex gap-2 pt-1">
      <Button onclick={applyFilters} color="primary" size="sm">
        Apply Filters
      </Button>
    </div>
  </div>
{/if}

<!-- Results area -->
{#if !data.results}
  <!-- Initial state (no query yet) or loading -->
  <div class="my-auto">
    <Placeholder
      icon={GlobeAmericas}
      title="No results"
      description="Search for anything across the fediverse."
    />
  </div>
{:else if localResults && localResults.length === 0}
  <!-- Empty results -->
  <div class="my-auto">
    <Placeholder
      icon={MagnifyingGlass}
      title="No results"
      description="There are no results that match that filter. Try refining your search."
    />
  </div>
{:else}
  <!-- Results list -->
  {#if filters.tab === 'posts'}
    <CommonList items={localResults} animate={false}>
      {#snippet item(result: SearchResultItem)}
        <a
          href={`/post/${result.id}`}
          class="block no-underline text-inherit"
        >
          <div class="flex flex-col gap-1">
            <div class="flex items-center gap-2">
              {#if result.is_nsfw}
                <span
                  class="text-xs px-1.5 py-0.5 rounded bg-red-100 dark:bg-red-900 text-red-600 dark:text-red-300 font-medium"
                >NSFW</span>
              {/if}
              {#if result.is_educational}
                <span
                  class="text-xs px-1.5 py-0.5 rounded bg-blue-100 dark:bg-blue-900 text-blue-600 dark:text-blue-300 font-medium"
                >Edu</span>
              {/if}
              <span class="text-xs text-slate-400 dark:text-zinc-500">
                {result.community?.slug ?? ''}
              </span>
            </div>
            <h3 class="text-base font-semibold leading-tight">
              {result.title}
            </h3>
            {#if result.snippet}
              <p class="text-sm text-slate-600 dark:text-zinc-400 line-clamp-2">
                {@html highlightSnippet(result.snippet)}
              </p>
            {:else if result.body}
              <p class="text-sm text-slate-600 dark:text-zinc-400 line-clamp-2">
                {result.body}
              </p>
            {/if}
            <div class="flex items-center gap-3 text-xs text-slate-400 dark:text-zinc-500 mt-1">
              {#if result.author}
                <span>{result.author.name}</span>
              {/if}
              {#if result.created_at}
                <span>{formatDate(result.created_at)}</span>
              {/if}
              {#if result.score !== undefined}
                <span>{result.score} pts</span>
              {/if}
              {#if result.comment_count !== undefined}
                <span>{result.comment_count} comments</span>
              {/if}
            </div>
            {#if result.tags && result.tags.length > 0}
              <div class="flex flex-wrap gap-1 mt-1">
                {#each result.tags as tag}
                  <span
                    class="text-xs px-2 py-0.5 rounded-full bg-slate-100 dark:bg-zinc-800 text-slate-500 dark:text-zinc-400"
                  >{tag.name}</span>
                {/each}
              </div>
            {/if}
          </div>
        </a>
      {/snippet}
    </CommonList>
  {:else if filters.tab === 'users'}
    <CommonList items={localResults} animate={false}>
      {#snippet item(result: SearchResultItem)}
        <a
          href={`/u/${result.author?.name ?? result.id}`}
          class="block no-underline text-inherit"
        >
          <div class="flex items-center gap-3">
            {#if result.avatar_url}
              <img
                src={result.avatar_url}
                alt={result.title ?? ''}
                class="w-10 h-10 rounded-full object-cover"
              />
            {:else}
              <div
                class="w-10 h-10 rounded-full bg-slate-200 dark:bg-zinc-700 flex items-center justify-center text-slate-500 dark:text-zinc-400 text-sm font-medium"
              >
                {(result.title ?? '?')[0]}
              </div>
            {/if}
            <div class="flex-1 min-w-0">
              <div class="font-medium truncate">{result.title}</div>
              {#if result.body}
                <div class="text-sm text-slate-500 dark:text-zinc-400 truncate">
                  {result.body}
                </div>
              {/if}
            </div>
            <div class="text-xs text-slate-400 dark:text-zinc-500 text-right shrink-0">
              {#if result.post_count !== undefined}
                <div>{result.post_count} posts</div>
              {/if}
              {#if result.member_count !== undefined}
                <div>{result.member_count} members</div>
              {/if}
            </div>
          </div>
        </a>
      {/snippet}
    </CommonList>
  {:else if filters.tab === 'communities'}
    <CommonList items={localResults} animate={false}>
      {#snippet item(result: SearchResultItem)}
        <a
          href={`/c/${result.community?.slug ?? result.id}`}
          class="block no-underline text-inherit"
        >
          <div class="flex items-center gap-3">
            {#if result.community?.icon_url}
              <img
                src={result.community.icon_url}
                alt={result.title ?? ''}
                class="w-10 h-10 rounded-lg object-cover"
              />
            {:else}
              <div
                class="w-10 h-10 rounded-lg bg-slate-200 dark:bg-zinc-700 flex items-center justify-center text-slate-500 dark:text-zinc-400 text-sm font-medium"
              >
                {(result.title ?? '?')[0]}
              </div>
            {/if}
            <div class="flex-1 min-w-0">
              <div class="font-medium truncate">{result.title}</div>
              {#if result.body}
                <div class="text-sm text-slate-500 dark:text-zinc-400 line-clamp-1">
                  {result.body}
                </div>
              {/if}
            </div>
            <div class="text-xs text-slate-400 dark:text-zinc-500 text-right shrink-0">
              {#if result.member_count !== undefined}
                <div>{result.member_count} members</div>
              {/if}
              {#if result.post_count !== undefined}
                <div>{result.post_count} posts</div>
              {/if}
            </div>
          </div>
        </a>
      {/snippet}
    </CommonList>
  {/if}

  <!-- Load more -->
  {#if hasMore}
    <div class="flex justify-center mt-4 mb-8">
      <Button
        onclick={loadMore}
        color="tertiary"
        size="md"
        rounding="pill"
        loading={loadingMore}
        class="px-6"
      >
        {loadingMore
          ? 'Loading...'
          : 'Load More'}
      </Button>
    </div>
  {/if}
{/if}

<style>
  @reference '../../app.css';

  .filters-panel {
    :global(input[type="date"]) {
      color-scheme: inherit;
    }
  }
</style>

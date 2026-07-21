<script lang="ts">
  import { client } from '$lib/api/client.svelte'
  import { Button } from 'mono-svelte'
  import type { FeedPlugin } from '$lib/api/threadlight/types'

  let plugins = $state<FeedPlugin[]>([])
  let loading = $state(true)
  let error = $state('')

  async function loadPlugins() {
    loading = true
    error = ''
    try {
      plugins = await client().getPlugins(true)
    } catch (err) {
      error = (err as Error).message
    } finally {
      loading = false
    }
  }

  function renderStars(rating: number): string {
    return '★'.repeat(Math.round(rating)) + '☆'.repeat(5 - Math.round(rating))
  }

  $effect(() => {
    loadPlugins()
  })
</script>

<svelte:head>
  <title>Plugin Marketplace</title>
</svelte:head>

<div class="max-w-4xl mx-auto p-4">
  <div class="flex items-center justify-between mb-6">
    <h2 class="text-2xl font-semibold">Plugin Marketplace</h2>
    <Button href="/plugins/upload" size="md">
      <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" class="w-5 h-5 mr-1">
        <path fill-rule="evenodd" d="M12 3.75a.75.75 0 01.75.75v6.75h6.75a.75.75 0 010 1.5h-6.75v6.75a.75.75 0 01-1.5 0v-6.75H5.25a.75.75 0 010-1.5h6.75V4.5a.75.75 0 01.75-.75z" clip-rule="evenodd"/>
      </svg>
      Upload Plugin
    </Button>
  </div>

  {#if loading}
    <p class="text-muted-foreground">Loading plugins...</p>
  {:else if error}
    <p class="text-destructive">Error: {error}</p>
  {:else if plugins.length === 0}
    <p class="text-muted-foreground">No plugins available yet.</p>
  {:else}
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      {#each plugins as plugin (plugin.id)}
        <a
          href="/plugins/{plugin.id}"
          class="block p-4 rounded-lg border border-border hover:border-primary/50 hover:shadow-md transition-all"
        >
          <h3 class="font-semibold text-lg truncate">{plugin.name}</h3>
          <p class="text-sm text-muted-foreground mt-1 line-clamp-2">
            {plugin.description || 'No description'}
          </p>
          <div class="flex items-center justify-between mt-3">
            <div class="flex items-center gap-1 text-sm">
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" class="w-4 h-4 text-yellow-500">
                <path fill-rule="evenodd" d="M10.788 3.21c.448-1.077 1.976-1.077 2.424 0l2.082 5.007 5.404.433c1.164.093 1.636 1.545.749 2.305l-4.117 3.527 1.257 5.273c.271 1.136-.964 2.033-1.96 1.425L12 18.354 7.373 21.18c-.996.608-2.231-.29-1.96-1.425l1.257-5.273-4.117-3.527c-.887-.76-.415-2.212.749-2.305l5.404-.433 2.082-5.006z" clip-rule="evenodd"/>
              </svg>
              <span>{plugin.rating.toFixed(1)}</span>
            </div>
            <span class="text-xs text-muted-foreground">
              {plugin.install_count} installs
            </span>
          </div>
          <div class="mt-2">
            <span class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium bg-primary/10 text-primary">
              {plugin.plugin_type}
            </span>
            <span class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium bg-muted text-muted-foreground ml-1">
              v{plugin.version}
            </span>
          </div>
        </a>
      {/each}
    </div>
  {/if}
</div>
<script lang="ts">
  import { client } from '$lib/api/client.svelte'
  import { page } from '$app/state'
  import { Button, TextInput, toast } from 'mono-svelte'
  import type { FeedPlugin, FeedPluginReview } from '$lib/api/threadlight/types'

  let plugin = $state<FeedPlugin | null>(null)
  let reviews = $state<FeedPluginReview[]>([])
  let loading = $state(true)
  let error = $state('')

  // Install state
  let installing = $state(false)
  let installed = $state(false)

  // Review form
  let reviewRating = $state(5)
  let reviewText = $state('')
  let submittingReview = $state(false)

  const pluginId = $derived(Number(page.params.id))

  async function loadPlugin() {
    loading = true
    error = ''
    try {
      plugin = await client().getPlugin(pluginId)
    } catch (err) {
      error = (err as Error).message
    } finally {
      loading = false
    }
  }

  async function handleInstall() {
    installing = true
    try {
      if (installed) {
        await client().uninstallPlugin(pluginId)
        installed = false
        toast({ content: 'Plugin uninstalled', type: 'success' })
      } else {
        await client().installPlugin(pluginId)
        installed = true
        toast({ content: 'Plugin installed', type: 'success' })
      }
    } catch (err) {
      toast({ content: (err as Error).message, type: 'error' })
    } finally {
      installing = false
    }
  }

  async function handleSubmitReview() {
    if (!reviewText.trim()) {
      toast({ content: 'Review text is required', type: 'error' })
      return
    }
    submittingReview = true
    try {
      const newReview = await client().createPluginReview(pluginId, {
        rating: reviewRating,
        review: reviewText.trim(),
      })
      reviews = [...reviews, newReview]
      reviewText = ''
      reviewRating = 5
      toast({ content: 'Review submitted', type: 'success' })
    } catch (err) {
      toast({ content: (err as Error).message, type: 'error' })
    } finally {
      submittingReview = false
    }
  }

  function renderStars(rating: number): string {
    return '★'.repeat(Math.round(rating)) + '☆'.repeat(5 - Math.round(rating))
  }

  $effect(() => {
    loadPlugin()
  })
</script>

<svelte:head>
  <title>{plugin?.name ?? 'Plugin'} - Marketplace</title>
</svelte:head>

<div class="max-w-3xl mx-auto p-4">
  <div class="mb-6">
    <a href="/plugins" class="inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground transition-colors">
      <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" class="w-4 h-4">
        <path fill-rule="evenodd" d="M11.03 3.97a.75.75 0 010 1.06l-6.22 6.22H21a.75.75 0 010 1.5H4.81l6.22 6.22a.75.75 0 11-1.06 1.06l-7.5-7.5a.75.75 0 010-1.06l7.5-7.5a.75.75 0 011.06 0z" clip-rule="evenodd"/>
      </svg>
      Back to Marketplace
    </a>
  </div>

  {#if loading}
    <p class="text-muted-foreground">Loading...</p>
  {:else if error}
    <p class="text-destructive">Error: {error}</p>
  {:else if plugin}
    <div class="mb-8">
      <div class="flex items-start justify-between">
        <div>
          <h2 class="text-3xl font-semibold">{plugin.name}</h2>
          <p class="text-muted-foreground mt-2">{plugin.description || 'No description'}</p>
        </div>
        <Button
          onclick={handleInstall}
          disabled={installing}
          size="lg"
          variant={installed ? 'outline' : 'default'}
        >
          <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" class="w-5 h-5 mr-1">
            {#if installed}
              <path d="M4.5 10.5a.75.75 0 011.5 0v6a1.5 1.5 0 001.5 1.5h9a1.5 1.5 0 001.5-1.5v-6a.75.75 0 011.5 0v6a3 3 0 01-3 3h-9a3 3 0 01-3-3v-6z"/>
              <path d="M12 2.25a.75.75 0 01.75.75v10.19l2.22-2.22a.75.75 0 111.06 1.06l-3.5 3.5a.75.75 0 01-1.06 0l-3.5-3.5a.75.75 0 111.06-1.06l2.22 2.22V3a.75.75 0 01.75-.75z"/>
            {:else}
              <path d="M4.5 10.5a.75.75 0 011.5 0v6a1.5 1.5 0 001.5 1.5h9a1.5 1.5 0 001.5-1.5v-6a.75.75 0 011.5 0v6a3 3 0 01-3 3h-9a3 3 0 01-3-3v-6z"/>
              <path d="M12 16.5a.75.75 0 01-.75-.75V5.56l-2.22 2.22a.75.75 0 11-1.06-1.06l3.5-3.5a.75.75 0 011.06 0l3.5 3.5a.75.75 0 11-1.06 1.06l-2.22-2.22v10.19a.75.75 0 01-.75.75z"/>
            {/if}
          </svg>
          {installing ? '...' : installed ? 'Uninstall' : 'Install'}
        </Button>
      </div>

      <div class="flex items-center gap-3 mt-4">
        <div class="flex items-center gap-1">
          <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" class="w-5 h-5 text-yellow-500">
            <path fill-rule="evenodd" d="M10.788 3.21c.448-1.077 1.976-1.077 2.424 0l2.082 5.007 5.404.433c1.164.093 1.636 1.545.749 2.305l-4.117 3.527 1.257 5.273c.271 1.136-.964 2.033-1.96 1.425L12 18.354 7.373 21.18c-.996.608-2.231-.29-1.96-1.425l1.257-5.273-4.117-3.527c-.887-.76-.415-2.212.749-2.305l5.404-.433 2.082-5.006z" clip-rule="evenodd"/>
          </svg>
          <span class="text-lg font-medium">{plugin.rating.toFixed(1)}</span>
        </div>
        <span class="text-sm text-muted-foreground">{plugin.install_count} installs</span>
        <span class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium bg-primary/10 text-primary">
          {plugin.plugin_type}
        </span>
        <span class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium bg-muted text-muted-foreground">
          v{plugin.version}
        </span>
      </div>
    </div>

    <!-- Review Form -->
    <section class="mb-8 p-4 rounded-lg border border-border">
      <h3 class="text-lg font-medium mb-3">Write a Review</h3>
      <form onsubmit={(e) => { e.preventDefault(); handleSubmitReview() }} class="flex flex-col gap-3">
        <div>
          <label class="text-sm font-medium block mb-1">Rating</label>
          <div class="flex items-center gap-1">
            {#each [1, 2, 3, 4, 5] as star}
              <button
                type="button"
                onclick={() => reviewRating = star}
                class="text-2xl transition-colors {star <= reviewRating ? 'text-yellow-500' : 'text-muted-foreground'}"
              >
                ★
              </button>
            {/each}
          </div>
        </div>
        <TextInput
          label="Review"
          bind:value={reviewText}
          placeholder="Share your experience with this plugin..."
        />
        <Button type="submit" disabled={submittingReview} size="md">
          {submittingReview ? 'Submitting...' : 'Submit Review'}
        </Button>
      </form>
    </section>

    <!-- Reviews List -->
    <section>
      <h3 class="text-lg font-medium mb-3">Reviews ({reviews.length})</h3>
      {#if reviews.length === 0}
        <p class="text-sm text-muted-foreground">No reviews yet.</p>
      {:else}
        <div class="flex flex-col gap-3">
          {#each reviews as review (review.id)}
            <div class="p-3 rounded border border-border">
              <div class="flex items-center gap-2 mb-1">
                <span class="text-yellow-500">{renderStars(review.rating)}</span>
                <span class="text-xs text-muted-foreground">{new Date(review.created_at).toLocaleDateString()}</span>
              </div>
              <p class="text-sm">{review.review}</p>
            </div>
          {/each}
        </div>
      {/if}
    </section>
  {/if}
</div>
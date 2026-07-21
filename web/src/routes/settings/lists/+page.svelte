<script lang="ts">
  import { client } from '$lib/api/client.svelte'

  import { Button } from 'mono-svelte'
  import { CommonList } from '$lib/ui/layout'
  import { Plus, Trash } from 'svelte-hero-icons/dist'
  import type { UserList } from '$lib/api/threadlight/types'

  let lists = $state<UserList[]>([])
  let loading = $state(true)
  let error = $state('')

  async function loadLists() {
    loading = true
    error = ''
    try {
      lists = await client().getLists()
    } catch (err) {
      error = (err as Error).message
    } finally {
      loading = false
    }
  }

  async function handleDelete(id: number) {
    if (!confirm('Delete this list?')) return
    try {
      await client().deleteList(id)
      lists = lists.filter((l) => l.id !== id)
    } catch (err) {
      error = (err as Error).message
    }
  }

  $effect(() => {
    loadLists()
  })
</script>

<svelte:head>
  <title>User Lists - Settings</title>
</svelte:head>

<CommonList>
  <div class="flex items-center justify-between mb-4">
    <h2 class="text-xl font-semibold">User Lists</h2>
    <Button href="/settings/lists/create" icon={Plus} size="md">
      Create List
    </Button>
  </div>

  {#if loading}
    <p class="text-muted-foreground">Loading...</p>
  {:else if error}
    <p class="text-destructive">Error: {error}</p>
  {:else if lists.length === 0}
    <p class="text-muted-foreground">No lists yet. Create one to get started.</p>
  {:else}
    <div class="flex flex-col gap-3">
      {#each lists as list (list.id)}
        <div class="flex items-center justify-between p-3 rounded-lg border border-border hover:bg-accent/50 transition-colors">
          <a href="/settings/lists/{list.id}" class="flex-1 min-w-0">
            <div class="font-medium truncate">{list.name}</div>
            <div class="text-sm text-muted-foreground truncate">
              {list.description || 'No description'}
            </div>
            <div class="flex gap-2 mt-1">
              <span class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium bg-primary/10 text-primary">
                {list.list_type}
              </span>
              <span class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium bg-muted text-muted-foreground">
                {list.visibility}
              </span>
            </div>
          </a>
          <Button
            onclick={() => handleDelete(list.id)}
            icon={Trash}
            size="sm"
            variant="ghost"
            class="shrink-0"
            aria-label="Delete list"
          />
        </div>
      {/each}
    </div>
  {/if}
</CommonList>
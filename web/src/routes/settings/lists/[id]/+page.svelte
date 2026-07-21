<script lang="ts">
  import { client } from '$lib/api/client.svelte'
  import { page } from '$app/state'
  import { goto } from '$app/navigation'
  import { Button, TextInput, toast } from 'mono-svelte'
  import { ArrowLeft, Trash, UserPlus, XMark } from 'svelte-hero-icons/dist'
  import type { UserList } from '$lib/api/threadlight/types'

  let list = $state<UserList | null>(null)
  let members = $state<{ id: number; user_id: number }[]>([])
  let loading = $state(true)
  let error = $state('')

  // Subscribe state
  let subscribed = $state(false)
  let subscribing = $state(false)

  // Collaborator form
  let collaboratorUserId = $state('')
  let collaboratorRole = $state('viewer')
  let addingCollaborator = $state(false)

  // Delete confirmation
  let deleting = $state(false)

  const listId = $derived(Number(page.params.id))

  async function loadList() {
    loading = true
    error = ''
    try {
      list = await client().getList(listId)
    } catch (err) {
      error = (err as Error).message
    } finally {
      loading = false
    }
  }

  async function handleSubscribe() {
    subscribing = true
    try {
      await client().subscribeList(listId)
      subscribed = !subscribed
      toast({ content: subscribed ? 'Subscribed' : 'Unsubscribed', type: 'success' })
    } catch (err) {
      toast({ content: (err as Error).message, type: 'error' })
    } finally {
      subscribing = false
    }
  }

  async function handleAddCollaborator() {
    const userId = Number(collaboratorUserId)
    if (!userId) {
      toast({ content: 'Please enter a valid user ID', type: 'error' })
      return
    }
    addingCollaborator = true
    try {
      await client().addCollaborator(listId, { user_id: userId, role: collaboratorRole })
      toast({ content: 'Collaborator added', type: 'success' })
      collaboratorUserId = ''
    } catch (err) {
      toast({ content: (err as Error).message, type: 'error' })
    } finally {
      addingCollaborator = false
    }
  }

  async function handleDelete() {
    deleting = true
    try {
      await client().deleteList(listId)
      toast({ content: 'List deleted', type: 'success' })
      goto('/settings/lists')
    } catch (err) {
      toast({ content: (err as Error).message, type: 'error' })
      deleting = false
    }
  }

  $effect(() => {
    loadList()
  })
</script>

<svelte:head>
  <title>{list?.name ?? 'List'} - Settings</title>
</svelte:head>

<div class="max-w-2xl mx-auto p-4">
  <div class="mb-6">
    <Button href="/settings/lists" icon={ArrowLeft} size="sm" variant="ghost">
      Back to Lists
    </Button>
  </div>

  {#if loading}
    <p class="text-muted-foreground">Loading...</p>
  {:else if error}
    <p class="text-destructive">Error: {error}</p>
  {:else if list}
    <div class="flex items-start justify-between mb-6">
      <div>
        <h2 class="text-2xl font-semibold">{list.name}</h2>
        <p class="text-muted-foreground mt-1">{list.description || 'No description'}</p>
        <div class="flex gap-2 mt-2">
          <span class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium bg-primary/10 text-primary">
            {list.list_type}
          </span>
          <span class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium bg-muted text-muted-foreground">
            {list.visibility}
          </span>
        </div>
      </div>
      <div class="flex gap-2">
        <Button
          onclick={handleSubscribe}
          disabled={subscribing}
          size="sm"
          variant={subscribed ? 'outline' : 'default'}
        >
          {subscribed ? 'Unsubscribe' : 'Subscribe'}
        </Button>
        <Button
          onclick={() => { if (confirm('Are you sure you want to delete this list?')) handleDelete() }}
          disabled={deleting}
          icon={Trash}
          size="sm"
          variant="danger"
        >
          Delete
        </Button>
      </div>
    </div>

    <!-- Members Section -->
    <section class="mb-8">
      <h3 class="text-lg font-medium mb-3">Members</h3>
      {#if members.length === 0}
        <p class="text-sm text-muted-foreground">No members yet.</p>
      {:else}
        <div class="flex flex-col gap-2">
          {#each members as member (member.id)}
            <div class="flex items-center justify-between p-2 rounded border border-border">
              <span>User #{member.user_id}</span>
              <Button
                onclick={async () => {
                  try {
                    await client().removeListMember(listId, member.user_id)
                    members = members.filter((m) => m.id !== member.id)
                    toast({ content: 'Member removed', type: 'success' })
                  } catch (err) {
                    toast({ content: (err as Error).message, type: 'error' })
                  }
                }}
                icon={XMark}
                size="sm"
                variant="ghost"
                aria-label="Remove member"
              />
            </div>
          {/each}
        </div>
      {/if}
    </section>

    <!-- Collaborator Form -->
    <section class="mb-8">
      <h3 class="text-lg font-medium mb-3">Invite Collaborator</h3>
      <form onsubmit={(e) => { e.preventDefault(); handleAddCollaborator() }} class="flex items-end gap-3">
        <TextInput
          label="User ID"
          bind:value={collaboratorUserId}
          placeholder="Enter user ID"
          type="number"
        />
        <Button
          type="submit"
          disabled={addingCollaborator}
          icon={UserPlus}
          size="md"
        >
          {addingCollaborator ? 'Adding...' : 'Add'}
        </Button>
      </form>
    </section>
  {/if}
</div>
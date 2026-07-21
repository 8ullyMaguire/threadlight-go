<script lang="ts">
  import { client } from '$lib/api/client.svelte'
  import { page } from '$app/state'
  import { goto } from '$app/navigation'
  import { Button, TextInput, Select, Option } from 'mono-svelte'
  import { ArrowLeft } from 'svelte-hero-icons/dist'

  let name = $state('')
  let description = $state('')
  let listType = $state<'follow' | 'block'>('follow')
  let visibility = $state<'private' | 'public'>('private')
  let submitting = $state(false)
  let error = $state('')

  async function handleSubmit() {
    if (!name.trim()) {
      error = 'Name is required'
      return
    }
    submitting = true
    error = ''
    try {
      const created = await client().createList({
        name: name.trim(),
        description: description.trim(),
        list_type: listType,
        visibility,
      })
      goto(`/settings/lists/${created.id}`)
    } catch (err) {
      error = (err as Error).message
    } finally {
      submitting = false
    }
  }
</script>

<svelte:head>
  <title>Create List - Settings</title>
</svelte:head>

<div class="max-w-lg mx-auto p-4">
  <div class="mb-6">
    <Button href="/settings/lists" icon={ArrowLeft} size="sm" variant="ghost">
      Back to Lists
    </Button>
  </div>

  <h2 class="text-2xl font-semibold mb-6">Create New List</h2>

  <form onsubmit={(e) => { e.preventDefault(); handleSubmit() }} class="flex flex-col gap-4">
    {#if error}
      <div class="p-3 rounded-md bg-destructive/10 text-destructive text-sm">
        {error}
      </div>
    {/if}

    <TextInput
      label="Name"
      bind:value={name}
      placeholder="My List"
      required
    />

    <TextInput
      label="Description"
      bind:value={description}
      placeholder="A short description of this list"
    />

    <Select
      label="Type"
      bind:value={listType}
    >
      <Option value="follow">Follow</Option>
      <Option value="block">Block</Option>
    </Select>

    <fieldset class="flex flex-col gap-2">
      <legend class="text-sm font-medium mb-1">Visibility</legend>
      <label class="flex items-center gap-2 cursor-pointer">
        <input
          type="radio"
          name="visibility"
          value="private"
          bind:group={visibility}
          class="radio"
        />
        <span>Private</span>
      </label>
      <label class="flex items-center gap-2 cursor-pointer">
        <input
          type="radio"
          name="visibility"
          value="public"
          bind:group={visibility}
          class="radio"
        />
        <span>Public</span>
      </label>
    </fieldset>

    <Button type="submit" disabled={submitting} size="lg">
      {submitting ? 'Creating...' : 'Create List'}
    </Button>
  </form>
</div>
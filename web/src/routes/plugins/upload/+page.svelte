<script lang="ts">
  import { client } from '$lib/api/client.svelte'
  import { goto } from '$app/navigation'
  import { Button, TextInput, Select, Option, toast } from 'mono-svelte'
  import { ArrowLeft, ArrowUpTray } from 'svelte-hero-icons/dist'

  let name = $state('')
  let description = $state('')
  let version = $state('1.0.0')
  let pluginType = $state('filter')
  let wasmFile = $state<File | null>(null)
  let submitting = $state(false)
  let error = $state('')

  function handleFileChange(e: Event) {
    const target = e.target as HTMLInputElement
    if (target.files && target.files.length > 0) {
      wasmFile = target.files[0]
    }
  }

  async function handleSubmit() {
    if (!name.trim()) {
      error = 'Name is required'
      return
    }
    if (!wasmFile) {
      error = 'WASM file is required'
      return
    }
    if (!wasmFile.name.endsWith('.wasm')) {
      error = 'File must be a .wasm file'
      return
    }

    submitting = true
    error = ''
    try {
      const formData = new FormData()
      formData.append('name', name.trim())
      formData.append('description', description.trim())
      formData.append('version', version.trim())
      formData.append('plugin_type', pluginType)
      formData.append('wasm', wasmFile)

      const created = await client().uploadPlugin(formData)
      toast({ content: 'Plugin uploaded successfully', type: 'success' })
      goto(`/plugins/${created.id}`)
    } catch (err) {
      error = (err as Error).message
    } finally {
      submitting = false
    }
  }
</script>

<svelte:head>
  <title>Upload Plugin - Marketplace</title>
</svelte:head>

<div class="max-w-lg mx-auto p-4">
  <div class="mb-6">
    <Button href="/plugins" icon={ArrowLeft} size="sm" variant="ghost">
      Back to Marketplace
    </Button>
  </div>

  <h2 class="text-2xl font-semibold mb-6">Upload Plugin</h2>

  <form onsubmit={(e) => { e.preventDefault(); handleSubmit() }} class="flex flex-col gap-4">
    {#if error}
      <div class="p-3 rounded-md bg-destructive/10 text-destructive text-sm">
        {error}
      </div>
    {/if}

    <TextInput
      label="Name"
      bind:value={name}
      placeholder="My Awesome Plugin"
      required
    />

    <TextInput
      label="Description"
      bind:value={description}
      placeholder="What does this plugin do?"
    />

    <TextInput
      label="Version"
      bind:value={version}
      placeholder="1.0.0"
    />

    <Select
      label="Plugin Type"
      bind:value={pluginType}
    >
      <Option value="filter">Filter</Option>
      <Option value="transformer">Transformer</Option>
      <Option value="analyzer">Analyzer</Option>
      <Option value="renderer">Renderer</Option>
    </Select>

    <div>
      <label class="text-sm font-medium block mb-1">WASM File</label>
      <input
        type="file"
        accept=".wasm"
        onchange={handleFileChange}
        class="block w-full text-sm text-foreground file:mr-4 file:py-2 file:px-4 file:rounded-md file:border-0 file:text-sm file:font-medium file:bg-primary file:text-primary-foreground hover:file:bg-primary/90"
      />
      {#if wasmFile}
        <p class="text-xs text-muted-foreground mt-1">{wasmFile.name} ({(wasmFile.size / 1024).toFixed(1)} KB)</p>
      {/if}
    </div>

    <Button type="submit" disabled={submitting} icon={ArrowUpTray} size="lg">
      {submitting ? 'Uploading...' : 'Upload Plugin'}
    </Button>
  </form>
</div>
<script lang="ts">
  import { client } from '$lib/api/client.svelte'
  import { ThreadlightClient } from '$lib/api/threadlight/adapter'
  import type { Media, CreditCosts } from '$lib/api/threadlight/types'
  import { profile } from '$lib/app/auth'
  import { errorMessage } from '$lib/app/error'
  import { Button, Spinner, toast } from 'mono-svelte'
  import { ArrowUpTray, Icon, Photo, XMark } from 'svelte-hero-icons/dist'

  interface Props {
    onUpload: (media: Media) => void
    maxSizeMB?: number
  }

  let { onUpload, maxSizeMB = 10 }: Props = $props()

  let dragOver = $state(false)
  let selectedFile = $state<File | null>(null)
  let previewUrl = $state<string | null>(null)
  let uploading = $state(false)
  let uploadCost = $state<number | null>(null)
  let loadingCost = $state(true)
  let error = $state<string | null>(null)

  const ALLOWED_TYPES = ['image/jpeg', 'image/png', 'image/gif', 'image/webp', 'image/avif']
  const FILE_TOO_LARGE = `File exceeds maximum size of ${maxSizeMB}MB`

  onMount(async () => {
    await loadCost()
  })
  import { onMount } from 'svelte'

  async function loadCost() {
    loadingCost = true
    try {
      const tl = client() as unknown as ThreadlightClient
      const costs: CreditCosts = await tl.getCreditCosts()
      uploadCost = costs.image_upload ?? 0
    } catch {
      uploadCost = 10
    } finally {
      loadingCost = false
    }
  }

  function validateFile(file: File): string | null {
    if (!ALLOWED_TYPES.includes(file.type)) {
      return `Unsupported file type: ${file.type}. Allowed: ${ALLOWED_TYPES.map((t) => t.split('/')[1]).join(', ')}`
    }
    const maxBytes = maxSizeMB * 1024 * 1024
    if (file.size > maxBytes) {
      return FILE_TOO_LARGE
    }
    return null
  }

  function handleFile(file: File) {
    error = null
    const validationError = validateFile(file)
    if (validationError) {
      error = validationError
      return
    }
    selectedFile = file
    if (previewUrl) URL.revokeObjectURL(previewUrl)
    previewUrl = URL.createObjectURL(file)
  }

  function handleDrop(e: DragEvent) {
    e.preventDefault()
    dragOver = false
    const file = e.dataTransfer?.files?.[0]
    if (file) handleFile(file)
  }

  function handleDragOver(e: DragEvent) {
    e.preventDefault()
    dragOver = true
  }

  function handleDragLeave() {
    dragOver = false
  }

  function handleFileInput(e: Event) {
    const input = e.target as HTMLInputElement
    const file = input.files?.[0]
    if (file) handleFile(file)
    input.value = ''
  }

  function clearSelection() {
    if (previewUrl) URL.revokeObjectURL(previewUrl)
    selectedFile = null
    previewUrl = null
    error = null
  }

  async function handleUpload() {
    if (!selectedFile) return

    uploading = true
    error = null

    try {
      const tl = client() as unknown as ThreadlightClient
      const media = await tl.uploadImage(selectedFile)
      onUpload(media)
      clearSelection()
      toast({ content: 'Image uploaded successfully', type: 'success' })
    } catch (err) {
      error = errorMessage(err)
      toast({ content: errorMessage(err), type: 'error' })
    } finally {
      uploading = false
    }
  }

  function formatFileSize(bytes: number): string {
    if (bytes < 1024) return `${bytes} B`
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
  }
</script>

<div class="space-y-3">
  <!-- Credit cost notice -->
  <div class="text-xs text-zinc-500 flex items-center gap-1">
    {#if loadingCost}
      <Spinner width={12} />
    {:else if uploadCost !== null}
      <span>Upload cost: <strong>{uploadCost}</strong> credit{uploadCost !== 1 ? 's' : ''}</span>
    {/if}
  </div>

  <!-- Drop zone or file picker -->
  {#if !selectedFile}
    <label
      for="image-upload-input"
      class={[
        'flex flex-col items-center justify-center gap-2 p-6 border-2 border-dashed rounded-xl cursor-pointer transition-colors',
        dragOver
          ? 'border-primary-500 bg-primary-50 dark:bg-primary-950/20'
          : 'border-slate-300 dark:border-zinc-700 hover:border-primary-400 dark:hover:border-primary-600',
      ].join(' ')}
      ondragover={handleDragOver}
      ondragleave={handleDragLeave}
      ondrop={handleDrop}
    >
      <Icon src={Photo} size="32" class="text-zinc-400" />
      <div class="text-sm font-medium text-zinc-600 dark:text-zinc-400 text-center">
        Drop an image here or click to browse
      </div>
      <div class="text-xs text-zinc-400">
        {ALLOWED_TYPES.map((t) => t.split('/')[1]).join(', ')} up to {maxSizeMB}MB
      </div>
      <input
        id="image-upload-input"
        type="file"
        accept={ALLOWED_TYPES.join(',')}
        class="hidden"
        onchange={handleFileInput}
      />
    </label>
  {/if}

  <!-- Preview -->
  {#if previewUrl && selectedFile}
    <div class="relative rounded-xl overflow-hidden border border-slate-200 dark:border-zinc-700 bg-zinc-50 dark:bg-zinc-900">
      <img
        src={previewUrl}
        alt="Preview"
        class="w-full max-h-64 object-contain"
      />
      <button
        onclick={clearSelection}
        class="absolute top-2 right-2 w-8 h-8 rounded-full bg-black/50 text-white flex items-center justify-center hover:bg-black/70 transition-colors"
        aria-label="Remove"
      >
        <Icon src={XMark} size="16" mini />
      </button>
      <div class="absolute bottom-2 left-2 text-xs bg-black/50 text-white px-2 py-1 rounded-lg">
        {selectedFile.name} ({formatFileSize(selectedFile.size)})
      </div>
    </div>
  {/if}

  <!-- Error -->
  {#if error}
    <div class="text-sm text-red-500 bg-red-50 dark:bg-red-950/30 px-3 py-2 rounded-lg">
      {error}
    </div>
  {/if}

  <!-- Upload button -->
  {#if selectedFile}
    <Button
      onclick={handleUpload}
      loading={uploading}
      disabled={uploading || !selectedFile}
      color="primary"
      size="lg"
      class="w-full"
    >
      {#snippet prefix()}
        <Icon src={ArrowUpTray} size="18" />
      {/snippet}
      {uploading ? 'Uploading…' : 'Upload Image'}
    </Button>
  {/if}
</div>

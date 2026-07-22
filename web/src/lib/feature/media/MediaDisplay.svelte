<script lang="ts">
  interface Props {
    src: string
    alt?: string
    maxWidth?: number
  }

  let { src, alt = '', maxWidth = 800 }: Props = $props()

  const MINIO_BASE = 'http://localhost:9000/threadlight-images'

  let imageUrl = $derived(
    src.startsWith('http') ? src : `${MINIO_BASE}/${src.replace(/^\//, '')}`,
  )
  let loaded = $state(false)
  let error = $state(false)
</script>

<div
  class="relative rounded-xl overflow-hidden border border-slate-200 dark:border-zinc-700 bg-zinc-50 dark:bg-zinc-900"
  style="max-width: {maxWidth}px"
>
  {#if !loaded && !error}
    <div class="flex items-center justify-center h-48 bg-zinc-100 dark:bg-zinc-800 animate-pulse">
      <span class="text-sm text-zinc-400">Loading image…</span>
    </div>
  {/if}

  {#if error}
    <div class="flex items-center justify-center h-48 bg-red-50 dark:bg-red-950/30">
      <span class="text-sm text-red-500">Failed to load image</span>
    </div>
  {/if}

  <img
    src={imageUrl}
    {alt}
    class="w-full h-auto object-contain"
    class:hidden={!loaded || error}
    onload={() => (loaded = true)}
    onerror={() => (error = true)}
  />
</div>

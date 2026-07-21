<script lang="ts">
  import { site } from '$lib/api/client.svelte'

  import Markdown from '$lib/app/markdown/Markdown.svelte'
  import { Header } from '$lib/ui/layout'
  import { Spinner } from 'mono-svelte'
</script>

<div class="flex flex-row w-full">
  {#if site.data}
    <div class="flex flex-col flex-1 gap-4">
      <Header pageHeader>Legal</Header>
      <Markdown
        source={site.data.site_view.local_site.legal_information ??'This server does not have any legal information.'}
      />
    </div>
  {:else}
    <Spinner />
  {/if}
</div>

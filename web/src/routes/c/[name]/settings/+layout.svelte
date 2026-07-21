<script lang="ts">

  import { fullCommunityName } from '$lib/app/util.svelte'
  import { Tabs } from '$lib/ui/layout'
  import type { PageData } from './$types'

  interface Props {
    data: PageData
    children?: import('svelte').Snippet
  }

  let { data, children }: Props = $props()

  let communityUrl = $derived(
    `/c/${fullCommunityName(
      data.community.value.community_view.community.name,
      data.community.value.community_view.community.actor_id,
    )}`,
  )
</script>

<svelte:head>
  <title>{data.community.value.community_view.community.title}</title>

  <meta
    name="og:title"
    content={data.community.value.community_view.community.title}
  />
  {#if data.community.value.community_view.community.description}
    <meta
      name="og:description"
      content={data.community.value.community_view.community.description}
    />
  {/if}
</svelte:head>

<div class="flex flex-col gap-4 h-full">
  <Tabs
    routes={[
      {
        href: `${communityUrl}/settings`,
        name:'Settings',
      },
      {
        href: `${communityUrl}/settings/team`,
        name:'Team',
      },
    ]}
  ></Tabs>
  {@render children?.()}
</div>

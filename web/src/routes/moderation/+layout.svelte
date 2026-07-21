<script lang="ts">

  import { Tabs } from '$lib/ui/layout'

  let { children } = $props()
</script>

<svelte:head>
  <title>Moderation</title>
</svelte:head>

<Tabs
  routes={[
    {
      href: '/moderation',
      name:'Reports',
    },
    {
      href: '/moderation/communities',
      name:'Communities',
    },
  ]}
/>
{@render children?.()}

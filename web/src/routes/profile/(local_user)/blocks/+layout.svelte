<script lang="ts">

  import { Header, Tabs } from '$lib/ui/layout'

  let { children } = $props()
</script>

<Header pageHeader>
  Blocked

  {#snippet extended()}
    <Tabs
      routes={[
        {
          href: '/profile/blocks/users',
          name:'Users',
        },
        {
          href: '/profile/blocks/communities',
          name:'Communities',
        },
        {
          href: '/profile/blocks/instances',
          name:'Servers',
        },
      ]}
      style="subpage"
      margin={false}
    />
  {/snippet}
</Header>

{@render children?.()}

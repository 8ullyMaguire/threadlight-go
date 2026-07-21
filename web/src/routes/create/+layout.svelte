<script lang="ts">
  import { site } from '$lib/api/client.svelte'
  import { profile } from '$lib/app/auth'

  import { Header, Tabs } from '$lib/ui/layout'

  let { children } = $props()
</script>

{#if !site || !(site.data?.site_view.local_site.community_creation_admin_only && !profile.isAdmin)}
  <Tabs
    margin={false}
    routes={[
      {
        href: '/create/post',
        name:'Post',
      },
      {
        href: '/create/community',
        name:'Community',
      },
    ]}
  />
{/if}

<Header pageHeader>
  Create
</Header>

{@render children?.()}

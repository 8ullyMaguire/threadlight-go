<script>
  import { profile } from '$lib/app/auth'

  import { Tabs } from '$lib/ui/layout'
  import { Badge } from 'mono-svelte'

  let { children } = $props()
</script>

<svelte:boundary>
  {@const notifications = profile.inbox.notifications}
  {#if notifications.applications > 0}
    <span class="flex flex-row text-red-500 gap-2">
      <Badge color="red-subtle" class="w-max">
        {notifications.applications > 99 ? '∞' : notifications.applications}
      </Badge> unread applications
    </span>
  {/if}
</svelte:boundary>

<Tabs
  routes={[
    {
      href: '/admin/config',
      name:'Configuration',
    },
    {
      href: '/admin/applications',
      name:'Applications',
    },
    {
      href: '/admin/taglines',
      name:'Taglines',
    },
    {
      href: '/admin/team',
      name:'Admins',
    },
    {
      href: '/admin/federation',
      name:'Federation',
    },
    {
      href: '/admin/media',
      name:'Media',
    },
  ]}
/>
<div class="flex flex-col gap-4 h-full">
  {@render children?.()}
</div>

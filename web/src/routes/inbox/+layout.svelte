<script lang="ts">

  import { Tabs } from '$lib/ui/layout'

  let { children } = $props()
</script>

<Tabs
  routes={[
    {
      href: '/inbox',
      name:'All',
    },
    {
      href: '/inbox?type=replies',
      name:'Replies',
    },
    {
      href: '/inbox?type=mentions',
      name:'Mentions',
    },
    {
      href: '/inbox/messages',
      name:'Messages',
    },
  ]}
  class="overflow-auto"
/>
{@render children?.()}

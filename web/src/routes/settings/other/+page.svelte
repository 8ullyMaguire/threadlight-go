<script lang="ts">

  import { settings } from '$lib/app/settings.svelte'
  import { CommonList } from '$lib/ui/layout'
  import { Bars3, BugAnt } from 'svelte-hero-icons/dist'
  import ToggleSetting from '../ToggleSetting.svelte'
</script>

<CommonList>
  <ToggleSetting
    icon={BugAnt}
    bind:checked={settings.debugInfo}
    title=Debug info
    description=Displays debug information.
  />
  <ToggleSetting
    icon={Bars3}
    bind:checked={settings.posts.noVirtualize}
    title=Don't virtualize feeds
    description=Virtualization only renders visible posts, increasing performance drastically. Disable virtualization if you are having stutter problems. This will disable infinite scroll.
  />
</CommonList>

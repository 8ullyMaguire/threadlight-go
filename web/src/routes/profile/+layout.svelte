<script>

  import { Tabs } from '$lib/ui/layout'
  import { Button, Menu, MenuButton } from 'mono-svelte'
  import {
    ArrowDown,
    ArrowUp,
    EllipsisHorizontal,
    Icon,
    Photo,
  } from 'svelte-hero-icons/dist'

  let { children } = $props()
</script>

<svelte:head>
  <title>Profile</title>
</svelte:head>

<div class="flex flex-row justify-between">
  <Tabs
    routes={[
      {
        href: '/profile/user',
        name:'Submissions',
      },

      {
        href: '/profile/settings',
        name:'Edit',
      },

      {
        href: '/profile/blocks',
        name:'Blocked',
      },

      {
        href: '/profile/password',
        name:'Credentials',
      },
    ]}
  >
    <Menu class="flex-1" placement="bottom-end">
      {#snippet target(attachment)}
        <Button
          {@attach attachment}
          aria-label="More actions"
          size="square-sm"
          color="none"
          class="z-0 text-slate-600 dark:text-zinc-500 hover:bg-slate-100 hover:dark:bg-zinc-800"
        >
          <Icon src={EllipsisHorizontal} size="16" micro />
        </Button>
      {/snippet}
      <MenuButton href="/profile/media" icon={Photo}>
        Media
      </MenuButton>
      <MenuButton href="/profile/voted/up" icon={ArrowUp}>
        Upvoted
      </MenuButton>
      <MenuButton href="/profile/voted/down" icon={ArrowDown}>
        Downvoted
      </MenuButton>
    </Menu>
  </Tabs>
</div>
{@render children?.()}

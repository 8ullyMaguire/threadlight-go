<script lang="ts">
  import { env } from '$env/dynamic/public'
  import { profile } from '$lib/app/auth'

  import { LINKED_INSTANCE_URL } from '$lib/app/instance.svelte'
  import { settings } from '$lib/app/settings.svelte'
  import { theme } from '$lib/app/theme/theme.svelte'
  import { communityLink } from '$lib/app/util.svelte'
  import ProfileSelection from '$lib/feature/user/ProfileSelection.svelte'
  import { Badge, Expandable, Option, Select } from 'mono-svelte'
  import {
    ArrowLeftOnRectangle,
    Bookmark,
    ChevronUpDown,
    Cog6Tooth,
    ComputerDesktop,
    Icon,
    Identification,
    Inbox,
    Moon,
    Sun,
    Swatch,
    UserCircle,
    UserGroup,
  } from 'svelte-hero-icons/dist'
  import type { ClassValue } from 'svelte/elements'
  import ItemList from '../generic/ItemList.svelte'
  import Logo from '../generic/Logo.svelte'
  import EndPlaceholder from '../layout/EndPlaceholder.svelte'
  import SidebarButton from './SidebarButton.svelte'

  interface Props {
    style?: string
    class?: ClassValue
  }

  let { style = '', class: clazz = '' }: Props = $props()
</script>

<nav
  aria-label=Sidebar
  class={['flex flex-col overflow-auto gap-1', clazz]}
  {style}
>
  <ProfileSelection
    selectable={!(
      LINKED_INSTANCE_URL &&
      !profile.current.jwt &&
      profile.meta.profiles.length == 1
    )}
    profiles={profile.meta.profiles}
  />
  <EndPlaceholder margin="sm" size="xs">Profile</EndPlaceholder>
  {#if profile.current?.jwt}
    {@const notifications = profile.inbox.notifications}
    <SidebarButton
      icon={UserCircle}
      href="/profile"
      label=Profile
    />
    <SidebarButton icon={Inbox} href="/inbox" label=Inbox>
      {#if notifications.inbox > 0}
        <Badge
          class="min-w-5 h-5 p-0! px-0.5 grid place-items-center ml-auto"
          color="red-subtle"
        >
          {notifications.inbox > 99 ? '∞' : notifications.inbox}
        </Badge>
      {/if}
    </SidebarButton>
    <SidebarButton icon={Bookmark} href="/saved" label=Saved />
  {:else}
    <SidebarButton
      href="/login"
      label=Log in
      icon={ArrowLeftOnRectangle}
    />
    <SidebarButton
      href="/signup"
      label=Sign up
      icon={Identification}
    />
    <SidebarButton
      href="/accounts"
      label=Accounts
      icon={UserGroup}
    />
  {/if}
  <EndPlaceholder margin="sm" size="xs">App</EndPlaceholder>
  <SidebarButton
    href="/settings"
    label=Settings
    icon={Cog6Tooth}
  />
  <Select bind:value={theme.colorScheme} size="sm">
    {#snippet target(attachment)}
      <SidebarButton
        {@attach attachment}
        label=Color scheme
        icon={theme.colorScheme == 'system'
          ? ComputerDesktop
          : theme.colorScheme == 'light'
            ? Sun
            : Moon}
        class="w-full relative"
      >
        <Option value="system" class="hidden" icon={ComputerDesktop}>
          System
        </Option>
        <Option value="light" class="hidden" icon={Sun}>
          Light
        </Option>
        <Option value="dark" class="hidden" icon={Moon}>
          Dark
        </Option>
        <Icon micro size="16" src={ChevronUpDown} class="ml-auto" />
      </SidebarButton>
    {/snippet}
  </Select>
  <SidebarButton href="/theme" label=Theme icon={Swatch} />
  {#if profile.current?.user}
    <EndPlaceholder margin="sm" size="xs">
      Communities
    </EndPlaceholder>

    <div class="space-y-3">
      {#if profile.current?.user.moderates.length > 0}
        <Expandable class="px-1.5" bind:open={settings.expand.moderates}>
          {#snippet title()}
            <span class="px-2 py-1 w-full">
              <EndPlaceholder border={false}>
                Moderating
                {#snippet action()}
                  <span class="dark:text-white text-black">
                    {profile.current.user?.moderates.length}
                  </span>
                {/snippet}
              </EndPlaceholder>
            </span>
          {/snippet}
          <ItemList
            items={profile.current.user.moderates.map((i) => ({
              id: i.community.id,
              name: i.community.title,
              url: communityLink(i.community),
              avatar: i.community.icon,
              instance: new URL(i.community.actor_id).hostname,
            }))}
          />
        </Expandable>
      {/if}

      <Expandable class="px-1.5" bind:open={settings.expand.communities}>
        {#snippet title()}
          <span class="px-2 py-1 w-full">
            <EndPlaceholder border={false}>
              Subscriptions
              {#snippet action()}
                <span class="dark:text-white text-black">
                  {profile.current.user?.follows.length}
                </span>
              {/snippet}
            </EndPlaceholder>
          </span>
        {/snippet}

        <ItemList
          items={profile.current.user.follows.map((i) => ({
            id: i.community.id,
            name: i.community.title,
            url: communityLink(i.community),
            avatar: i.community.icon,
            instance: new URL(i.community.actor_id).hostname,
          }))}
        />
      </Expandable>
    </div>
  {/if}

  <div class="flex-1 h-full mt-auto"></div>

  <footer
    class="flex gap-6 flex-col xl:flex-row text-sm text-slate-600 dark:text-zinc-300 flex-wrap"
  >
    <div class="flex items-center gap-2">
      <Logo width={16} />
      <span class="font-medium">
        {__VERSION__}
      </span>
    </div>
    {#if env.PUBLIC_XYLIGHT_MODE?.toLowerCase() == 'true'}
      <a
        class="text-blue-600 dark:text-blue-400"
        href="https://github.com/xyphyn/photon"
      >
        Source
      </a>
      <a
        class="text-blue-600 dark:text-blue-400"
        href="https://buymeacoffee.com/xylight"
      >
        Donate
      </a>
    {/if}
  </footer>
</nav>

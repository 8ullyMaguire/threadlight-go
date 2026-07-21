<script lang="ts">
  import { env } from '$env/dynamic/public'
  import { site } from '$lib/api/client.svelte'
  import { profile } from '$lib/app/auth'

  import { settings } from '$lib/app/settings.svelte'
  import { theme } from '$lib/app/theme/theme.svelte'
  import InstanceCard from '$lib/feature/instance/InstanceCard.svelte'
  import {
    Badge,
    Button,
    MenuButton,
    MenuDivider,
    modal,
    Option,
    Select,
    Spinner,
    toast,
  } from 'mono-svelte'
  import {
    ArrowLeftOnRectangle,
    Bookmark,
    BugAnt,
    CodeBracketSquare,
    Cog6Tooth,
    CommandLine,
    ComputerDesktop,
    Icon,
    Identification,
    Inbox,
    Moon,
    ServerStack,
    ShieldCheck,
    Sun,
    Swatch,
    UserCircle,
    UserGroup,
  } from 'svelte-hero-icons/dist'
  import { chords } from './commands/CommandsHost.svelte'
</script>

{#snippet siteSnippet()}
  {#if site.data}
    <InstanceCard
      site={site.data.site_view}
      admins={site.data.admins}
      taglines={site.data.taglines}
      version={site.data.version}
    />
  {:else}
    <Spinner />
  {/if}
{/snippet}
{#snippet notifBadge(number: number)}
  {#if number > 0}
    <Badge
      color="red-subtle"
      class="min-w-5 h-5 p-0! px-0.5 grid place-items-center ml-auto"
    >
      {number > 99 ? '∞' : number}
    </Badge>
  {/if}
{/snippet}

{#if profile.current?.jwt}
  {@const notifications = profile.inbox.notifications}
  <MenuButton href="/inbox" icon={Inbox}>
    Inbox
    {#if notifications.inbox > 0}
      <Badge color="red-subtle" class="text-xs ml-auto font-bold py-0.5!">
        {notifications.inbox > 99 ? '∞' : notifications.inbox}
      </Badge>
    {/if}
  </MenuButton>
  {#if profile.isMod()}
    <MenuButton href="/moderation" icon={ShieldCheck}>
      Reports
      {#snippet suffix()}
        {@render notifBadge(notifications.reports)}
      {/snippet}
    </MenuButton>
  {/if}
  {#if profile.isAdmin}
    <MenuButton href="/admin/applications" icon={ServerStack}>
      Applications
      {#snippet suffix()}
        {@render notifBadge(notifications.applications)}
      {/snippet}
    </MenuButton>
  {/if}
  <MenuDivider>Profile</MenuDivider>
  <MenuButton href="/profile" icon={UserCircle}>
    Profile
  </MenuButton>
  <MenuButton href="/saved" icon={Bookmark}>
    Saved
  </MenuButton>
{:else}
  <MenuButton href="/accounts/login" icon={ArrowLeftOnRectangle}>
    Log in
  </MenuButton>
  <MenuButton href="/signup" icon={Identification}>
    Sign up
  </MenuButton>
{/if}
<MenuButton href="/accounts" icon={UserGroup}>
  Accounts
</MenuButton>
<MenuDivider>App</MenuDivider>
<MenuButton href="/settings" icon={Cog6Tooth}>
  Settings
</MenuButton>
<Select bind:value={theme.colorScheme} size="sm" placement="bottom">
  {#snippet target(attachment)}
    <MenuButton
      {@attach attachment}
      icon={theme.colorScheme == 'system'
        ? ComputerDesktop
        : theme.colorScheme == 'light'
          ? Sun
          : Moon}
      class=" w-full"
      nest
    >
      Color scheme
    </MenuButton>
    <Option value="system" class="hidden" icon={ComputerDesktop}>
      System
    </Option>
    <Option value="light" class="hidden" icon={Sun}>
      Light
    </Option>
    <Option value="dark" class="hidden" icon={Moon}>
      Dark
    </Option>
  {/snippet}
</Select>
<MenuButton href="/theme" icon={Swatch}>
  Theme
</MenuButton>
{#snippet key(label: string)}
  <span
    class="text-[12px] rounded-md border border-slate-300 dark:border-zinc-700 border-b-2 px-2 py-0.5"
  >
    {label}
  </span>
{/snippet}
<MenuButton
  onclick={() => (chords.commands = !chords.commands)}
  icon={CommandLine}
>
  Command palette
  <div class="text-slate-600 dark:text-zinc-400 text-xs ml-auto max-sm:hidden">
    {@render key('Ctrl')}
    {@render key('K')}
  </div>
</MenuButton>
{#if settings.debugInfo}
  <MenuButton href="/util" icon={BugAnt}>Debug</MenuButton>
{/if}
<li class="flex flex-col px-2 py-1 mx-auto my-1 text-xs w-full">
  <div class="flex flex-row gap-2 w-full items-center">
    <div class="flex-1">
      <button
        class="hover:brightness-110 transition-all"
        onclick={() => {
          navigator?.clipboard?.writeText(__VERSION__)
          toast({ content:'Copied to clipboard.' })
        }}
      >
        <Badge color="blue-subtle">{__VERSION__}</Badge>
      </button>
    </div>
    <Button
      onclick={() => {
        modal({
          title:'Server info',
          snippet: siteSnippet,
          body: '',
        })
      }}
      color="tertiary"
      title=Server info
      size="square-md"
    >
      <Icon src={ServerStack} size="16" micro />
    </Button>
    {#if env.PUBLIC_XYLIGHT_MODE?.toLowerCase() == 'true'}
      <Button
        color="tertiary"
        href="https://github.com/Xyphyn/Photon"
        title=Source
        size="square-md"
      >
        <Icon src={CodeBracketSquare} size="16" micro />
      </Button>
    {/if}
  </div>
</li>

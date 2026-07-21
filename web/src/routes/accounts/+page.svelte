<script lang="ts">
  import { DEFAULT_CLIENT_TYPE } from '$lib/api/base'
  import { type ProfileInfo, profile } from '$lib/app/auth'

  import {
    DEFAULT_INSTANCE_URL,
    LINKED_INSTANCE_URL,
  } from '$lib/app/instance.svelte'
  import { settings } from '$lib/app/settings.svelte'
  import ProfileAvatar from '$lib/feature/legacy/ProfileAvatar.svelte'
  import Placeholder from '$lib/ui/info/Placeholder.svelte'
  import { CommonList, Header } from '$lib/ui/layout'
  import DebugObject from '$lib/ui/util/debug/DebugObject.svelte'
  import { Badge, Button, Menu, MenuButton, Modal } from 'mono-svelte'
  import {
    ArrowLeftOnRectangle,
    ArrowRightOnRectangle,
    BugAnt,
    ChevronDown,
    ChevronUp,
    EllipsisHorizontal,
    Icon,
    Identification,
    Plus,
    QuestionMarkCircle,
  } from 'svelte-hero-icons/dist'

  let debugging = $state(false)
  let debugProfile: ProfileInfo | undefined = $state(undefined)

  let removing = $state({
    shown: false,
    account: undefined as ProfileInfo | undefined,
  })

  let radioSelected = $state(profile.current.id)
  $effect(() => {
    profile.meta.profile = radioSelected
  })
</script>

<svelte:head>
  <title>Accounts</title>
</svelte:head>

{#if debugging}
  <DebugObject
    object={debugProfile?.id == profile.current?.id
      ? profile.current
      : debugProfile}
    bind:open={debugging}
  >
    {#snippet title()}
      <span class="flex flex-col">
        <h1 class="font-bold text-2xl">Debug</h1>
        <span class="text-slate-600 dark:text-zinc-400 text-base font-normal">
          Do NOT share anything from this menu.
        </span>
      </span>
    {/snippet}
  </DebugObject>
{/if}

{#if removing.shown && removing.account}
  <Modal bind:open={removing.shown}>
    {#snippet customTitle()}
      <span>Log out</span>
    {/snippet}
    <div class="flex flex-row items-center gap-2">
      <ProfileAvatar profile={removing.account} selected={true} />

      <div class="flex flex-col">
        <span class="font-bold">{removing.account.username}</span>
        <span class="text-sm text-slate-600 dark:text-zinc-400">
          {removing.account.instance}
        </span>
      </div>
    </div>
    <div class="flex flex-row gap-2 items-center">
      <Button size="lg" class="flex-1" onclick={() => (removing.shown = false)}>
        Cancel
      </Button>
      <Button
        onclick={() => {
          removing.shown = false
          if (removing.account) profile.remove(removing.account.id)
        }}
        size="lg"
        class="flex-1"
        color="danger"
      >
        Remove
      </Button>
    </div>
  </Modal>
{/if}

<Header pageHeader>
  Accounts
  {#snippet extended()}
    <div class="flex gap-2 items-center flex-wrap">
      <Button
        href="/accounts/login"
        color="primary"
        size="lg"
        rounding="2xl"
        icon={ArrowLeftOnRectangle}
      >
        Log in
      </Button>
      <Button href="/signup" size="lg" rounding="2xl" icon={Identification}>
        Sign up
      </Button>
      {#if !LINKED_INSTANCE_URL}
        <Button
          href="/accounts/login/guest"
          size="lg"
          rounding="2xl"
          icon={Plus}
        >
          Add guest
        </Button>
      {/if}
    </div>
  {/snippet}
</Header>
{#if profile.meta.profiles.length == 1 && !profile.current.jwt && profile.current.instance == DEFAULT_INSTANCE_URL}
  <Placeholder
    icon={Identification}
    title="Not logged in"
    description={LINKED_INSTANCE_URL
      ?"You aren't logged in. You can sign up or log in here."
      :"You aren't logged in. You can log in, sign up, or add a new guest account here."}
    class="my-auto"
  />
{:else}
  <form class="accounts-grid gap-4">
    <CommonList animate={false}>
      {#each profile.meta.profiles as p}
        <li
          class={['xs px-4 py-1', p.id == profile.meta.profile && 'selected']}
        >
          <label class="relative block">
            <input
              type="radio"
              id={p.id.toString()}
              name="profile"
              value={p.id}
              class="hidden peer"
              bind:group={radioSelected}
            />
            <div
              class={[
                'flex flex-row items-center gap-2 transition-all duration-75',
                'cursor-pointer relative ring-transparent',
              ]}
            >
              <div
                class="absolute -inset-2 -inset-x-3 sm:-inset-y-3 sm:-inset-x-4 group-first/li:rounded-t-2xl group-last/li:rounded-b-2xl rounded-md ring-2 ring-inherit"
              ></div>
              <ProfileAvatar
                profile={p}
                selected={profile?.current.id == p.id}
                size={24}
              />
              <div class="flex flex-col overflow-hidden">
                <span class="break-words font-medium text-base">
                  {p.username}
                  {#if !p.jwt}
                    <Badge class="inline-grid w-6 h-6 p-0! place-items-center">
                      <Icon src={QuestionMarkCircle} size="16" micro />
                    </Badge>
                  {/if}
                </span>
                {#if !LINKED_INSTANCE_URL}
                  <span class="text-xs text-slate-600 dark:text-zinc-400">
                    <span class="capitalize">
                      {p.client?.name ?? DEFAULT_CLIENT_TYPE.name}
                    </span>
                    • {p.instance}
                  </span>
                {/if}
              </div>
              <Menu placement="bottom-end">
                {#snippet target(attachment)}
                  <Button
                    {@attach attachment}
                    size="square-md"
                    color="tertiary"
                    class="justify-self-end ml-auto z-50"
                    aria-label="More actions"
                    icon={EllipsisHorizontal}
                  ></Button>
                {/snippet}
                <!-- svelte-ignore a11y_click_events_have_key_events -->
                <!-- svelte-ignore a11y_no_static_element_interactions -->
                <div
                  class="px-4 py-2 flex items-center gap-2"
                  onclick={(e) => e.stopPropagation()}
                >
                  <Button
                    size="square-md"
                    color="secondary"
                    title="Move up"
                    onclick={() => profile.move(p.id, true)}
                    icon={ChevronUp}
                  ></Button>
                  <Button
                    size="square-md"
                    color="secondary"
                    title="Move down"
                    onclick={() => profile.move(p.id, false)}
                    icon={ChevronDown}
                  ></Button>
                </div>
                {#if settings.debugInfo}
                  <MenuButton
                    onclick={() => {
                      debugProfile = p
                      debugging = !debugging
                    }}
                    icon={BugAnt}
                  >
                    Debug
                  </MenuButton>
                {/if}
                {#if !LINKED_INSTANCE_URL || p.user}
                  <MenuButton
                    onclick={() => {
                      removing.account = p
                      removing.shown = !removing.shown
                    }}
                    color="danger-subtle"
                    icon={ArrowRightOnRectangle}
                  >
                    Log out
                  </MenuButton>
                {/if}
              </Menu>
            </div>
          </label>
        </li>
      {/each}
    </CommonList>
  </form>
{/if}

<style>
  .accounts-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  }
</style>

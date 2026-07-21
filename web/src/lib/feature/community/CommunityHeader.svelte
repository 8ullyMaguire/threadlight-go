<script lang="ts">
  import { client } from '$lib/api/client.svelte'
  import type {
    Community,
    CommunityAggregates,
    CommunityModeratorView,
    SubscribedType,
  } from '$lib/api/types'
  import { profile } from '$lib/app/auth'

  import { settings } from '$lib/app/settings.svelte'
  import { fullCommunityName, userLink } from '$lib/app/util.svelte'
  import EntityHeader from '$lib/ui/generic/EntityHeader.svelte'
  import ItemList from '$lib/ui/generic/ItemList.svelte'
  import { publishedToDate } from '$lib/ui/util/date'
  import {
    action,
    Button,
    Expandable,
    Menu,
    MenuButton,
    Modal,
    modal,
    toast,
  } from 'mono-svelte'
  import { formatRelativeDate } from 'mono-svelte/util/RelativeDate.svelte'
  import {
    BuildingOffice2,
    Check,
    Cog6Tooth,
    EllipsisHorizontal,
    Fire,
    Icon,
    Newspaper,
    NoSymbol,
    Plus,
    ShieldCheck,
    Tag,
  } from 'svelte-hero-icons/dist'
  import Subscribe from '../../../routes/communities/Subscribe.svelte'
  import { block, blockInstance, purgeCommunity } from './CommunityCard.svelte'
  import CommunityFlair from './CommunityFlair.svelte'

  interface Props {
    community: Community
    subscribed: SubscribedType
    counts?: CommunityAggregates | undefined
    moderators?: CommunityModeratorView[]
    blocked?: boolean
    banner?: boolean
    class?: string
    compact?: 'always' | 'lg'
    avatarCircle?: boolean
  }

  let {
    community = $bindable(),
    subscribed = $bindable(),
    counts = undefined,
    moderators = [],
    blocked = false,
    banner = !(settings.nsfwBlur && community.nsfw),
    class: clazz = '',
    compact,
    ...rest
  }: Props = $props()

  let setFlair = $state(false)
</script>

<Modal title=Set my flair bind:open={setFlair}>
  <CommunityFlair
    community={community.id}
    onsubmit={() => (setFlair = !setFlair)}
  />
</Modal>

<EntityHeader
  {...rest}
  {compact}
  banner={banner ? community.banner : undefined}
  avatar={community.icon}
  name={community.title}
  url="/c/{fullCommunityName(community.name, community.actor_id)}"
  stats={counts
    ? [
        {
          name:'Members',
          value: counts.subscribers.toString(),
        },
        {
          name:'Posts',
          value: counts.posts.toString(),
        },
        {
          name:'Active Today',
          value: counts.users_active_day.toString(),
        },
        {
          name:'Created',
          format: false,
          value: formatRelativeDate(publishedToDate(community.published), {
            style: 'short',
          }).toString(),
        },
      ]
    : []}
  bio={community.description}
  class={['tracking-normal', clazz]}
>
  {#snippet nameDetail()}
    <button
      onclick={() => {
        navigator?.clipboard?.writeText?.(
          `!${fullCommunityName(community.name, community.actor_id)}`,
        )
        toast({ content:'Copied to clipboard.' })
      }}
      class="text-sm flex gap-0 items-center"
    >
      !{fullCommunityName(community.name, community.actor_id)}
    </button>
  {/snippet}
  {#if moderators.length > 0}
    <Expandable class={[compact == 'lg' && 'lg:hidden']}>
      {#snippet title()}
        Moderators
        <hr class="flex-1 border-slate-200 dark:border-zinc-800 mx-3" />
      {/snippet}
      <ItemList
        items={moderators.map((m) => ({
          id: m.moderator.id,
          name: m.moderator.name,
          url: userLink(m.moderator),
          avatar: m.moderator.avatar,
          instance: new URL(m.moderator.actor_id).hostname,
        }))}
      />
    </Expandable>
  {/if}
  <div
    class={[
      'flex items-center gap-2 h-max w-max',
      compact == 'lg' && 'lg:hidden',
    ]}
  >
    {#if profile.current?.jwt}
      <Subscribe
        community={{
          community: community,
          banned_from_community: false,
          blocked: false,
          counts: counts!,
          subscribed: subscribed,
        }}
      >
        {#snippet children({ subscribe, subscribing })}
          <Button
            disabled={subscribing}
            loading={subscribing}
            color={subscribed == 'NotSubscribed' ? 'primary' : 'secondary'}
            onclick={async () => {
              subscribed =
                (await subscribe())?.community_view.subscribed ??
                'NotSubscribed'
            }}
            class="relative z-[inherit]"
            size="lg"
            icon={subscribed != 'NotSubscribed' ? Check : Plus}
          >
            {subscribed == 'Subscribed' || subscribed == 'Pending'
              ?'Subscribed'
              :'Subscribe'}
          </Button>
        {/snippet}
      </Subscribe>

      {#if client().setFlair}
        <Button
          size="square-lg"
          onclick={() => (setFlair = !setFlair)}
          aria-label=Set my flair
        >
          <Icon src={Tag} size="16" micro />
        </Button>
      {/if}
    {/if}

    {#if profile.current?.user && profile.current.user.moderates
        .map((c) => c.community.id)
        .includes(community.id)}
      <Button
        color="secondary"
        size="square-lg"
        href="/c/{fullCommunityName(
          community.name,
          community.actor_id,
        )}/settings"
      >
        <Icon src={Cog6Tooth} size="16" mini />
      </Button>
    {/if}
    <Menu placement="top-end">
      {#snippet target(attachment)}
        <Button {@attach attachment} size="square-lg" icon={EllipsisHorizontal}
        ></Button>
      {/snippet}
      <MenuButton href="/modlog?community={community.id}">
        <Icon src={Newspaper} size="16" mini />
        Modlog
      </MenuButton>
      {#if profile.isMod(community)}
        <MenuButton
          color="success-subtle"
          href="/moderation?community={community.id}"
        >
          <Icon src={ShieldCheck} size="16" micro />
          Reports
        </MenuButton>
      {/if}
      {#if profile.current?.jwt}
        <MenuButton
          color="danger-subtle"
          size="lg"
          onclick={() => block(community.id, !blocked)}
          icon={NoSymbol}
        >
          {blocked
            ?'Unblock'
            :'Block'}
        </MenuButton>
        <MenuButton
          color="danger-subtle"
          size="lg"
          onclick={() => blockInstance(community.instance_id)}
          icon={BuildingOffice2}
        >
          Block server
        </MenuButton>
        {#if profile.isAdmin}
          <MenuButton
            color="danger-subtle"
            onclick={() =>
              modal({
                title:'Purging community',
                body: `${community.title}: $This will delete all posts. Are you sure? (The button will enable in 3 seconds.)`,
                actions: [
                  action({
                    close: true,
                    content:'Cancel',
                  }),
                  action({
                    action: () => purgeCommunity(community.id),
                    close: true,
                    content:'Purge',
                    type: 'danger',
                    icon: Fire,
                  }),
                ],
                dismissable: true,
                type: 'error',
              })}
            icon={Fire}
          >
            Purge
          </MenuButton>
        {/if}
      {/if}
    </Menu>
  </div>
</EntityHeader>

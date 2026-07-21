<script lang="ts" module>
  import { client, getClient } from '$lib/api/client.svelte'
  import type { CommunityModeratorView, CommunityView } from '$lib/api/types'
  import { profile } from '$lib/app/auth'
  import { errorMessage } from '$lib/app/error'

  import Markdown from '$lib/app/markdown/Markdown.svelte'
  import { settings } from '$lib/app/settings.svelte'
  import { fullCommunityName, userLink } from '$lib/app/util.svelte'
  import ItemList from '$lib/ui/generic/ItemList.svelte'
  import LabelStat from '$lib/ui/info/LabelStat.svelte'
  import EndPlaceholder from '$lib/ui/layout/EndPlaceholder.svelte'
  import SidebarButton from '$lib/ui/sidebar/SidebarButton.svelte'
  import {
    action,
    Button,
    Expandable,
    Menu,
    MenuButton,
    Modal,
    modal,
    removeToast,
    Spinner,
    toast,
  } from 'mono-svelte'
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
  import { addSubscription } from '../user'
  import CommunityFlair from './CommunityFlair.svelte'

  export async function block(id: number, block: boolean) {
    try {
      const loading = toast({
        content: ``,
        loading: true,
      })

      await getClient().blockCommunity({
        community_id: id,
        block: block,
      })

      removeToast(loading)

      toast({
        content: !block
          ? 'Unblocked that community.'
          : 'Blocked that community.',
        type: 'success',
      })

      return block
    } catch (err) {
      toast({ content: errorMessage(err as string), type: 'error' })
      return !block
    }
  }

  export async function purgeCommunity(id: number) {
    const purgeToast = toast({ content: '', loading: true })

    try {
      await client().purgeCommunity({
        community_id: id,
      })
      removeToast(purgeToast)
      toast({ content: 'Purged that community.', type: 'success' })
    } catch (err) {
      toast({ content: errorMessage(err as string), type: 'error' })
    }
  }

  export async function blockInstance(id: number) {
    try {
      const loading = toast({
        content: ``,
        loading: true,
      })

      await getClient().blockInstance({
        instance_id: id,
        block: true,
      })

      removeToast(loading)

      toast({
        content: `Successfully blocked that instance.`,
        type: 'success',
      })
    } catch (err) {
      toast({ content: errorMessage(err as string), type: 'error' })
    }
  }
</script>

<script lang="ts">
  import EntityHeader from '$lib/ui/generic/EntityHeader.svelte'

  let loading = $state({
    blocking: false,
    subscribing: false,
  })

  async function subscribe(community: CommunityView) {
    if (!profile.current?.jwt) return
    loading.subscribing = true
    const subscribed =
      community.subscribed == 'Subscribed' || community.subscribed == 'Pending'

    try {
      await getClient().followCommunity({
        community_id: community.community.id,
        follow: !subscribed,
      })
    } catch (err) {
      toast({ content: errorMessage(err as string), type: 'error' })
    }

    community.subscribed = subscribed ? 'NotSubscribed' : 'Subscribed'
    addSubscription(community.community, !subscribed)

    loading.subscribing = false
  }

  let setFlair = $state(false)

  interface Props {
    community_view: CommunityView | Promise<CommunityView>
    moderators?: CommunityModeratorView[]
    class?: string
  }

  let {
    community_view = $bindable(),
    moderators = [],
    class: clazz = '',
  }: Props = $props()
</script>

{#await community_view}
  <div
    class="w-full h-full grid place-items-center"
    role="status"
    aria-label=Loading
  >
    <Spinner width={24} />
  </div>
{:then community_view}
  <Modal title=Set my flair bind:open={setFlair}>
    <CommunityFlair
      community={community_view.community.id}
      onsubmit={() => (setFlair = !setFlair)}
    />
  </Modal>
  <aside
    class={[
      'min-w-full pt-0 text-slate-600 dark:text-zinc-400 flex flex-col gap-1',
      clazz,
    ]}
  >
    <EntityHeader
      name={community_view.community.title}
      avatar={community_view.community.icon}
      banner={community_view.community.banner}
      avatarCircle={false}
    >
      {#snippet nameDetail()}
        !{fullCommunityName(
          community_view.community.name,
          community_view.community.actor_id,
        )}
      {/snippet}
    </EntityHeader>

    <EndPlaceholder size="xs" margin="sm">
      Community
    </EndPlaceholder>
    {#if profile.current?.jwt}
      {@const subscribed = profile.current.user?.follows.some(
        (i) => i.community.id == community_view.community.id,
      )}
      <Button
        disabled={loading.subscribing}
        loading={loading.subscribing}
        size="md"
        color={subscribed ? 'secondary' : 'primary'}
        onclick={() => subscribe(community_view)}
        class="px-4 relative z-[inherit]"
        alignment="left"
        icon={community_view.subscribed == 'Subscribed' ? Check : Plus}
      >
        {subscribed
          ?'Subscribed'
          :'Subscribe'}
      </Button>
      {#if client().setFlair}
        <SidebarButton
          onclick={() => (setFlair = !setFlair)}
          icon={Tag}
          label=Set my flair
        />
      {/if}
    {/if}
    {#if profile.isMod(community_view.community)}
      <SidebarButton
        href="/c/{fullCommunityName(
          community_view.community.name,
          community_view.community.actor_id,
        )}/settings"
        icon={Cog6Tooth}
        label=Edit
      />
    {/if}
    <Menu placement="bottom-start">
      {#snippet target(attachment)}
        <SidebarButton
          {@attach attachment}
          label=More actions
          icon={EllipsisHorizontal}
        />
      {/snippet}
      <MenuButton href="/modlog?community={community_view.community.id}">
        <Icon src={Newspaper} size="16" mini />
        Modlog
      </MenuButton>
      {#if profile.current?.jwt}
        {#if profile.isMod(community_view.community)}
          <MenuButton
            color="success-subtle"
            href="/moderation?community={community_view.community.id}"
          >
            <Icon src={ShieldCheck} size="16" micro />
            Reports
          </MenuButton>
        {/if}
        <MenuButton
          color="danger-subtle"
          size="lg"
          onclick={() =>
            block(community_view.community.id, !community_view.blocked)}
          icon={NoSymbol}
        >
          {community_view.blocked
            ?'Unblock'
            :'Block'}
        </MenuButton>
        {#if profile.current?.user}
          <MenuButton
            color="danger-subtle"
            size="lg"
            onclick={() => blockInstance(community_view.community.instance_id)}
            icon={BuildingOffice2}
          >
            Block server
          </MenuButton>
        {/if}
        {#if profile.isAdmin}
          <MenuButton
            color="danger-subtle"
            onclick={() =>
              modal({
                title:'Purging community',
                body: `${community_view.community.title}: $This will delete all posts. Are you sure? (The button will enable in 3 seconds.)`,
                actions: [
                  action({
                    close: true,
                    content:'Cancel',
                  }),
                  action({
                    action: () => purgeCommunity(community_view.community.id),
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
    <EndPlaceholder size="xs" margin="sm">
      Statistics
    </EndPlaceholder>

    <div class="flex flex-row gap-4 flex-wrap px-3">
      <LabelStat
        label=Members
        content={community_view.counts.subscribers.toString()}
        formatted
      />
      <LabelStat
        label=Posts
        content={community_view.counts.posts.toString()}
        formatted
      />
      <LabelStat
        label=Active Today
        content={community_view.counts.users_active_day.toString()}
        formatted
      />
    </div>

    <EndPlaceholder size="xs" margin="sm">
      Info
    </EndPlaceholder>
    <div class="space-y-3 px-1.5 text-sm">
      <Expandable bind:open={settings.expand.about}>
        {#snippet title()}
          <span class="px-2 py-1 w-full">
            About
          </span>
        {/snippet}
        <Markdown source={community_view.community.description} />
      </Expandable>

      {#if moderators && moderators.length > 0}
        <Expandable bind:open={settings.expand.team}>
          {#snippet title()}
            <span class="px-2 py-1 w-full">
              Moderators
            </span>
          {/snippet}
          <ItemList
            items={moderators.map((i) => ({
              id: i.moderator.id,
              name: i.moderator.display_name || i.moderator.name,
              url: userLink(i.moderator),
              avatar: i.moderator.avatar,
              instance: new URL(i.moderator.actor_id).hostname,
            }))}
          />
        </Expandable>
      {/if}
    </div>
  </aside>
{/await}

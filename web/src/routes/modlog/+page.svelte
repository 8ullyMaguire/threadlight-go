<script lang="ts">
  import { browser } from '$app/environment'
  import { page } from '$app/state'
  import { client } from '$lib/api/client.svelte'
  import { ThreadlightClient } from '$lib/api/threadlight/adapter'
  import { profile } from '$lib/app/auth'
  import { errorMessage } from '$lib/app/error'

  import { settings } from '$lib/app/settings.svelte'
  import { searchParam } from '$lib/app/util.svelte'
  import CommunityLink from '$lib/feature/community/CommunityLink.svelte'
  import { postLink } from '$lib/feature/post'
  import UserAutocomplete from '$lib/feature/user/UserAutocomplete.svelte'
  import UserLink from '$lib/feature/user/UserLink.svelte'
  import ObjectAutocomplete from '$lib/ui/form/ObjectAutocomplete.svelte'
  import Placeholder from '$lib/ui/info/Placeholder.svelte'
  import { Header, Pageination } from '$lib/ui/layout'
  import { Button, Option, Select, Spinner, toast } from 'mono-svelte'
  import {
    Bars3BottomRight,
    CheckCircle,
    HandThumbDown,
    HandThumbUp,
    Icon,
    MagnifyingGlass,
    ViewColumns,
    XCircle,
    XMark,
  } from 'svelte-hero-icons/dist'
  import ModlogItemCard from './item/ModlogItemCard.svelte'
  import ModlogItemTable from './item/ModlogItemTable.svelte'
  import type { ModLog } from './+page'

  let { data = $bindable() } = $props()

  let view = $state(
    `${
      (settings.modlogCardView ?? browser)
        ? !window.matchMedia('(min-width: 1600px)').matches
        : false
    }`,
  )

  $effect(() => {
    settings.modlogCardView =
      view == 'true' ? true : view == 'false' ? false : undefined
  })

  // ── Voting state ─────────────────────────────────────────
  let votingState = $state<Record<number, 'idle' | 'voting' | 'done' | 'error'>>({})
  let voteMessages = $state<Record<number, string>>({})
  let trustPenalties = $state<Record<number, string>>({})
  let userVotes = $state<Record<number, number>>({})

  async function castVote(actionId: number, vote: number) {
    if (!profile.current?.jwt) {
      toast({ content: 'You must be logged in to vote.', type: 'warning' })
      return
    }
    votingState[actionId] = 'voting'
    try {
      const apiClient = client() as unknown as ThreadlightClient
      const res = await apiClient.castReview(actionId, vote)
      votingState[actionId] = 'done'
      userVotes[actionId] = vote
      if (res.trust_penalty) {
        trustPenalties[actionId] = res.trust_penalty
      }
      voteMessages[actionId] = res.message || 'Vote recorded.'
      if (res.trust_penalty) {
        toast({ content: `Trust penalty: ${res.trust_penalty}`, type: 'warning' })
      }
      toast({ content: res.message || 'Vote recorded.', type: 'success' })
    } catch (err) {
      votingState[actionId] = 'error'
      voteMessages[actionId] = errorMessage(err as string)
      toast({ content: errorMessage(err as string), type: 'error' })
    }
  }

  function getDecision(actionId: number | undefined) {
    if (actionId === undefined || actionId === null) return null
    return data.decisionsById?.[actionId] ?? null
  }
</script>

<svelte:head>
  <title>Modlog</title>
</svelte:head>

<div class="flex flex-col gap-4">
  <Header pageHeader>
    Modlog
    {#snippet extended()}
      <span class="font-medium text-lg">Filters</span>
      <ul class="font-normal flex flex-col gap-2 mt-2">
        {#if data.params.community}
          <li>
            <div class="text-sm text-slate-600 dark:text-zinc-400">
              Community
            </div>
            {#await client().getCommunity({ id: data.params.community })}
              <Spinner width={24} />
            {:then community}
              <CommunityLink community={community.community_view.community} />
            {/await}
          </li>
        {/if}
        {#if data.params.user}
          <li>
            <div class="text-sm text-slate-600 dark:text-zinc-400">
              User
            </div>
            {#await client().getPersonDetails( { person_id: data.params.user, limit: 1 }, )}
              <Spinner width={24} />
            {:then person}
              <UserLink class="inline" user={person.person_view.person} />
            {/await}
          </li>
        {/if}
        {#if data.params.moderator}
          <li>
            <div class="text-sm text-slate-600 dark:text-zinc-400">
              <!--TODO add translation key-->
              Moderator
            </div>
            {#await client().getPersonDetails( { person_id: data.params.moderator, limit: 1 }, )}
              <Spinner width={24} />
            {:then person}
              <UserLink class="inline" user={person.person_view.person} />
            {/await}
          </li>
        {/if}
        {#if data.params.post}
          <li>
            <div class="text-sm text-slate-600 dark:text-zinc-400">
              Post
            </div>
            {#await client().getPost({ id: data.params.post })}
              <Spinner width={24} />
            {:then post}
              <a
                class="hover:underline block"
                href={postLink(post.post_view.post)}
              >
                {post.post_view.post.name}
              </a>
            {/await}
          </li>
        {/if}
        {#if data.params.comment}
          <li>
            <div class="text-sm text-slate-600 dark:text-zinc-400">
              Comment
            </div>
            {#await client().getComment({ id: data.params.comment })}
              <Spinner width={24} />
            {:then comment}
              <a
                class="hover:underline block"
                href="/comment/{data.params.comment}"
              >
                {comment.comment_view.comment.content.slice(1, 200)}...
              </a>
            {/await}
          </li>
        {/if}
      </ul>
    {/snippet}
  </Header>
  <div class="flex flex-row flex-wrap gap-2">
    <Select
      bind:value={data.type}
      onchange={() => searchParam(page.url, 'type', data.type, 'page')}
      class="w-48"
    >
      {#snippet customLabel()}
        <span class="flex gap-1 items-center">
          <Icon src={Bars3BottomRight} size="15" mini />
          Type
        </span>
      {/snippet}
      <Option value="All">All</Option>
      <Option value="ModRemovePost">Remove post</Option>
      <Option value="ModLockPost">Lock post</Option>
      <Option value="ModFeaturePost">Feature post</Option>
      <Option value="ModRemoveComment">Remove comment</Option>
      <Option value="ModRemoveCommunity">Remove community</Option>
      <Option value="ModBanFromCommunity">Ban from community</Option>
      <Option value="ModAddCommunity">Add moderator</Option>
      <Option value="ModTransferCommunity">Transfer community</Option>
      <Option value="ModAdd">Add admin</Option>
      <Option value="ModBan">Ban admin</Option>
      <Option value="ModHideCommunity">Hide community</Option>
      <Option value="AdminPurgePerson">Purge user</Option>
      <Option value="AdminPurgeCommunity">Purge community</Option>
      <Option value="AdminPurgePost">Purge post</Option>
      <Option value="AdminPurgeComment">Purge comment</Option>
    </Select>

    <Select bind:value={view} class="w-36">
      {#snippet customLabel()}
        <span class="flex gap-1 items-center">
          <Icon src={ViewColumns} size="15" mini />
          View
        </span>
      {/snippet}
      <Option value="false">Table</Option>
      <Option value="true">Cards</Option>
      <Option value="undefined">Default</Option>
    </Select>
  </div>
  <div class="flex flex-col md:flex-row md:items-center gap-2 w-full">
    <ObjectAutocomplete
      placeholder="Filter by community"
      listing_type="All"
      showWhenEmpty={true}
      label="Community"
      q={page.url.searchParams.get('community') ? 'Selected' : ''}
      onselect={(e) =>
        searchParam(
          page.url,
          'community',
          e?.community.id.toString() ?? '',
          'page',
        )}
    />
    <UserAutocomplete
      instance={page.url.searchParams.get('instance') || undefined}
      placeholder="Filter by user"
      listing_type="All"
      showWhenEmpty={true}
      label="User"
      q={page.url.searchParams.get('user')
        ? (data.filters.user ?? 'Selected')
        : ''}
      onselect={(e) =>
        searchParam(page.url, 'user', e?.id.toString() ?? '', 'page')}
    />
    {#if profile.isAdmin}
      <UserAutocomplete
        placeholder="Filter by moderator"
        listing_type="All"
        showWhenEmpty={true}
        label="Moderator"
        q={page.url.searchParams.get('mod_id')
          ? (data.filters.moderator ?? 'Selected')
          : ''}
        onselect={(e) =>
          searchParam(page.url, 'mod_id', e?.id.toString() ?? '', 'page')}
      />
    {/if}
    <Button
      onclick={() => {
        searchParam(
          page.url,
          '',
          '',
          'user',
          'moderator',
          'community',
          'instance',
        )
      }}
      size="custom"
      class="self-end shrink-0 h-8.5 aspect-square"
      title="Clear filters"
    >
      <Icon src={XMark} size="16" mini />
    </Button>
  </div>
  {#if data.modlog && data.modlog.length > 0}
    {#if settings.modlogCardView ?? !window.matchMedia('(min-width: 1600px)').matches}
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        {#each data.modlog as modlog (modlog)}
          {@const decision = getDecision(modlog.id)}
          <div class="flex flex-col gap-2">
            <ModlogItemCard item={modlog} />
            {#if decision}
              <div
                class="flex flex-row items-center gap-3 px-3 py-2 text-xs rounded-xl bg-slate-50 dark:bg-zinc-900 border border-slate-200 dark:border-zinc-800"
              >
                <span class="font-medium text-slate-500 dark:text-zinc-400">
                  Reviews:
                </span>
                <span class="flex items-center gap-1 text-green-600 dark:text-green-400">
                  <Icon src={HandThumbUp} size="14" mini />
                  {decision.reviews_fair}
                </span>
                <span class="flex items-center gap-1 text-red-600 dark:text-red-400">
                  <Icon src={HandThumbDown} size="14" mini />
                  {decision.reviews_unfair}
                </span>
                <span class="text-slate-400 dark:text-zinc-500">|</span>
                <span class="flex items-center gap-1">
                  <span class="font-medium">Score:</span>
                  <span class={decision.controversy_score > 0 ? 'text-orange-500' : 'text-slate-500'}>
                    {decision.controversy_score.toFixed(2)}
                  </span>
                </span>
                <div class="flex-1"></div>
                {#if profile.current?.jwt}
                  {#if userVotes[decision.action_id]}
                    <span class="text-xs text-slate-500 dark:text-zinc-400">
                      {userVotes[decision.action_id] === 1 ? '✓ Voted Fair' : '✗ Voted Unfair'}
                    </span>
                  {:else if votingState[decision.action_id] === 'voting'}
                    <Spinner width={16} />
                  {:else if votingState[decision.action_id] === 'done'}
                    <span class="text-xs text-green-600 dark:text-green-400">Voted</span>
                  {:else}
                    <button
                      onclick={() => castVote(decision.action_id, 1)}
                      class="cursor-pointer hover:text-green-500 transition-colors"
                      title="Vote Fair"
                    >
                      <Icon src={CheckCircle} size="16" mini />
                    </button>
                    <button
                      onclick={() => castVote(decision.action_id, -1)}
                      class="cursor-pointer hover:text-red-500 transition-colors"
                      title="Vote Unfair"
                    >
                      <Icon src={XCircle} size="16" mini />
                    </button>
                  {/if}
                {/if}
              </div>
            {/if}
            {#if decision && trustPenalties[decision.action_id]}
              <div
                class="px-3 py-1.5 text-xs rounded-xl bg-orange-50 dark:bg-orange-950/30 border border-orange-200 dark:border-orange-800 text-orange-700 dark:text-orange-300"
              >
                Trust penalty: {trustPenalties[decision.action_id]}
              </div>
            {/if}
          </div>
        {/each}
      </div>
    {:else}
      <div style="width:100%; overflow-x: auto;" class="table-container">
        <table
          class="table overflow-x-auto table-fixed relative w-full min-w-2xl"
        >
          <colgroup>
            <col style="width: 10.6%;" />
            <col style="width: 12%;" />
            <col style="width: 12%;" />
            <col style="width: 12%;" />
            <col style="width: 12%;" />
            <col style="width: 12%;" />
            <col style="width: 14%;" />
            <col style="width: 15.4%;" />
          </colgroup>
          <thead class="text-left">
            <tr class="rounded-t-lg overflow-hidden">
              <th>Time</th>
              <th>Moderator</th>
              <th>Community</th>
              <th>Action</th>
              <th>Reason</th>
              <th align="right">Content</th>
              <th align="center">Reviews</th>
              <th align="center">Vote</th>
            </tr>
          </thead>
          <tbody class="text-sm divide-y divide-slate-200 dark:divide-zinc-800">
            {#each data.modlog as modlog (modlog)}
              {@const decision = getDecision(modlog.id)}
              <ModlogItemTable
                item={modlog}
                decision={decision}
                userVote={decision ? userVotes[decision.action_id] : undefined}
                voting={decision ? votingState[decision.action_id] === 'voting' : false}
                trustPenalty={decision ? trustPenalties[decision.action_id] : undefined}
                onvote={(vote: number) => decision && castVote(decision.action_id, vote)}
              />
            {/each}
          </tbody>
        </table>
      </div>
    {/if}
    <Pageination
      page={data.page}
      hasMore={data.params.hasMore}
      href={(page) => `?page=${page}`}
    />
  {:else}
    <Placeholder
      title="No results"
      description="There are no mod logs with that filter. Try refining your search."
      icon={MagnifyingGlass}
    />
  {/if}
</div>

<style>
  @reference '../../app.css';

  .table-container {
    border: 1px solid var(--color-slate-200);
    background-color: white;
    border-radius: var(--radius-2xl);

    @variant dark {
      background-color: var(--color-zinc-950);
      border: 1px solid var(--color-zinc-800);
    }
  }

  .table {
    table-layout: fixed !important;

    thead {
      tr {
        th {
          padding: calc(var(--spacing) * 3);
          border-bottom: 1px solid var(--color-slate-200);
          font-weight: var(--font-weight-medium);
          font-size: var(--text-sm);
          background-color: var(--color-slate-25);

          @variant dark {
            background-color: var(--color-zinc-900);
            border-bottom: 1px solid var(--color-zinc-800);
          }
        }
      }
    }
  }

  .table tbody {
    font-size: var(--text-xs);
    :global(tr td) {
      padding: calc(var(--spacing) * 1) calc(var(--spacing) * 3);
    }
  }
</style>

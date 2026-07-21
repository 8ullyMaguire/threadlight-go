<script lang="ts">

  import CommunityLink from '$lib/feature/community/CommunityLink.svelte'
  import UserLink from '$lib/feature/user/UserLink.svelte'
  import { Button, modal, Spinner } from 'mono-svelte'
  import RelativeDate from 'mono-svelte/util/RelativeDate.svelte'
  import { CheckCircle, HandThumbDown, HandThumbUp, Icon, XCircle } from 'svelte-hero-icons/dist'
  import type { ModLog } from '../+page'
  import ModlogAction from '../ModlogAction.svelte'
  import type { ControversialDecision } from '$lib/api/threadlight/types'

  interface Props {
    item: ModLog
    decision?: ControversialDecision | null
    userVote?: number
    voting?: boolean
    trustPenalty?: number
    onvote?: (vote: number) => void
  }

  let { item, decision = null, userVote = undefined, voting = false, trustPenalty = undefined, onvote = undefined }: Props = $props()
</script>

{#snippet itemInfo()}
  {#if item.link}
    <Button color="primary" rounding="pill" class="w-max" href={item.link}>
      Jump
    </Button>
  {/if}
  {#if item.moderatee}
    <UserLink avatar user={item.moderatee} />
  {/if}
  <div>
    {item.content}
  </div>
{/snippet}

<tr class="">
  <td style="width: 10%;">
    <span>
      <RelativeDate date={new Date(item.timestamp)} />
    </span>
  </td>
  <td>
    {#if item.moderator}
      <UserLink
        showInstance={false}
        avatar
        avatarSize={20}
        user={item.moderator}
      />
    {:else}
      <p class="text-slate-500 dark:text-zinc-500">Unknown</p>
      <p></p>
    {/if}
  </td>
  <td>
    {#if item.community}
      <CommunityLink
        showInstance={false}
        avatar
        avatarSize={20}
        community={item.community}
      />
    {/if}
  </td>
  <td>
    <ModlogAction action={item.actionName} />
  </td>
  <td>
    {#if item.reason}
      <p>{item.reason}</p>
    {/if}
  </td>
  <td align="right">
    {#if item.content}
      <Button
        size="sm"
        onclick={() => modal({ title:'Info', snippet: itemInfo })}
      >
        Info
      </Button>
    {:else if item.moderatee}
      <UserLink user={item.moderatee} />
    {/if}
  </td>
  <td align="center">
    {#if decision}
      <div class="flex flex-row items-center gap-2 justify-center">
        <span class="flex items-center gap-0.5 text-green-600 dark:text-green-400 text-xs" title="Fair reviews">
          <Icon src={HandThumbUp} size="12" mini />
          {decision.reviews_fair}
        </span>
        <span class="flex items-center gap-0.5 text-red-600 dark:text-red-400 text-xs" title="Unfair reviews">
          <Icon src={HandThumbDown} size="12" mini />
          {decision.reviews_unfair}
        </span>
        <span class="text-[10px] font-mono {decision.controversy_score > 0.5 ? 'text-orange-500' : 'text-slate-400'}">
          {decision.controversy_score.toFixed(2)}
        </span>
      </div>
    {:else}
      <span class="text-slate-400 dark:text-zinc-500 text-xs">—</span>
    {/if}
  </td>
  <td align="center">
    {#if decision && onvote}
      {#if userVote}
        <span class="text-xs text-slate-500 dark:text-zinc-400">
          {userVote === 1 ? '✓ Fair' : '✗ Unfair'}
        </span>
      {:else if voting}
        <Spinner width={14} />
      {:else}
        <div class="flex flex-row items-center gap-1 justify-center">
          <button
            onclick={() => onvote(1)}
            class="cursor-pointer hover:text-green-500 transition-colors p-0.5"
            title="Vote Fair"
          >
            <Icon src={CheckCircle} size="16" mini />
          </button>
          <button
            onclick={() => onvote(-1)}
            class="cursor-pointer hover:text-red-500 transition-colors p-0.5"
            title="Vote Unfair"
          >
            <Icon src={XCircle} size="16" mini />
          </button>
        </div>
      {/if}
    {:else if !onvote}
      <span class="text-slate-400 dark:text-zinc-500 text-xs">—</span>
    {/if}
    {#if trustPenalty}
      <div class="text-[10px] text-orange-500 mt-0.5">
        Penalty: {trustPenalty}
      </div>
    {/if}
  </td>
</tr>

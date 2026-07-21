<script lang="ts">
  import { client } from '$lib/api/client.svelte'
  import { profile } from '$lib/app/auth'
  import { errorMessage } from '$lib/app/error'
  import { t } from '$lib/app/i18n'
  import FormattedNumber from '$lib/ui/util/FormattedNumber.svelte'
  import { toast } from 'mono-svelte'
  import { ChevronDown, ChevronUp, Icon } from 'svelte-hero-icons/dist'
  import { backOut } from 'svelte/easing'
  import { fly } from 'svelte/transition'

  interface Props {
    tagId: number
    name: string
    totalVotes?: number
    voteScore?: number
    userVote?: number
  }

  let {
    tagId,
    name,
    totalVotes = $bindable(0),
    voteScore = $bindable(0),
    userVote = $bindable(0),
  }: Props = $props()

  let pending = $state(false)

  const voteOnTag = async (newVote: number) => {
    if (navigator.vibrate) navigator.vibrate(1)
    if (!profile.current?.jwt) {
      toast({ content: $t('toast.loginVoteGate'), type: 'warning' })
      return
    }

    // Optimistic update
    const prevVote = userVote ?? 0
    const prevScore = voteScore ?? 0
    const prevTotal = totalVotes ?? 0

    if (prevVote === newVote) {
      // Same vote -> remove it
      userVote = 0
      voteScore = newVote === 1 ? prevScore - 1 : prevScore + 1
      totalVotes = Math.max(0, prevTotal - 1)
    } else {
      // New vote
      if (prevVote === 0) {
        totalVotes = prevTotal + 1
      }
      if (prevVote === 1) {
        voteScore = prevScore - 1
      } else if (prevVote === -1) {
        voteScore = prevScore + 1
      }
      userVote = newVote
      voteScore = newVote === 1 ? (voteScore ?? 0) + 1 : (voteScore ?? 0) - 1
    }

    pending = true

    try {
      await (client() as any).voteOnTag(tagId, newVote === prevVote ? 0 : newVote)
    } catch (e) {
      // Revert on error
      userVote = prevVote
      voteScore = prevScore
      totalVotes = prevTotal
      toast({ content: errorMessage(e), type: 'error' })
    } finally {
      pending = false
    }
  }
</script>

<div class="flex items-center gap-1 tag-vote">
  <button
    onclick={() => voteOnTag(1)}
    disabled={pending}
    class={[
      'flex items-center gap-0.5 transition-colors cursor-pointer px-1.5 py-0.5 rounded-l-full',
      'border border-slate-200 dark:border-zinc-700',
      userVote === 1
        ? 'bg-primary-500 text-white border-primary-500'
        : 'btn-tertiary hover:bg-primary-100 dark:hover:bg-primary-900',
    ]}
    aria-pressed={userVote === 1}
    aria-label={$t('post.actions.vote.upvote')}
  >
    <Icon src={ChevronUp} size="14" micro />
    <span class="text-xs font-medium tabular-nums">
      {#key voteScore}
        <span
          style="grid-column: 1; grid-row: 1;"
          in:fly={{ duration: 300, y: -8, easing: backOut }}
          out:fly={{ duration: 300, y: 8, easing: backOut }}
        >
          <FormattedNumber number={voteScore ?? 0} />
        </span>
      {/key}
    </span>
  </button>
  <span class="text-xs text-slate-400 dark:text-zinc-500 px-1">
    <FormattedNumber number={totalVotes ?? 0} />
  </span>
  <button
    onclick={() => voteOnTag(-1)}
    disabled={pending}
    class={[
      'flex items-center gap-0.5 transition-colors cursor-pointer px-1.5 py-0.5 rounded-r-full',
      'border border-slate-200 dark:border-zinc-700',
      userVote === -1
        ? 'bg-red-500 text-white border-red-500'
        : 'btn-tertiary hover:bg-red-100 dark:hover:bg-red-900',
    ]}
    aria-pressed={userVote === -1}
    aria-label={$t('post.actions.vote.downvote')}
  >
    <Icon src={ChevronDown} size="14" micro />
  </button>
</div>

<style>
  .tag-vote button:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
</style>

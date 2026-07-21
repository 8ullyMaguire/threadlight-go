<script lang="ts">
  import { client } from '$lib/api/client.svelte'
  import { ThreadlightClient } from '$lib/api/threadlight/adapter'
  import type { CreditBalance } from '$lib/api/threadlight/types'
  import { profile } from '$lib/app/auth'
  import { onMount } from 'svelte'
  import { CheckCircle, Fire, Icon, Trophy, XCircle } from 'svelte-hero-icons/dist'
  import { Spinner } from 'mono-svelte'

  interface Props {
    compact?: boolean
  }

  let { compact = false }: Props = $props()

  let balance = $state<CreditBalance | null>(null)
  let loading = $state(true)
  let error = $state<string | null>(null)

  const QUEST_LABELS: Record<string, string> = {
    vote_tag: 'Vote on a tag',
    react_post: 'React to a post',
    add_tag: 'Add a tag to a post',
    vote_note: 'Vote on a community note',
  }

  const QUEST_REWARDS: Record<string, number> = {
    vote_tag: 5,
    react_post: 3,
    add_tag: 10,
    vote_note: 5,
  }

  let allComplete = $derived(
    balance?.quests?.length > 0 && balance.quests.every((q) => q.completed),
  )

  onMount(async () => {
    if (!profile.current?.user) {
      loading = false
      return
    }
    try {
      const tl = client() as unknown as ThreadlightClient
      balance = await tl.getCreditBalance()
    } catch (e) {
      error = String(e)
    } finally {
      loading = false
    }
  })

  function refresh() {
    loading = true
    error = null
    const tl = client() as unknown as ThreadlightClient
    tl.getCreditBalance()
      .then((b) => {
        balance = b
        loading = false
      })
      .catch((e) => {
        error = String(e)
        loading = false
      })
  }
</script>

{#if loading}
  <div class="flex justify-center p-4">
    <Spinner width={20} />
  </div>
{:else if error}
  <div class="text-red-500 text-sm p-2">Failed to load quests</div>
{:else if balance}
  <div class="rounded-xl border border-slate-200 dark:border-zinc-700 bg-white dark:bg-zinc-900 p-4 space-y-3">
    <div class="flex items-center justify-between">
      <h3 class="font-semibold text-sm flex items-center gap-1.5">
        <Icon src={Trophy} size="16" micro />
        Daily Quests
      </h3>
      <div class="flex items-center gap-1.5 text-sm" class:text-amber-500={balance.streak > 0}>
        <Icon src={Fire} size="14" micro />
        <span class="font-medium">{balance.streak} day streak</span>
      </div>
    </div>

    {#if !compact}
      <div class="space-y-2">
        {#each balance.quests as quest (quest.quest_type)}
          <div
            class={[
              'flex items-center gap-2 text-sm p-2 rounded-lg transition-colors',
              quest.completed ? 'bg-green-50 dark:bg-green-950/30' : '',
            ].join(' ')}
          >
            {#if quest.completed}
              <Icon src={CheckCircle} size="18" class="text-green-500 shrink-0" />
            {:else}
              <Icon src={XCircle} size="18" class="text-zinc-400 shrink-0" />
            {/if}
            <span
              class={[
                'flex-1',
                quest.completed ? 'line-through text-zinc-500' : '',
              ].join(' ')}
            >
              {QUEST_LABELS[quest.quest_type] ?? quest.description}
            </span>
            <span class="text-xs font-medium text-primary-600 dark:text-primary-400">
              +{QUEST_REWARDS[quest.quest_type] ?? quest.reward}
            </span>
          </div>
        {/each}
      </div>

      {#if allComplete}
        <div class="text-center text-xs font-medium text-green-600 dark:text-green-400 bg-green-50 dark:bg-green-950/30 rounded-lg py-2">
          All quests complete! Bonus streak maintained.
        </div>
      {:else if balance.quests.length > 0}
        <div class="text-center text-xs text-zinc-500">
          Complete all 4 quests to maintain your streak!
        </div>
      {:else}
        <div class="text-center text-xs text-zinc-500">
          No quests available today. Check back tomorrow!
        </div>
      {/if}

      <button
        onclick={refresh}
        class="text-xs text-primary-600 dark:text-primary-400 hover:underline w-full text-center"
      >
        Refresh
      </button>
    {/if}
  </div>
{/if}

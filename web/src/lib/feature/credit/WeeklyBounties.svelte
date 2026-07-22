<script lang="ts">
  import { client } from '$lib/api/client.svelte'
  import { ThreadlightClient } from '$lib/api/threadlight/adapter'
  import type { BountyConfig } from '$lib/api/threadlight/types'
  import { onMount } from 'svelte'
  import { Spinner } from 'mono-svelte'
  import { Trophy, Icon, Clock } from 'svelte-hero-icons/dist'

  let bountyConfig = $state<BountyConfig | null>(null)
  let loading = $state(true)
  let error = $state<string | null>(null)
  let countdown = $state('')

  const CATEGORIES: { key: keyof BountyConfig; label: string; emoji: string }[] = [
    { key: 'weekly_bounty_poster', label: 'Poster', emoji: '📝' },
    { key: 'weekly_bounty_tagger', label: 'Tagger', emoji: '🏷️' },
    { key: 'weekly_bounty_commenter', label: 'Commenter', emoji: '💬' },
    { key: 'weekly_bounty_curator', label: 'Curator', emoji: '⭐' },
  ]

  onMount(() => {
    loadBounties()

    updateCountdown()
    const interval = setInterval(updateCountdown, 60000)
    return () => clearInterval(interval)
  })

  async function loadBounties() {
    try {
      const tl = client() as unknown as ThreadlightClient
      bountyConfig = await tl.getWeeklyBountyConfig()
    } catch (e) {
      error = String(e)
    } finally {
      loading = false
    }
  }

  function updateCountdown() {
    const now = new Date()
    // Next Monday 00:00 UTC
    const next = new Date(now)
    next.setUTCDate(next.getUTCDate() + ((8 - next.getUTCDay()) % 7 || 7))
    next.setUTCHours(0, 0, 0, 0)
    const diff = next.getTime() - now.getTime()
    const days = Math.floor(diff / (1000 * 60 * 60 * 24))
    const hours = Math.floor((diff % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60))
    const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60))
    if (diff <= 0) {
      countdown = 'Distributing now…'
    } else {
      const parts: string[] = []
      if (days > 0) parts.push(`${days}d`)
      if (hours > 0) parts.push(`${hours}h`)
      parts.push(`${minutes}m`)
      countdown = parts.join(' ')
    }
  }
</script>

{#if loading}
  <div class="flex justify-center p-4">
    <Spinner width={20} />
  </div>
{:else if error}
  <div class="text-red-500 text-sm p-2">Failed to load bounties</div>
{:else if bountyConfig}
  <div class="rounded-xl border border-slate-200 dark:border-zinc-700 bg-white dark:bg-zinc-900 p-4 space-y-3">
    <div class="flex items-center justify-between">
      <h3 class="font-semibold text-sm flex items-center gap-1.5">
        <Icon src={Trophy} size="16" micro />
        Weekly Bounties
      </h3>
      <div class="flex items-center gap-1 text-xs text-zinc-500">
        <Icon src={Clock} size="12" micro />
        <span>{countdown}</span>
      </div>
    </div>

    <div class="space-y-2">
      {#each CATEGORIES as cat (cat.key)}
        <div
          class="flex items-center justify-between text-sm p-2 rounded-lg bg-zinc-50 dark:bg-zinc-950/50"
        >
          <span>
            <span class="mr-1.5">{cat.emoji}</span>
            {cat.label}
          </span>
          <span class="font-semibold text-primary-600 dark:text-primary-400">
            {bountyConfig[cat.key]} credits
          </span>
        </div>
      {/each}
    </div>

    <div class="text-xs text-zinc-500 text-center">
      Top contributor in each category wins the bounty at the end of the week.
    </div>
  </div>
{/if}

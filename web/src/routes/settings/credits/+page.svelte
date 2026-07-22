<script lang="ts">
  import { client } from '$lib/api/client.svelte'
  import { ThreadlightClient } from '$lib/api/threadlight/adapter'
  import type { CreditBalance, CreditTransactionsResponse, CreditCosts } from '$lib/api/threadlight/types'
  import { profile } from '$lib/app/auth'
  import { onMount } from 'svelte'
  import DailyQuests from '$lib/feature/credit/DailyQuests.svelte'
  import WeeklyBounties from '$lib/feature/credit/WeeklyBounties.svelte'
  import { Spinner, Button } from 'mono-svelte'
  import { ArrowPath, Fire, Icon, CurrencyDollar } from 'svelte-hero-icons/dist'
  import { Header } from '$lib/ui/layout'

  let balance = $state<CreditBalance | null>(null)
  let transactions = $state<CreditTransactionsResponse | null>(null)
  let costs = $state<CreditCosts | null>(null)
  let loading = $state(true)
  let error = $state<string | null>(null)
  let txPage = $state(1)

  onMount(() => {
    loadData()
  })

  async function loadData() {
    loading = true
    error = null
    try {
      const tl = client() as unknown as ThreadlightClient
      const [bal, txs, cst] = await Promise.all([
        tl.getCreditBalance(),
        tl.getCreditTransactions({ page: txPage, limit: 20 }),
        tl.getCreditCosts(),
      ])
      balance = bal
      transactions = txs
      costs = cst
    } catch (e) {
      error = String(e)
    } finally {
      loading = false
    }
  }
</script>

<svelte:head>
  <title>Credits</title>
</svelte:head>

<Header pageHeader>
  Credits
</Header>

{#if loading && !balance}
  <div class="flex justify-center p-8">
    <Spinner width={24} />
  </div>
{:else if error}
  <div class="text-red-500 p-4">Failed to load credit data: {error}</div>
{:else}
  <div class="space-y-6 max-w-2xl">
    <!-- Balance card -->
    <div class="rounded-xl border border-slate-200 dark:border-zinc-700 bg-white dark:bg-zinc-900 p-6">
      <div class="flex items-center gap-3">
        <div class="w-12 h-12 rounded-full bg-amber-100 dark:bg-amber-900/50 flex items-center justify-center">
          <Icon src={CurrencyDollar} size="24" class="text-amber-600 dark:text-amber-400" />
        </div>
        <div>
          <div class="text-3xl font-bold">{balance?.credits ?? 0}</div>
          <div class="text-sm text-zinc-500">Available credits</div>
        </div>
      </div>
      {#if balance}
        <div class="flex gap-4 mt-4 text-sm text-zinc-500">
          <div>
            <span class="font-medium text-zinc-700 dark:text-zinc-300">{balance.lifetime_credits}</span> lifetime earned
          </div>
          <div class="flex items-center gap-1" class:text-amber-500={balance.streak > 0}>
            <Icon src={Fire} size="14" micro />
            <span class="font-medium">{balance.streak} day streak</span>
          </div>
        </div>
      {/if}
    </div>

    <!-- Daily quests -->
    <DailyQuests />

    <!-- Weekly bounties -->
    <WeeklyBounties />

    <!-- Credit transfer tax notice -->
    <div class="rounded-xl border border-amber-200 dark:border-amber-800 bg-amber-50 dark:bg-amber-950/30 p-4 text-sm">
      <p>
        <strong>Transfer tax:</strong> Transfers are subject to a 10% platform tax.
        Sending 100 credits delivers 90 to the recipient.
      </p>
    </div>

    <!-- Action costs -->
    {#if costs}
      <div class="rounded-xl border border-slate-200 dark:border-zinc-700 bg-white dark:bg-zinc-900 p-4">
        <h3 class="font-semibold text-sm mb-3">Action Costs</h3>
        <table class="w-full text-sm">
          <thead>
            <tr class="text-left text-zinc-500 border-b border-slate-200 dark:border-zinc-700">
              <th class="pb-2 font-medium">Action</th>
              <th class="pb-2 font-medium text-right">Cost</th>
            </tr>
          </thead>
          <tbody>
            <tr class="border-b border-slate-100 dark:border-zinc-800">
              <td class="py-2">Post creation</td>
              <td class="py-2 text-right">{costs.post_creation}</td>
            </tr>
            <tr class="border-b border-slate-100 dark:border-zinc-800">
              <td class="py-2">Image upload</td>
              <td class="py-2 text-right">{costs.image_upload}</td>
            </tr>
            <tr class="border-b border-slate-100 dark:border-zinc-800">
              <td class="py-2">Search</td>
              <td class="py-2 text-right">{costs.search}</td>
            </tr>
            <tr class="border-b border-slate-100 dark:border-zinc-800">
              <td class="py-2">Private message</td>
              <td class="py-2 text-right">{costs.private_message}</td>
            </tr>
            <tr>
              <td class="py-2">Reaction</td>
              <td class="py-2 text-right">{costs.reaction}</td>
            </tr>
          </tbody>
        </table>
      </div>
    {/if}

    <!-- Recent transactions -->
    {#if transactions?.transactions?.length}
      <div class="rounded-xl border border-slate-200 dark:border-zinc-700 bg-white dark:bg-zinc-900 p-4">
        <h3 class="font-semibold text-sm mb-3">Recent Transactions</h3>
        <div class="space-y-2">
          {#each transactions.transactions as tx (tx.id)}
            <div class="flex items-center justify-between text-sm py-1.5 border-b border-slate-100 dark:border-zinc-800 last:border-0">
              <div>
                <div class="font-medium">{tx.reason}</div>
                <div class="text-xs text-zinc-500">{new Date(tx.created_at).toLocaleDateString()}</div>
              </div>
              <div class="font-medium" class:text-green-600={tx.amount > 0} class:text-red-500={tx.amount < 0}>
                {tx.amount > 0 ? '+' : ''}{tx.amount}
              </div>
            </div>
          {/each}
        </div>
      </div>
    {/if}

    <Button onclick={loadData} icon={ArrowPath}>Refresh</Button>
  </div>
{/if}

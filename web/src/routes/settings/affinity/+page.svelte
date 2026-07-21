<script lang="ts">
  import type { PageData } from './+page.ts'
  import { client } from '$lib/api/client.svelte'
  import { Button, TextInput } from 'mono-svelte'
  import type { AffinityEntry } from '$lib/api/threadlight/types'

  let { data } = $props<{ data: PageData }>()

  let activeTab = $state<'my' | 'similar'>('my')
  let minScore = $state('0.5')
  let similarUsers = $state<number[]>([])
  let loadingSimilar = $state(false)

  function getBreakdown(breakdownStr: string) {
    try {
      return JSON.parse(breakdownStr) as {
        tag_overlap: number
        co_community: number
        reaction_agreement: number
        trust_distance: number
      }
    } catch {
      return { tag_overlap: 0, co_community: 0, reaction_agreement: 0, trust_distance: 0 }
    }
  }

  async function loadSimilarUsers() {
    loadingSimilar = true
    try {
      const score = parseFloat(minScore) || 0.5
      const result = await client().getSimilarUsers(score)
      similarUsers = result.user_ids
    } catch (err) {
      console.error('Failed to load similar users', err)
      similarUsers = []
    } finally {
      loadingSimilar = false
    }
  }

  $effect(() => {
    if (activeTab === 'similar') {
      loadSimilarUsers()
    }
  })
</script>

<svelte:head>
  <title>Affinity - Settings</title>
</svelte:head>

<div class="p-4">
  <h2 class="text-2xl font-semibold mb-6">User Affinity</h2>

  <div class="flex gap-1 mb-6 border-b border-border">
    <button
      onclick={() => activeTab = 'my'}
      class="px-4 py-2 text-sm font-medium transition-colors border-b-2 -mb-px {activeTab === 'my' ? 'border-primary text-primary' : 'border-transparent text-muted-foreground hover:text-foreground'}"
    >
      My Affinities
    </button>
    <button
      onclick={() => activeTab = 'similar'}
      class="px-4 py-2 text-sm font-medium transition-colors border-b-2 -mb-px {activeTab === 'similar' ? 'border-primary text-primary' : 'border-transparent text-muted-foreground hover:text-foreground'}"
    >
      Similar Users
    </button>
  </div>

  {#if activeTab === 'my'}
    {#if data.affinities.length === 0}
      <p class="text-muted-foreground mt-4">No affinity data available yet.</p>
    {:else}
      <div class="overflow-x-auto mt-4">
        <table class="w-full border-collapse">
          <thead>
            <tr class="border-b border-border">
              <th class="text-left py-2 px-3 font-medium text-sm">User ID</th>
              <th class="text-left py-2 px-3 font-medium text-sm">Score</th>
              <th class="text-left py-2 px-3 font-medium text-sm">Tag Overlap</th>
              <th class="text-left py-2 px-3 font-medium text-sm">Co-Community</th>
              <th class="text-left py-2 px-3 font-medium text-sm">Reaction Agreement</th>
              <th class="text-left py-2 px-3 font-medium text-sm">Trust Distance</th>
            </tr>
          </thead>
          <tbody>
            {#each data.affinities as entry (entry.user_b_id)}
              {@const breakdown = getBreakdown(entry.breakdown)}
              <tr class="border-b border-border hover:bg-accent/50 transition-colors">
                <td class="py-2 px-3">{entry.user_b_id}</td>
                <td class="py-2 px-3 font-mono">{(entry.affinity_score * 100).toFixed(1)}%</td>
                <td class="py-2 px-3 font-mono">{(breakdown.tag_overlap * 100).toFixed(1)}%</td>
                <td class="py-2 px-3 font-mono">{(breakdown.co_community * 100).toFixed(1)}%</td>
                <td class="py-2 px-3 font-mono">{(breakdown.reaction_agreement * 100).toFixed(1)}%</td>
                <td class="py-2 px-3 font-mono">{(breakdown.trust_distance * 100).toFixed(1)}%</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {/if}
  {:else}
    <div class="mt-4">
      <div class="flex items-end gap-3 mb-4">
        <TextInput
          label="Min Score"
          bind:value={minScore}
          placeholder="0.5"
          type="number"
          min="0"
          max="1"
          step="0.1"
        />
        <Button onclick={loadSimilarUsers} disabled={loadingSimilar} size="md">
          {loadingSimilar ? 'Loading...' : 'Filter'}
        </Button>
      </div>

      {#if similarUsers.length === 0}
        <p class="text-muted-foreground">No similar users found.</p>
      {:else}
        <div class="flex flex-col gap-2">
          {#each similarUsers as uid (uid)}
            <div class="flex items-center gap-2 p-2 rounded border border-border">
              <span class="w-5 h-5 flex items-center justify-center text-muted-foreground">
                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" class="w-4 h-4">
                  <path d="M4.5 6.375a4.125 4.125 0 118.25 0 4.125 4.125 0 01-8.25 0zM14.25 8.625a3.375 3.375 0 116.75 0 3.375 3.375 0 01-6.75 0zM1.5 19.125a7.125 7.125 0 0114.25 0v.003l-.001.119a.75.75 0 01-.363.63 13.067 13.067 0 01-6.761 1.873c-2.472 0-4.786-.684-6.76-1.873a.75.75 0 01-.364-.63l-.001-.122zM17.25 19.128l-.001.144a2.25 2.25 0 01-.233.96 10.088 10.088 0 005.06-1.01.75.75 0 00.42-.643 4.875 4.875 0 00-6.957-4.611 8.586 8.586 0 011.71 5.157v.003z"/>
                </svg>
              </span>
              <span>User #{uid}</span>
            </div>
          {/each}
        </div>
      {/if}
    </div>
  {/if}
</div>
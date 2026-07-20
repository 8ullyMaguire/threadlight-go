<script>
  import { onMount } from 'svelte';
  import api from '../api/client';

  export let goTo;

  let leaderboard = [];
  let loading = true;

  onMount(async () => {
    try {
      const data = await api.getLeaderboard('users', 'weekly');
      leaderboard = data.users || data || [];
    } catch (e) {
      // ignore
    } finally {
      loading = false;
    }
  });
</script>

<aside class="sidebar">
  <h3>Trending</h3>
  {#if loading}
    <p class="placeholder">Loading...</p>
  {:else if leaderboard.length > 0}
    <ol class="trending-list">
      {#each leaderboard.slice(0, 5) as entry, i}
        <li>
          <span class="rank">{i + 1}</span>
          <span class="name">{entry.username || entry.name}</span>
          <span class="score">{entry.score || entry.reputation || 0}</span>
        </li>
      {/each}
    </ol>
    <button class="see-all" on:click={() => goTo('leaderboard')}>See all</button>
  {:else}
    <p class="placeholder">No trending data yet</p>
  {/if}
  <div class="sidebar-footer">
    <a href="/about" target="_blank" rel="noopener">About</a>
    <a href="/privacy" target="_blank" rel="noopener">Privacy</a>
  </div>
</aside>

<style>
  .sidebar {
    padding: 1rem;
    background: var(--code-bg);
    border-radius: 8px;
  }
  h3 {
    margin: 0 0 0.75rem;
    font-size: 1rem;
    color: var(--text-h);
  }
  .trending-list {
    list-style: none;
    margin: 0;
    padding: 0;
  }
  .trending-list li {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.375rem 0;
    border-bottom: 1px solid var(--border);
    font-size: 0.875rem;
  }
  .trending-list li:last-child {
    border-bottom: none;
  }
  .rank {
    font-weight: 600;
    color: var(--accent);
    width: 1.5rem;
  }
  .name {
    flex: 1;
    color: var(--text-h);
  }
  .score {
    color: var(--text);
    font-size: 0.8rem;
  }
  .see-all {
    margin-top: 0.5rem;
    padding: 0.25rem 0.5rem;
    border: none;
    background: none;
    color: var(--accent);
    cursor: pointer;
    font-size: 0.8rem;
  }
  .see-all:hover {
    text-decoration: underline;
  }
  .placeholder {
    font-size: 0.875rem;
    color: var(--text);
  }
  .sidebar-footer {
    margin-top: 1rem;
    display: flex;
    gap: 1rem;
  }
  .sidebar-footer a {
    color: var(--text);
    font-size: 0.8rem;
    text-decoration: none;
  }
  .sidebar-footer a:hover {
    color: var(--accent);
  }
</style>

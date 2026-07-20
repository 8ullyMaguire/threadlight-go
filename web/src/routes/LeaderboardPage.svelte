<script>
  import { onMount } from 'svelte';
  import api from '../lib/api/client';

  export let goTo;

  let period = 'daily';
  let type = 'posting';
  let entries = [];
  let loading = false;

  const periods = [
    { id: 'daily', label: 'Daily' },
    { id: 'weekly', label: 'Weekly' },
    { id: 'monthly', label: 'Monthly' },
    { id: 'yearly', label: 'Yearly' },
    { id: 'alltime', label: 'All-Time' },
  ];

  const types = [
    { id: 'posting', label: 'Posting' },
    { id: 'tagging', label: 'Tagging' },
    { id: 'gifts', label: 'Gifts' },
    { id: 'trust_network', label: 'Trust Network' },
    { id: 'achievement_points', label: 'Achievement Points' },
  ];

  const sampleEntries = {
    posting: [
      { rank: 1, username: 'literary_lion', display_name: 'Literary Lion', score: 15420, change: 'up' },
      { rank: 2, username: 'quantum_quokka', display_name: 'Quantum Quokka', score: 13200, change: 'up' },
      { rank: 3, username: 'pixel_pioneer', display_name: 'Pixel Pioneer', score: 11980, change: 'down' },
      { rank: 4, username: 'neon_nomad', display_name: 'Neon Nomad', score: 10100, change: 'same' },
      { rank: 5, username: 'velvet_voyager', display_name: 'Velvet Voyager', score: 9470, change: 'up' },
      { rank: 6, username: 'astro_amateur', display_name: 'Astro Amateur', score: 8820, change: 'down' },
      { rank: 7, username: 'cyber_sage', display_name: 'Cyber Sage', score: 8150, change: 'up' },
      { rank: 8, username: 'echo_explorer', display_name: 'Echo Explorer', score: 7630, change: 'same' },
      { rank: 9, username: 'frost_fox', display_name: 'Frost Fox', score: 7010, change: 'down' },
      { rank: 10, username: 'blaze_bard', display_name: 'Blaze Bard', score: 6540, change: 'up' },
    ],
    tagging: [
      { rank: 1, username: 'tag_master_99', display_name: 'Tag Master', score: 28900, change: 'up' },
      { rank: 2, username: 'label_legend', display_name: 'Label Legend', score: 24700, change: 'same' },
      { rank: 3, username: 'categorizer_king', display_name: 'Categorizer King', score: 21350, change: 'up' },
      { rank: 4, username: 'hash_hunter', display_name: 'Hash Hunter', score: 18900, change: 'down' },
      { rank: 5, username: 'meta_maven', display_name: 'Meta Maven', score: 16300, change: 'up' },
      { rank: 6, username: 'tag_rabbit', display_name: 'Tag Rabbit', score: 14100, change: 'down' },
      { rank: 7, username: 'stamp_collector', display_name: 'Stamp Collector', score: 12450, change: 'same' },
      { rank: 8, username: 'keyword_knight', display_name: 'Keyword Knight', score: 10900, change: 'up' },
      { rank: 9, username: 'genre_genius', display_name: 'Genre Genius', score: 9800, change: 'down' },
      { rank: 10, username: 'taxonomy_titan', display_name: 'Taxonomy Titan', score: 8700, change: 'up' },
    ],
    gifts: [
      { rank: 1, username: 'santa_claus', display_name: 'Santa Claus', score: 50000, change: 'up' },
      { rank: 2, username: 'gift_giver_42', display_name: 'Gift Giver', score: 42500, change: 'up' },
      { rank: 3, username: 'philanthropy_fan', display_name: 'Philanthropy Fan', score: 38100, change: 'down' },
      { rank: 4, username: 'donation_dynamo', display_name: 'Donation Dynamo', score: 34200, change: 'up' },
      { rank: 5, username: 'present_pal', display_name: 'Present Pal', score: 30150, change: 'same' },
      { rank: 6, username: 'tipping_titan', display_name: 'Tipping Titan', score: 26900, change: 'down' },
      { rank: 7, username: 'generosity_guru', display_name: 'Generosity Guru', score: 23800, change: 'up' },
      { rank: 8, username: 'reward_royalty', display_name: 'Reward Royalty', score: 20200, change: 'up' },
      { rank: 9, username: 'bonus_bringer', display_name: 'Bonus Bringer', score: 17100, change: 'down' },
      { rank: 10, username: 'credit_czar', display_name: 'Credit Czar', score: 14800, change: 'same' },
    ],
    trust_network: [
      { rank: 1, username: 'trusted_elder', display_name: 'Trusted Elder', score: 9800, change: 'up' },
      { rank: 2, username: 'reliable_root', display_name: 'Reliable Root', score: 9200, change: 'same' },
      { rank: 3, username: 'verified_veteran', display_name: 'Verified Veteran', score: 8700, change: 'up' },
      { rank: 4, username: 'honest_halo', display_name: 'Honest Halo', score: 8300, change: 'down' },
      { rank: 5, username: 'integrity_icon', display_name: 'Integrity Icon', score: 7800, change: 'up' },
      { rank: 6, username: 'authentic_ace', display_name: 'Authentic Ace', score: 7400, change: 'same' },
      { rank: 7, username: 'genuine_genie', display_name: 'Genuine Genie', score: 7000, change: 'up' },
      { rank: 8, username: 'sincere_sage', display_name: 'Sincere Sage', score: 6600, change: 'down' },
      { rank: 9, username: 'truth_teller', display_name: 'Truth Teller', score: 6200, change: 'up' },
      { rank: 10, username: 'earnest_owl', display_name: 'Earnest Owl', score: 5800, change: 'same' },
    ],
    achievement_points: [
      { rank: 1, username: 'platinum_puma', display_name: 'Platinum Puma', score: 27500, change: 'up' },
      { rank: 2, username: 'golden_gazelle', display_name: 'Golden Gazelle', score: 24300, change: 'down' },
      { rank: 3, username: 'diamond_deer', display_name: 'Diamond Deer', score: 21800, change: 'up' },
      { rank: 4, username: 'silver_shark', display_name: 'Silver Shark', score: 19500, change: 'up' },
      { rank: 5, username: 'bronze_bear', display_name: 'Bronze Bear', score: 17200, change: 'down' },
      { rank: 6, username: 'crystal_coyote', display_name: 'Crystal Coyote', score: 15100, change: 'same' },
      { rank: 7, username: 'ruby_rabbit', display_name: 'Ruby Rabbit', score: 13400, change: 'up' },
      { rank: 8, username: 'emerald_eagle', display_name: 'Emerald Eagle', score: 11800, change: 'down' },
      { rank: 9, username: 'sapphire_serpent', display_name: 'Sapphire Serpent', score: 10200, change: 'up' },
      { rank: 10, username: 'amethyst_ant', display_name: 'Amethyst Ant', score: 8900, change: 'same' },
    ],
  };

  onMount(() => load());

  async function load() {
    loading = true;
    try {
      const data = await api.getLeaderboard(period, type);
      entries = data.length ? data : (sampleEntries[type] || []);
    } catch (_) {
      entries = sampleEntries[type] || [];
    } finally {
      loading = false;
    }
  }

  $: load();

  function changeIcon(ch) {
    if (ch === 'up') return '▲';
    if (ch === 'down') return '▼';
    return '—';
  }

  function changeClass(ch) {
    if (ch === 'up') return 'up';
    if (ch === 'down') return 'down';
    return 'same';
  }

  function rankEmoji(rank) {
    if (rank === 1) return '🥇';
    if (rank === 2) return '🥈';
    if (rank === 3) return '🥉';
    return '';
  }
</script>

<div class="leaderboard-page">
  <div class="page-header">
    <h1>Leaderboard</h1>
    <button class="btn-icon" on:click={() => goTo && goTo('home')}>&times;</button>
  </div>

  <div class="selectors">
    <div class="selector-group">
      {#each periods as p}
        <button class="selector-btn" class:active={period === p.id} on:click={() => period = p.id}>
          {p.label}
        </button>
      {/each}
    </div>
    <div class="selector-group">
      {#each types as t}
        <button class="selector-btn selector-btn-type" class:active={type === t.id} on:click={() => type = t.id}>
          {t.label}
        </button>
      {/each}
    </div>
  </div>

  {#if loading}
    <div class="loading">Loading...</div>
  {:else}
    <div class="leaderboard-list">
      {#each entries as entry, i (entry.rank)}
        <div class="entry" class:top-three={entry.rank <= 3}>
          <div class="rank-section">
            {#if entry.rank <= 3}
              <span class="rank-emoji">{rankEmoji(entry.rank)}</span>
            {:else}
              <span class="rank-num">#{entry.rank}</span>
            {/if}
          </div>
          <div class="avatar">
            {entry.display_name.charAt(0).toUpperCase()}
          </div>
          <div class="info">
            <span class="display-name">{entry.display_name}</span>
            <span class="username">{entry.username}</span>
          </div>
          <div class="score-section">
            <span class="score">{entry.score.toLocaleString()}</span>
            <span class="change {changeClass(entry.change)}">{changeIcon(entry.change)}</span>
          </div>
        </div>
      {/each}
    </div>
  {/if}
</div>

<style>
  .leaderboard-page {
    max-width: 720px;
    margin: 0 auto;
    padding: 1.5rem;
    font-family: system-ui, -apple-system, sans-serif;
    color: var(--text);
  }

  .page-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 1.25rem;
  }

  .page-header h1 {
    font-size: 1.5rem;
    font-weight: 700;
    margin: 0;
  }

  .btn-icon {
    background: none;
    border: none;
    font-size: 1.5rem;
    color: var(--text);
    cursor: pointer;
    padding: 0.25rem;
    line-height: 1;
  }

  .selectors {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    margin-bottom: 1.25rem;
  }

  .selector-group {
    display: flex;
    flex-wrap: wrap;
    gap: 0.25rem;
  }

  .selector-btn {
    padding: 0.375rem 0.75rem;
    background: var(--code-bg);
    border: 1px solid var(--border);
    border-radius: 999px;
    font-size: 0.8125rem;
    font-weight: 500;
    color: var(--text);
    cursor: pointer;
    transition: background 0.15s, border-color 0.15s;
  }

  .selector-btn:hover {
    background: var(--accent-bg);
  }

  .selector-btn.active {
    background: var(--accent);
    color: white;
    border-color: var(--accent);
  }

  .selector-btn-type {
    font-size: 0.75rem;
  }

  .loading {
    text-align: center;
    color: var(--text);
    padding: 2rem 0;
  }

  .leaderboard-list {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .entry {
    display: flex;
    align-items: center;
    gap: 0.875rem;
    padding: 0.75rem 1rem;
    border: 1px solid var(--border);
    border-radius: 8px;
    background: var(--code-bg);
  }

  .top-three {
    border-width: 2px;
  }

  .entry:nth-child(1).top-three {
    border-color: #f59e0b;
    background: #fffbeb;
  }

  .entry:nth-child(2).top-three {
    border-color: #d1d5db;
    background: #f9fafb;
  }

  .entry:nth-child(3).top-three {
    border-color: #d97706;
    background: #fff7ed;
  }

  .rank-section {
    width: 2.5rem;
    text-align: center;
    flex-shrink: 0;
  }

  .rank-emoji {
    font-size: 1.35rem;
  }

  .rank-num {
    font-size: 0.875rem;
    font-weight: 600;
    color: var(--text);
  }

  .avatar {
    width: 2.5rem;
    height: 2.5rem;
    border-radius: 50%;
    background: var(--accent);
    color: white;
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: 700;
    font-size: 1rem;
    flex-shrink: 0;
  }

  .info {
    flex: 1;
    display: flex;
    flex-direction: column;
    min-width: 0;
  }

  .display-name {
    font-weight: 600;
    font-size: 0.9375rem;
  }

  .username {
    font-size: 0.8125rem;
    color: var(--text);
  }

  .score-section {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    flex-shrink: 0;
  }

  .score {
    font-weight: 700;
    font-size: 1rem;
  }

  .change {
    font-size: 0.8125rem;
    font-weight: 600;
  }

  .change.up { color: #10b981; }
  .change.down { color: #ef4444; }
  .change.same { color: var(--text); }
</style>

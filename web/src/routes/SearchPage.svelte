<script>
  import { onMount } from 'svelte';
  import api from '../lib/api/client';

  export let goTo;

  let query = '';
  let results = [];
  let loading = false;
  let tab = 'posts';

  const tabs = [
    { id: 'posts', label: 'Posts' },
    { id: 'users', label: 'Users' },
    { id: 'tags', label: 'Tags' },
  ];

  async function search() {
    if (!query.trim()) return;
    loading = true;
    try {
      const res = await api.search({ q: query, type: tab });
      results = res.results || res.posts || res.users || res.tags || [];
    } catch (e) {
      results = [];
    } finally {
      loading = false;
    }
  }

  function handleKeydown(e) {
    if (e.key === 'Enter') search();
  }
</script>

<div class="search-page">
  <div class="search-bar">
    <input
      type="text"
      bind:value={query}
      placeholder="Search posts, users, tags..."
      on:keydown={handleKeydown}
    />
    <button type="submit" on:click={search} disabled={loading || !query.trim()}>
      {loading ? 'Searching...' : 'Search'}
    </button>
  </div>

  <div class="tabs">
    {#each tabs as t}
      <button
        class="tab"
        class:active={tab === t.id}
        on:click={() => { tab = t.id; if (query.trim()) search(); }}
      >{t.label}</button>
    {/each}
  </div>

  {#if loading}
    <p class="placeholder">Searching...</p>
  {:else if results.length > 0}
    <div class="results">
      {#if tab === 'users'}
        {#each results as user}
          <div class="result-user" on:click={() => goTo('profile', { username: user.username })}>
            <div class="user-avatar">{user.username?.charAt(0).toUpperCase()}</div>
            <div>
              <strong>{user.display_name || user.username}</strong>
              <span class="user-username">@{user.username}</span>
            </div>
          </div>
        {/each}
      {:else if tab === 'tags'}
        {#each results as tag}
          <div class="result-tag" on:click={() => goTo('feed', { tag: tag.name || tag }) }>
            <span class="tag-icon">#</span>
            <span class="tag-name">{tag.name || tag}</span>
            <span class="tag-count">{tag.post_count || tag.count || 0} posts</span>
          </div>
        {/each}
      {:else}
        {#each results as post}
          <div class="result-post">
            <span class="post-author">{post.author_username || 'unknown'}</span>
            <p>{post.body}</p>
            <span class="post-time">{new Date(post.created_at).toLocaleDateString()}</span>
          </div>
        {/each}
      {/if}
    </div>
  {:else if query.trim() && !loading}
    <p class="placeholder">No results found.</p>
  {:else}
    <p class="placeholder">Type something and press Enter to search.</p>
  {/if}
</div>

<style>
  .search-page {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }
  .search-bar {
    display: flex;
    gap: 0.5rem;
  }
  .search-bar input {
    flex: 1;
    padding: 0.5rem 0.75rem;
    border: 1px solid var(--border);
    border-radius: 6px;
    font-size: 1rem;
    background: var(--bg);
    color: var(--text-h);
  }
  .search-bar input:focus {
    outline: 2px solid var(--accent);
    outline-offset: -1px;
  }
  .search-bar button {
    padding: 0.5rem 1rem;
    background: var(--accent);
    color: white;
    border: none;
    border-radius: 6px;
    cursor: pointer;
    font-size: 0.875rem;
  }
  .search-bar button:disabled {
    opacity: 0.5;
  }
  .tabs {
    display: flex;
    gap: 0.5rem;
  }
  .tab {
    padding: 0.375rem 0.75rem;
    border: 1px solid var(--border);
    border-radius: 6px;
    background: var(--bg);
    color: var(--text);
    cursor: pointer;
    font-size: 0.8rem;
  }
  .tab.active {
    background: var(--accent);
    color: white;
    border-color: var(--accent);
  }
  .results {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }
  .result-user {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 0.75rem;
    background: var(--code-bg);
    border-radius: 8px;
    cursor: pointer;
    transition: background 0.2s;
  }
  .result-user:hover {
    background: var(--accent-bg);
  }
  .user-avatar {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    background: var(--accent);
    color: white;
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: 600;
  }
  .user-username {
    display: block;
    font-size: 0.8rem;
    color: var(--text);
  }
  .result-tag {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.75rem;
    background: var(--code-bg);
    border-radius: 8px;
    cursor: pointer;
    transition: background 0.2s;
  }
  .result-tag:hover {
    background: var(--accent-bg);
  }
  .tag-icon {
    color: var(--accent);
    font-weight: 700;
    font-size: 1.25rem;
  }
  .tag-name {
    font-weight: 500;
  }
  .tag-count {
    margin-left: auto;
    font-size: 0.8rem;
    color: var(--text);
  }
  .result-post {
    padding: 0.75rem;
    background: var(--code-bg);
    border-radius: 8px;
  }
  .result-post .post-author {
    font-weight: 600;
    font-size: 0.875rem;
    color: var(--accent);
  }
  .result-post p {
    margin: 0.25rem 0;
    color: var(--text-h);
  }
  .result-post .post-time {
    font-size: 0.75rem;
    color: var(--text);
  }
  .placeholder {
    text-align: center;
    padding: 3rem;
    color: var(--text);
  }
</style>

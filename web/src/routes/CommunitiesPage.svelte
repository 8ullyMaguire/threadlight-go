<script>
  import { onMount } from 'svelte';
  import { createEventDispatcher } from 'svelte';
  import api from '../lib/api/client';
  import { currentUser } from '../lib/stores/auth';

  export let goTo;

  const dispatch = createEventDispatcher();

  let communities = [];
  let myCommunities = [];
  let loading = false;
  let tab = 'browse';
  let showCreate = false;
  let newName = '';
  let newDescription = '';
  let newTags = '';
  let creating = false;

  onMount(async () => {
    loading = true;
    try {
      const all = await api.listCommunities();
      communities = all.communities || all || [];
      const mine = await api.listMyCommunities();
      myCommunities = mine.communities || mine || [];
    } catch (e) {
      // ignore
    } finally {
      loading = false;
    }
  });

  async function handleCreate() {
    if (!newName.trim()) return;
    creating = true;
    try {
      const tags = newTags.split(',').map(t => t.trim()).filter(Boolean);
      await api.createCommunity({
        name: newName.trim(),
        description: newDescription.trim(),
        tags,
      });
      showCreate = false;
      newName = '';
      newDescription = '';
      newTags = '';
      const all = await api.listCommunities();
      communities = all.communities || all || [];
    } catch (e) {
      // ignore
    } finally {
      creating = false;
    }
  }

  function handleJoin(communityId) {
    dispatch('navigate', { page: 'community', id: communityId });
  }
</script>

<div class="communities-page">
  <div class="header">
    <h2>Communities</h2>
    <button class="create-btn" on:click={() => showCreate = !showCreate}>
      {showCreate ? 'Cancel' : 'Create Community'}
    </button>
  </div>

  {#if showCreate}
    <div class="create-form">
      <input type="text" bind:value={newName} placeholder="Community name" />
      <textarea bind:value={newDescription} placeholder="Description" rows="2"></textarea>
      <input type="text" bind:value={newTags} placeholder="Tags (comma separated)" />
      <button on:click={handleCreate} disabled={creating || !newName.trim()}>
        {creating ? 'Creating...' : 'Create'}
      </button>
    </div>
  {/if}

  <div class="tabs">
    <button class="tab" class:active={tab === 'browse'} on:click={() => tab = 'browse'}>Browse</button>
    <button class="tab" class:active={tab === 'mine'} on:click={() => tab = 'mine'}>My Communities</button>
  </div>

  {#if loading}
    <p class="placeholder">Loading communities...</p>
  {:else}
    {#if tab === 'browse'}
      <div class="community-grid">
        {#each communities as c}
          <div class="community-card">
            <h3>{c.name}</h3>
            {#if c.description}
              <p>{c.description}</p>
            {/if}
            <div class="community-meta">
              <span>{c.member_count || 0} members</span>
              {#if c.tags && c.tags.length > 0}
                <div class="tags">
                  {#each c.tags as tag}
                    <span class="tag">{tag}</span>
                  {/each}
                </div>
              {/if}
            </div>
            <button class="join-btn" on:click={() => handleJoin(c.id)}>View</button>
          </div>
        {/each}
        {#if communities.length === 0}
          <p class="empty">No communities yet.</p>
        {/if}
      </div>
    {:else}
      <div class="community-grid">
        {#each myCommunities as c}
          <div class="community-card">
            <h3>{c.name}</h3>
            {#if c.description}
              <p>{c.description}</p>
            {/if}
            <div class="community-meta">
              <span>{c.member_count || 0} members</span>
              {#if c.role}
                <span class="role-badge">{c.role}</span>
              {/if}
            </div>
            <button class="join-btn" on:click={() => handleJoin(c.id)}>Manage</button>
          </div>
        {/each}
        {#if myCommunities.length === 0}
          <p class="empty">You haven't joined any communities yet.</p>
        {/if}
      </div>
    {/if}
  {/if}
</div>

<style>
  .communities-page {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }
  .header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  .header h2 {
    margin: 0;
  }
  .create-btn {
    padding: 0.5rem 1rem;
    border: 1px solid var(--accent);
    border-radius: 6px;
    background: none;
    color: var(--accent);
    cursor: pointer;
    font-size: 0.875rem;
  }
  .create-btn:hover {
    background: var(--accent);
    color: white;
  }
  .create-form {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    padding: 1rem;
    background: var(--code-bg);
    border-radius: 8px;
  }
  .create-form input, .create-form textarea {
    padding: 0.5rem;
    border: 1px solid var(--border);
    border-radius: 6px;
    background: var(--bg);
    color: var(--text-h);
    font-family: inherit;
  }
  .create-form button {
    padding: 0.5rem 1rem;
    background: var(--accent);
    color: white;
    border: none;
    border-radius: 6px;
    cursor: pointer;
    align-self: flex-end;
  }
  .create-form button:disabled {
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
  .community-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
    gap: 0.75rem;
  }
  .community-card {
    padding: 1rem;
    background: var(--code-bg);
    border-radius: 8px;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }
  .community-card h3 {
    margin: 0;
    font-size: 1rem;
  }
  .community-card p {
    margin: 0;
    font-size: 0.875rem;
    color: var(--text);
  }
  .community-meta {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
    font-size: 0.8rem;
    color: var(--text);
    align-items: center;
  }
  .tags {
    display: flex;
    gap: 0.25rem;
    flex-wrap: wrap;
  }
  .tag {
    padding: 0.125rem 0.375rem;
    background: var(--accent-bg);
    color: var(--accent);
    border-radius: 4px;
    font-size: 0.7rem;
  }
  .role-badge {
    padding: 0.125rem 0.375rem;
    background: #dbeafe;
    color: #1d4ed8;
    border-radius: 4px;
    font-size: 0.7rem;
    text-transform: capitalize;
  }
  .join-btn {
    padding: 0.375rem 0.75rem;
    border: 1px solid var(--accent);
    border-radius: 6px;
    background: none;
    color: var(--accent);
    cursor: pointer;
    font-size: 0.8rem;
    align-self: flex-start;
    margin-top: auto;
  }
  .join-btn:hover {
    background: var(--accent);
    color: white;
  }
  .placeholder, .empty {
    text-align: center;
    padding: 2rem;
    color: var(--text);
  }
</style>

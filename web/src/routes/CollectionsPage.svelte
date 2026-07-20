<script>
  import { onMount } from 'svelte';
  import api from '../lib/api/client';

  export let goTo;

  let collections = [];
  let loading = false;
  let showCreate = false;
  let newName = '';
  let newDescription = '';
  let creating = false;

  onMount(async () => {
    loading = true;
    try {
      const data = await api.listCollections();
      collections = data.collections || data || [];
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
      await api.createCollection({
        name: newName.trim(),
        description: newDescription.trim(),
      });
      showCreate = false;
      newName = '';
      newDescription = '';
      const data = await api.listCollections();
      collections = data.collections || data || [];
    } catch (e) {
      // ignore
    } finally {
      creating = false;
    }
  }

  function handleView(collection) {
    goTo('collection', { id: collection.id });
  }
</script>

<div class="collections-page">
  <div class="header">
    <h2>Collections</h2>
    <button class="create-btn" on:click={() => showCreate = !showCreate}>
      {showCreate ? 'Cancel' : 'Create Collection'}
    </button>
  </div>

  {#if showCreate}
    <div class="create-form">
      <input type="text" bind:value={newName} placeholder="Collection name" />
      <textarea bind:value={newDescription} placeholder="Description (optional)" rows="2"></textarea>
      <button on:click={handleCreate} disabled={creating || !newName.trim()}>
        {creating ? 'Creating...' : 'Create'}
      </button>
    </div>
  {/if}

  {#if loading}
    <p class="placeholder">Loading collections...</p>
  {:else if collections.length > 0}
    <div class="collection-grid">
      {#each collections as c}
        <div class="collection-card" on:click={() => handleView(c)}>
          <h3>{c.name}</h3>
          {#if c.description}
            <p>{c.description}</p>
          {/if}
          <div class="collection-meta">
            <span>{c.item_count || 0} items</span>
          </div>
        </div>
      {/each}
    </div>
  {:else}
    <p class="placeholder">No collections yet. Create one to save your favorite posts.</p>
  {/if}
</div>

<style>
  .collections-page {
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
  .collection-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
    gap: 0.75rem;
  }
  .collection-card {
    padding: 1rem;
    background: var(--code-bg);
    border-radius: 8px;
    cursor: pointer;
    transition: background 0.2s;
  }
  .collection-card:hover {
    background: var(--accent-bg);
  }
  .collection-card h3 {
    margin: 0;
    font-size: 1rem;
  }
  .collection-card p {
    margin: 0.25rem 0;
    font-size: 0.875rem;
    color: var(--text);
  }
  .collection-meta {
    font-size: 0.8rem;
    color: var(--text);
    margin-top: 0.5rem;
  }
  .placeholder {
    text-align: center;
    padding: 3rem;
    color: var(--text);
  }
</style>

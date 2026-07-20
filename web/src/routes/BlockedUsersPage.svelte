<script>
  import { onMount } from 'svelte';
  import api from '../lib/api/client';

  export let goTo;

  let blocked = [];
  let loading = true;

  onMount(load);

  async function load() {
    loading = true;
    try {
      const data = await api.getBlockedUsers();
      blocked = data.users || data || [];
    } catch (e) {
      blocked = [];
    } finally {
      loading = false;
    }
  }

  async function unblock(username) {
    try {
      await api.unblockUser(username);
      blocked = blocked.filter(u => u.username !== username);
    } catch (e) {
      console.error(e);
    }
  }
</script>

<div class="page">
  <h2>Blocked Users</h2>

  {#if loading}
    <p class="placeholder">Loading...</p>
  {:else if blocked.length === 0}
    <p class="placeholder">You haven't blocked anyone yet.</p>
  {:else}
    <div class="blocked-list">
      {#each blocked as user}
        <div class="blocked-card">
          <div class="user-info">
            <span class="avatar">{user.username[0].toUpperCase()}</span>
            <div>
              <strong>{user.display_name || user.username}</strong>
              <span class="username">@{user.username}</span>
            </div>
          </div>
          <button class="btn btn-sm" on:click={() => unblock(user.username)}>Unblock</button>
        </div>
      {/each}
    </div>
  {/if}
</div>

<style>
  .page {
    max-width: 500px;
    margin: 0 auto;
    padding: 1rem;
  }

  h2 {
    margin-bottom: 1rem;
  }

  .placeholder {
    text-align: center;
    color: var(--text);
    opacity: 0.6;
    padding: 2rem;
  }

  .blocked-list {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .blocked-card {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0.75rem;
    border: 1px solid var(--border);
    border-radius: 6px;
  }

  .user-info {
    display: flex;
    align-items: center;
    gap: 0.75rem;
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
    font-weight: 600;
    flex-shrink: 0;
  }

  .username {
    display: block;
    font-size: 0.8125rem;
    opacity: 0.6;
  }

  .btn {
    padding: 0.5rem 1rem;
    border: 1px solid var(--border);
    border-radius: 6px;
    background: var(--bg);
    color: var(--text);
    cursor: pointer;
    font-size: 0.875rem;
  }

  .btn-sm {
    padding: 0.3rem 0.6rem;
    font-size: 0.8125rem;
  }
</style>

<script>
  import { onMount } from 'svelte';
  import api from '../lib/api/client';

  let loading = false;
  let tab = 'dashboard';
  let stats = {};
  let users = [];
  let flags = [];
  let systemMessage = '';
  let sendingMessage = false;

  const tabs = [
    { id: 'dashboard', label: 'Dashboard' },
    { id: 'users', label: 'Users' },
    { id: 'flags', label: 'Flags' },
    { id: 'announce', label: 'Announce' },
  ];

  onMount(async () => {
    await loadDashboard();
  });

  async function loadDashboard() {
    loading = true;
    try {
      const data = await api.getAdminStats();
      stats = data.stats || data || {};
    } catch (e) {
      // ignore
    } finally {
      loading = false;
    }
  }

  async function loadUsers() {
    loading = true;
    try {
      const data = await api.getAdminUsers();
      users = data.users || data || [];
    } catch (e) {
      users = [];
    } finally {
      loading = false;
    }
  }

  async function loadFlags() {
    loading = true;
    try {
      const data = await api.getAdminFlags();
      flags = data.flags || data || [];
    } catch (e) {
      flags = [];
    } finally {
      loading = false;
    }
  }

  async function handleUserAction(userId, action) {
    try {
      await api.adminUserAction(userId, action);
      await loadUsers();
    } catch (e) {
      // ignore
    }
  }

  async function handleFlagAction(flagId, action) {
    try {
      await api.adminFlagAction(flagId, action);
      await loadFlags();
    } catch (e) {
      // ignore
    }
  }

  async function handleAnnounce() {
    if (!systemMessage.trim()) return;
    sendingMessage = true;
    try {
      await api.adminAnnounce(systemMessage.trim());
      systemMessage = '';
    } catch (e) {
      // ignore
    } finally {
      sendingMessage = false;
    }
  }

  async function switchTab(t) {
    tab = t;
    if (t === 'users') await loadUsers();
    else if (t === 'flags') await loadFlags();
    else if (t === 'dashboard') await loadDashboard();
  }
</script>

<div class="admin-page">
  <h2>Admin Panel</h2>

  <div class="tabs">
    {#each tabs as t}
      <button
        class="tab"
        class:active={tab === t.id}
        on:click={() => switchTab(t.id)}
      >{t.label}</button>
    {/each}
  </div>

  {#if tab === 'dashboard'}
    <div class="stats-grid">
      <div class="stat-card">
        <span class="stat-value">{stats.user_count || 0}</span>
        <span class="stat-label">Users</span>
      </div>
      <div class="stat-card">
        <span class="stat-value">{stats.post_count || 0}</span>
        <span class="stat-label">Posts</span>
      </div>
      <div class="stat-card">
        <span class="stat-value">{stats.community_count || 0}</span>
        <span class="stat-label">Communities</span>
      </div>
      <div class="stat-card">
        <span class="stat-value">{stats.flag_count || 0}</span>
        <span class="stat-label">Open Flags</span>
      </div>
    </div>
  {:else if tab === 'users'}
    {#if loading}
      <p class="placeholder">Loading users...</p>
    {:else}
      <div class="user-list">
        {#each users as user}
          <div class="user-row">
            <span class="user-name">{user.username}</span>
            <span class="user-role">{user.role || 'user'}</span>
            <div class="user-actions">
              {#if user.role !== 'admin'}
                <button class="action-btn ban" on:click={() => handleUserAction(user.id, 'ban')}>Ban</button>
                <button class="action-btn warn" on:click={() => handleUserAction(user.id, 'warn')}>Warn</button>
              {/if}
            </div>
          </div>
        {/each}
        {#if users.length === 0}
          <p class="empty">Only you for now.</p>
        {/if}
      </div>
    {/if}
  {:else if tab === 'flags'}
    {#if loading}
      <p class="placeholder">Loading flags...</p>
    {:else}
      <div class="flag-list">
        {#each flags as flag}
          <div class="flag-row">
            <span class="flag-post">Post #{flag.post_id}</span>
            <span class="flag-reason">{flag.reason}</span>
            <span class="flag-user">by {flag.reporter_username}</span>
            <div class="flag-actions">
              <button class="action-btn approve" on:click={() => handleFlagAction(flag.id, 'approve')}>Dismiss</button>
              <button class="action-btn remove" on:click={() => handleFlagAction(flag.id, 'remove')}>Remove Post</button>
            </div>
          </div>
        {/each}
        {#if flags.length === 0}
          <p class="empty">No flagged posts.</p>
        {/if}
      </div>
    {/if}
  {:else if tab === 'announce'}
    <div class="announce-section">
      <p>Send a system-wide announcement to all users.</p>
      <textarea
        bind:value={systemMessage}
        placeholder="Type your announcement..."
        rows="4"
      ></textarea>
      <button on:click={handleAnnounce} disabled={sendingMessage || !systemMessage.trim()}>
        {sendingMessage ? 'Sending...' : 'Send Announcement'}
      </button>
    </div>
  {/if}
</div>

<style>
  .admin-page {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }
  .admin-page h2 {
    margin: 0;
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
  .stats-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
    gap: 0.75rem;
  }
  .stat-card {
    padding: 1rem;
    background: var(--code-bg);
    border-radius: 8px;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.25rem;
  }
  .stat-value {
    font-size: 1.75rem;
    font-weight: 700;
    color: var(--accent);
  }
  .stat-label {
    font-size: 0.8rem;
    color: var(--text);
  }
  .user-list, .flag-list {
    display: flex;
    flex-direction: column;
    gap: 0.375rem;
  }
  .user-row, .flag-row {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 0.625rem 0.75rem;
    background: var(--code-bg);
    border-radius: 6px;
    font-size: 0.875rem;
  }
  .user-name, .flag-post {
    font-weight: 500;
    min-width: 120px;
  }
  .user-role {
    padding: 0.125rem 0.375rem;
    background: var(--accent-bg);
    color: var(--accent);
    border-radius: 4px;
    font-size: 0.75rem;
    text-transform: capitalize;
  }
  .flag-reason {
    flex: 1;
    color: var(--text);
  }
  .flag-user {
    color: var(--text);
    font-size: 0.8rem;
  }
  .user-actions, .flag-actions {
    display: flex;
    gap: 0.375rem;
    margin-left: auto;
  }
  .action-btn {
    padding: 0.25rem 0.5rem;
    border: none;
    border-radius: 4px;
    cursor: pointer;
    font-size: 0.75rem;
  }
  .action-btn.ban { background: #fee2e2; color: #dc2626; }
  .action-btn.warn { background: #fef3c7; color: #d97706; }
  .action-btn.approve { background: #d1fae5; color: #059669; }
  .action-btn.remove { background: #fee2e2; color: #dc2626; }
  .announce-section {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    max-width: 500px;
  }
  .announce-section p {
    margin: 0;
    color: var(--text);
    font-size: 0.875rem;
  }
  .announce-section textarea {
    width: 100%;
    padding: 0.5rem;
    border: 1px solid var(--border);
    border-radius: 6px;
    background: var(--bg);
    color: var(--text-h);
    font-family: inherit;
    resize: vertical;
  }
  .announce-section button {
    padding: 0.5rem 1rem;
    background: var(--accent);
    color: white;
    border: none;
    border-radius: 6px;
    cursor: pointer;
    align-self: flex-start;
  }
  .announce-section button:disabled {
    opacity: 0.5;
  }
  .placeholder, .empty {
    text-align: center;
    padding: 2rem;
    color: var(--text);
  }
</style>

<script>
  import { onMount } from 'svelte';
  import api from '../lib/api/client';
  import { currentUser, logout } from '../lib/stores/auth';

  export let goTo;

  let activeTab = 'profile';
  let displayName = '';
  let bio = '';
  let currentPassword = '';
  let newPassword = '';
  let confirmPassword = '';
  let saving = false;
  let message = '';
  let messageType = '';
  let filters = [];
  let newFilterWord = '';
  let newFilterType = 'hide';
  let balance = 0;
  let claiming = false;
  let claimMessage = '';
  let privacyShowFollowers = true;
  let privacyShowOnline = true;
  let theme = 'light';
  let hideReadPosts = false;
  let deleting = false;
  $: emailVerified = $currentUser?.email_verified || false;
  let verifying = false;
  let verifyMessage = '';

  const tabs = [
    { id: 'profile', label: 'Profile' },
    { id: 'filters', label: 'Content Filters' },
    { id: 'credits', label: 'Credits' },
    { id: 'account', label: 'Account' },
  ];

  async function load() {
    const u = $currentUser;
    if (u) {
      displayName = u.display_name || u.username || '';
      bio = u.bio || '';
    }
    try {
      filters = await api.listFilters();
    } catch (_) {
      filters = [];
    }
    try {
      const b = await api.getBalance();
      balance = b.balance || 0;
    } catch (_) {
      balance = 0;
    }
    try {
      const pref = await api.get('/api/v1/users/me/preferences');
      theme = pref.theme || 'light';
      hideReadPosts = pref.hide_read_posts || false;
    } catch (_) {
      // ignore
    }
  }

  async function saveProfile() {
    saving = true;
    message = '';
    try {
      await api.patch('/api/v1/users/me', { display_name: displayName, bio });
      message = 'Profile updated.';
      messageType = 'success';
    } catch (e) {
      message = e.message || 'Failed to update profile.';
      messageType = 'error';
    } finally {
      saving = false;
    }
  }

  async function changePassword() {
    if (newPassword !== confirmPassword) {
      message = 'Passwords do not match.';
      messageType = 'error';
      return;
    }
    if (newPassword.length < 8) {
      message = 'Password must be at least 8 characters.';
      messageType = 'error';
      return;
    }
    saving = true;
    message = '';
    try {
      await api.put('/api/v1/auth/password', { current_password: currentPassword, new_password: newPassword });
      message = 'Password changed.';
      messageType = 'success';
      currentPassword = '';
      newPassword = '';
      confirmPassword = '';
    } catch (e) {
      message = e.message || 'Failed to change password.';
      messageType = 'error';
    } finally {
      saving = false;
    }
  }

  async function addFilter() {
    if (!newFilterWord.trim()) return;
    try {
      const f = await api.createFilter({ word: newFilterWord.trim(), type: newFilterType });
      filters = [...filters, f];
      newFilterWord = '';
    } catch (e) {
      message = e.message || 'Failed to add filter.';
      messageType = 'error';
    }
  }

  async function removeFilter(id) {
    try {
      await api.deleteFilter(id);
      filters = filters.filter(f => f.id !== id);
    } catch (e) {
      message = e.message || 'Failed to remove filter.';
      messageType = 'error';
    }
  }

  async function toggleTheme() {
    try {
      theme = theme === 'light' ? 'dark' : 'light';
      await api.patch('/api/v1/users/me/preferences', { theme });
      document.documentElement.setAttribute('data-theme', theme);
    } catch (e) {
      console.error(e);
    }
  }

  async function toggleHideRead() {
    try {
      hideReadPosts = !hideReadPosts;
      await api.patch('/api/v1/users/me/preferences', { hide_read_posts: hideReadPosts });
    } catch (e) {
      console.error(e);
    }
  }

  async function deleteAccount() {
    if (!confirm('Are you sure? This action cannot be undone.')) return;
    deleting = true;
    try {
      await api.del('/api/v1/users/me');
      logout();
      goTo && goTo('auth');
    } catch (e) {
      message = e.message || 'Failed to delete account';
      messageType = 'error';
    } finally {
      deleting = false;
    }
  }

  async function resendVerification() {
    verifying = true;
    verifyMessage = '';
    try {
      await api.resendVerification();
      verifyMessage = 'Verification email sent!';
    } catch (e) {
      verifyMessage = e.message || 'Failed to send';
    } finally {
      verifying = false;
    }
  }

  async function claimDaily() {
    claiming = true;
    claimMessage = '';
    try {
      const resp = await api.claimDaily();
      balance = resp.balance;
      claimMessage = resp.message || 'Daily credits claimed!';
    } catch (e) {
      claimMessage = e.message || 'Already claimed today.';
    } finally {
      claiming = false;
    }
  }
</script>

<div class="settings-page">
  <div class="page-header">
    <h1>Settings</h1>
    <button class="btn-icon" on:click={() => goTo && goTo('home')}>&times;</button>
  </div>

  <div class="tabs">
    {#each tabs as tab}
      <button
        class="tab"
        class:active={activeTab === tab.id}
        on:click={() => { activeTab = tab.id; message = ''; }}
      >
        {tab.label}
      </button>
    {/each}
  </div>

  {#if message}
    <div class="toast" class:toast-success={messageType === 'success'} class:toast-error={messageType === 'error'}>
      {message}
    </div>
  {/if}

  {#if activeTab === 'profile'}
    <div class="tab-content">
      <label>
        <span>Display Name</span>
        <input type="text" bind:value={displayName} placeholder="Your display name" />
      </label>
      <label>
        <span>Bio</span>
        <textarea bind:value={bio} placeholder="Tell us about yourself" rows="3"></textarea>
      </label>
      <button class="btn btn-primary" on:click={saveProfile} disabled={saving}>
        {saving ? 'Saving...' : 'Save Profile'}
      </button>

      <hr />

      <h3>Change Password</h3>
      <label>
        <span>Current Password</span>
        <input type="password" bind:value={currentPassword} />
      </label>
      <label>
        <span>New Password</span>
        <input type="password" bind:value={newPassword} />
      </label>
      <label>
        <span>Confirm New Password</span>
        <input type="password" bind:value={confirmPassword} />
      </label>
      <button class="btn btn-secondary" on:click={changePassword} disabled={saving}>
        {saving ? 'Changing...' : 'Change Password'}
      </button>
    </div>

  {:else if activeTab === 'filters'}
    <div class="tab-content">
      <p class="description">Words or phrases you want to hide or blur from your feed.</p>

      <div class="add-filter-row">
        <input type="text" bind:value={newFilterWord} placeholder="word or phrase" />
        <select bind:value={newFilterType}>
          <option value="hide">Hide</option>
          <option value="blur">Blur</option>
        </select>
        <button class="btn btn-primary btn-sm" on:click={addFilter}>Add</button>
      </div>

      {#if filters.length === 0}
        <p class="empty">No filters yet.</p>
      {:else}
        <div class="filter-list">
          {#each filters as f (f.id)}
            <div class="filter-item">
              <span class="filter-word">{f.word}</span>
              <span class="filter-type-badge">{f.type}</span>
              <button class="btn-icon-small" on:click={() => removeFilter(f.id)}>&times;</button>
            </div>
          {/each}
        </div>
      {/if}
    </div>

  {:else if activeTab === 'credits'}
    <div class="tab-content">
      <div class="balance-card">
        <span class="balance-label">Balance</span>
        <span class="balance-amount">{balance}</span>
      </div>

      <button class="btn btn-primary" on:click={claimDaily} disabled={claiming}>
        {claiming ? 'Claiming...' : 'Claim Daily Reward'}
      </button>
      {#if claimMessage}
        <p class="claim-message">{claimMessage}</p>
      {/if}

      <hr />

      <h3>Transaction History</h3>
      <p class="placeholder-text">Transaction history will appear here.</p>
    </div>

  {:else if activeTab === 'account'}
    <div class="tab-content">
      <h3>Appearance</h3>
      <label class="toggle-row">
        <span>Dark theme</span>
        <input type="checkbox" checked={theme === 'dark'} on:change={toggleTheme} />
      </label>

      <hr />

      <h3>Privacy</h3>
      <label class="toggle-row">
        <span>Show followers to others</span>
        <input type="checkbox" bind:checked={privacyShowFollowers} />
      </label>
      <label class="toggle-row">
        <span>Show online status</span>
        <input type="checkbox" bind:checked={privacyShowOnline} />
      </label>

      <hr />

      <h3>Feed</h3>
      <label class="toggle-row">
        <span>Hide read posts</span>
        <input type="checkbox" checked={hideReadPosts} on:change={toggleHideRead} />
      </label>

      <hr />

      <h3>Email</h3>
      {#if emailVerified}
        <p class="verified">Email verified</p>
      {:else}
        <p class="unverified">Email not verified</p>
        <button class="btn btn-primary" on:click={resendVerification} disabled={verifying}>
          {verifying ? 'Sending...' : 'Resend verification email'}
        </button>
        {#if verifyMessage}
          <p class="message">{verifyMessage}</p>
        {/if}
      {/if}

      <hr />

      <h3 class="danger-heading">Danger Zone</h3>
      <button class="btn btn-danger" on:click={() => { logout(); goTo && goTo('auth'); }}>Log Out</button>
      <button class="btn btn-danger" on:click={deleteAccount} disabled={deleting}>
        {deleting ? 'Deleting...' : 'Delete Account'}
      </button>
    </div>
  {/if}
</div>

<style>
  .settings-page {
    max-width: 640px;
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

  .tabs {
    display: flex;
    gap: 0.25rem;
    border-bottom: 2px solid var(--border);
    margin-bottom: 1rem;
  }

  .tab {
    padding: 0.625rem 1.125rem;
    background: none;
    border: none;
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--text);
    cursor: pointer;
    border-bottom: 2px solid transparent;
    margin-bottom: -2px;
    transition: color 0.15s, border-color 0.15s;
  }

  .tab.active {
    color: var(--accent);
    border-bottom-color: var(--accent);
  }

  .tab-content {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  label {
    display: flex;
    flex-direction: column;
    gap: 0.375rem;
    font-size: 0.875rem;
    font-weight: 500;
  }

  input[type="text"],
  input[type="password"],
  textarea,
  select {
    padding: 0.5rem;
    border: 1px solid var(--border);
    border-radius: 6px;
    font-size: 0.875rem;
    outline: none;
    background: var(--bg);
    color: var(--text);
  }

  input:focus,
  textarea:focus,
  select:focus {
    border-color: var(--accent);
  }

  textarea {
    resize: vertical;
  }

  hr {
    border: none;
    border-top: 1px solid var(--border);
    margin: 0.5rem 0;
  }

  h3 {
    font-size: 1.125rem;
    font-weight: 600;
    margin: 0.5rem 0 0;
  }

  .btn {
    padding: 0.5rem 1rem;
    border: none;
    border-radius: 6px;
    font-size: 0.875rem;
    font-weight: 500;
    cursor: pointer;
    transition: opacity 0.15s;
    align-self: flex-start;
  }

  .btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .btn-primary {
    background: var(--accent);
    color: white;
  }

  .btn-secondary {
    background: var(--code-bg);
    color: var(--text);
    border: 1px solid var(--border);
  }

  .btn-danger {
    background: #ef4444;
    color: white;
  }

  .btn-sm {
    padding: 0.375rem 0.75rem;
    font-size: 0.8125rem;
  }

  .toast {
    padding: 0.75rem 1rem;
    border-radius: 6px;
    font-size: 0.875rem;
    font-weight: 500;
    margin-bottom: 0.75rem;
  }

  .toast-success {
    background: #d1fae5;
    color: #065f46;
  }

  .toast-error {
    background: #fee2e2;
    color: #991b1b;
  }

  .description {
    font-size: 0.875rem;
    color: var(--text);
    margin: 0;
  }

  .add-filter-row {
    display: flex;
    gap: 0.5rem;
    align-items: center;
  }

  .add-filter-row input {
    flex: 1;
  }

  .filter-list {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .filter-item {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 0.5rem 0.75rem;
    border: 1px solid var(--border);
    border-radius: 6px;
    background: var(--code-bg);
  }

  .filter-word {
    flex: 1;
    font-size: 0.9375rem;
  }

  .filter-type-badge {
    font-size: 0.75rem;
    background: var(--code-bg);
    border: 1px solid var(--border);
    padding: 0.15rem 0.5rem;
    border-radius: 999px;
    text-transform: uppercase;
    font-weight: 600;
  }

  .btn-icon-small {
    background: none;
    border: none;
    font-size: 1.125rem;
    color: var(--text);
    cursor: pointer;
    line-height: 1;
    padding: 0.125rem;
  }

  .btn-icon-small:hover {
    color: #ef4444;
  }

  .empty {
    text-align: center;
    color: var(--text);
    padding: 1rem 0;
  }

  .balance-card {
    display: flex;
    align-items: baseline;
    gap: 0.75rem;
    padding: 1.25rem;
    border: 1px solid var(--border);
    border-radius: 8px;
    background: var(--code-bg);
  }

  .balance-label {
    font-size: 0.9375rem;
    color: var(--text);
  }

  .balance-amount {
    font-size: 2rem;
    font-weight: 700;
  }

  .claim-message {
    font-size: 0.875rem;
    color: var(--text);
    margin: 0;
  }

  .placeholder-text {
    color: var(--text);
    font-size: 0.875rem;
  }

  .toggle-row {
    flex-direction: row;
    align-items: center;
    justify-content: space-between;
  }

  .toggle-row input[type="checkbox"] {
    width: 1.25rem;
    height: 1.25rem;
    accent-color: var(--accent);
  }

  .danger-heading {
    color: #ef4444;
  }

  .verified {
    color: #16a34a;
    font-size: 0.875rem;
  }

  .unverified {
    color: #dc2626;
    font-size: 0.875rem;
  }

  .message {
    font-size: 0.875rem;
    color: var(--text);
    margin: 0.5rem 0 0 0;
  }

  .btn {
    padding: 0.5rem 1rem;
    border: 1px solid var(--border);
    border-radius: 6px;
    background: var(--bg);
    color: var(--text);
    cursor: pointer;
    font-size: 0.875rem;
    display: inline-flex;
    align-items: center;
    gap: 0.375rem;
  }

  .btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .btn-primary {
    background: var(--accent);
    color: white;
    border-color: var(--accent);
  }

  .btn-danger {
    background: #ef4444;
    color: white;
    border-color: #dc2626;
  }
</style>

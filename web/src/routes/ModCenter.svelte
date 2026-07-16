<script>
  import { onMount } from 'svelte';
  import api from '../lib/api/client';

  export let goTo;

  let activeTab = 'modlog';
  let searchQuery = '';
  let modLog = [];
  let loading = false;
  let error = '';
  let reports = [];
  let noteUserId = '';
  let noteBody = '';
  let modNotes = [];
  let resolving = {};
  let reportsLoaded = false;

  const tabs = [
    { id: 'modlog', label: 'Mod Log' },
    { id: 'reports', label: 'Reports' },
    { id: 'notes', label: 'Mod Notes' },
    { id: 'jury', label: 'Jury Duty' },
    { id: 'flagqueue', label: 'Flag Queue' },
  ];

  const sampleModLog = [
    { id: 1, action_type: 'warning', reason: 'Spam in community feed', target_user: '@spambot42', moderator: '@mod_alice', created_at: new Date(Date.now() - 2 * 3600000).toISOString() },
    { id: 2, action_type: 'timeout', reason: 'Harassment in DMs', target_user: '@angryuser', moderator: '@mod_bob', created_at: new Date(Date.now() - 5 * 3600000).toISOString() },
    { id: 3, action_type: 'suspension', reason: 'Posting NSFW content without label', target_user: '@edgelord', moderator: '@mod_alice', created_at: new Date(Date.now() - 86400000).toISOString() },
    { id: 4, action_type: 'ban', reason: 'Evading previous ban', target_user: '@alt_account99', moderator: '@admin_carol', created_at: new Date(Date.now() - 3 * 86400000).toISOString() },
    { id: 5, action_type: 'content_removed', reason: 'Copyrighted material', target_user: '@reposter', moderator: '@mod_bob', created_at: new Date(Date.now() - 7 * 86400000).toISOString() },
    { id: 6, action_type: 'warning', reason: 'Excessive self-promotion', target_user: '@promoter_88', moderator: '@mod_alice', created_at: new Date(Date.now() - 12 * 3600000).toISOString() },
    { id: 7, action_type: 'content_removed', reason: 'Misinformation about vaccines', target_user: '@truthseeker', moderator: '@admin_carol', created_at: new Date(Date.now() - 14 * 86400000).toISOString() },
    { id: 8, action_type: 'suspension', reason: 'Ban evasion via new account', target_user: '@backagain', moderator: '@mod_bob', created_at: new Date(Date.now() - 20 * 86400000).toISOString() },
  ];

  onMount(async () => {
    await loadModLog();
  });

  async function loadModLog() {
    loading = true;
    error = '';
    try {
      const data = await api.getModLog();
      modLog = data.length ? data : sampleModLog;
    } catch (e) {
      modLog = sampleModLog;
    } finally {
      loading = false;
    }
  }

  async function loadReports() {
    loading = true;
    error = '';
    try {
      const data = await api.getReports();
      reports = data.reports || data || [];
    } catch (e) {
      error = e.message || 'Failed to load reports';
      reports = [];
    } finally {
      loading = false;
    }
  }

  async function resolveReport(id, action) {
    resolving[id] = true;
    try {
      await api.resolveReport(id, action);
      reports = reports.filter(r => r.id !== id);
    } catch (e) {
      error = e.message || 'Failed to resolve report';
    } finally {
      resolving[id] = false;
    }
  }

  async function loadModNotes() {
    if (!noteUserId) return;
    loading = true;
    try {
      const data = await api.getModNotes(noteUserId);
      modNotes = data.notes || data || [];
    } catch (e) {
      modNotes = [];
    } finally {
      loading = false;
    }
  }

  async function addModNote() {
    if (!noteUserId || !noteBody.trim()) return;
    try {
      await api.createModNote(noteUserId, noteBody);
      noteBody = '';
      await loadModNotes();
    } catch (e) {
      error = e.message || 'Failed to add note';
    }
  }

  $: if (activeTab === 'reports' && !reportsLoaded) { reportsLoaded = true; loadReports(); }

  function relativeTime(iso) {
    const diff = Date.now() - new Date(iso).getTime();
    const mins = Math.floor(diff / 60000);
    if (mins < 1) return 'just now';
    if (mins < 60) return `${mins}m ago`;
    const hrs = Math.floor(mins / 60);
    if (hrs < 24) return `${hrs}h ago`;
    const days = Math.floor(hrs / 24);
    if (days < 30) return `${days}d ago`;
    return `${Math.floor(days / 30)}mo ago`;
  }

  function actionLabel(type) {
    const labels = {
      warning: 'Warning',
      timeout: 'Timeout',
      suspension: 'Suspension',
      ban: 'Ban',
      content_removed: 'Content Removed',
    };
    return labels[type] || type;
  }

  function actionClass(type) {
    const cls = {
      warning: 'badge-warning',
      timeout: 'badge-timeout',
      suspension: 'badge-suspension',
      ban: 'badge-ban',
      content_removed: 'badge-removed',
    };
    return cls[type] || '';
  }

  $: filteredLog = searchQuery
    ? modLog.filter(e => e.reason.toLowerCase().includes(searchQuery.toLowerCase()))
    : modLog;
</script>

<div class="mod-center">
  <div class="page-header">
    <h1>Moderation Center</h1>
    <button class="btn-icon" on:click={() => goTo && goTo('home')}>&times;</button>
  </div>

  <div class="tabs">
    {#each tabs as tab}
      <button
        class="tab"
        class:active={activeTab === tab.id}
        on:click={() => activeTab = tab.id}
      >
        {tab.label}
      </button>
    {/each}
  </div>

  {#if activeTab === 'modlog'}
    <div class="tab-content">
      <div class="search-bar">
        <input
          type="text"
          placeholder="Search by reason..."
          bind:value={searchQuery}
        />
      </div>

      {#if loading}
        <div class="loading">Loading...</div>
      {:else if error}
        <div class="error">{error}</div>
      {:else}
        <div class="mod-log-list">
          {#each filteredLog as entry (entry.id)}
            <div class="mod-action-card">
              <div class="card-top">
                <span class="badge {actionClass(entry.action_type)}">
                  {actionLabel(entry.action_type)}
                </span>
                <span class="timestamp">{relativeTime(entry.created_at)}</span>
              </div>
              <p class="reason">{entry.reason}</p>
              <div class="card-meta">
                <span><strong>Target:</strong> {entry.target_user}</span>
                <span><strong>Mod:</strong> {entry.moderator}</span>
              </div>
            </div>
          {/each}
          {#if filteredLog.length === 0}
            <p class="empty">No matching entries.</p>
          {/if}
        </div>
      {/if}
    </div>
  {:else if activeTab === 'reports'}
    <div class="tab-content">
      <h2>Reports</h2>
      {#if loading}
        <p class="loading">Loading reports...</p>
      {:else if error}
        <div class="error">{error}</div>
      {:else if reports.length === 0}
        <p class="placeholder">No pending reports.</p>
      {:else}
        <div class="reports-list">
          {#each reports as report (report.id)}
            <div class="mod-action-card">
              <div class="card-top">
                <span class="badge badge-warning">Report #{report.id}</span>
                <span class="timestamp">{relativeTime(report.created_at)}</span>
              </div>
              <p class="reason">Reported by <strong>@{report.reporter_username}</strong> against <strong>@{report.target_username}</strong></p>
              <p class="reason"><em>"{report.reason}"</em></p>
              {#if report.post_body}
                <div class="reported-post">
                  <p>{report.post_body}</p>
                </div>
              {/if}
              <div class="card-meta">
                <button class="btn btn-sm" on:click={() => resolveReport(report.id, 'dismiss')} disabled={resolving[report.id]}>
                  Dismiss
                </button>
                <button class="btn btn-sm btn-warning" on:click={() => resolveReport(report.id, 'warn')} disabled={resolving[report.id]}>
                  Warn user
                </button>
                <button class="btn btn-sm btn-danger" on:click={() => resolveReport(report.id, 'remove')} disabled={resolving[report.id]}>
                  Remove content
                </button>
              </div>
            </div>
          {/each}
        </div>
      {/if}
    </div>

  {:else if activeTab === 'notes'}
    <div class="tab-content">
      <h2>Mod Notes</h2>
      <div class="note-form">
        <input type="text" bind:value={noteUserId} placeholder="Username or user ID" />
        <textarea bind:value={noteBody} placeholder="Note body"></textarea>
        <button class="btn btn-primary" on:click={addModNote} disabled={!noteUserId || !noteBody.trim()}>Add note</button>
        <button class="btn" on:click={loadModNotes}>Load notes for user</button>
      </div>
      {#if modNotes.length > 0}
        <div class="notes-list">
          {#each modNotes as note}
            <div class="mod-action-card">
              <p class="reason">{note.body}</p>
              <span class="timestamp">Added by @{note.moderator_username} — {relativeTime(note.created_at)}</span>
            </div>
          {/each}
        </div>
      {/if}
    </div>

  {:else if activeTab === 'jury'}
    <div class="tab-content placeholder">
      <p>Jury Duty — coming soon.</p>
    </div>
  {:else if activeTab === 'flagqueue'}
    <div class="tab-content placeholder">
      <p>Flag Queue — coming soon.</p>
    </div>
  {/if}
</div>

<style>
  .mod-center {
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
    min-height: 200px;
  }

  .tab-content.placeholder {
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--text);
    font-size: 1rem;
  }

  .search-bar {
    margin-bottom: 1rem;
  }

  .search-bar input {
    width: 100%;
    padding: 0.625rem 0.875rem;
    border: 1px solid var(--border);
    border-radius: 6px;
    font-size: 0.875rem;
    outline: none;
    box-sizing: border-box;
    background: var(--bg);
    color: var(--text);
  }

  .search-bar input:focus {
    border-color: var(--accent);
  }

  .loading {
    text-align: center;
    color: var(--text);
    padding: 2rem 0;
  }

  .error {
    color: #ef4444;
    text-align: center;
    padding: 1rem;
    background: #fef2f2;
    border-radius: 6px;
    margin-bottom: 1rem;
  }

  .mod-log-list {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .mod-action-card {
    border: 1px solid var(--border);
    border-radius: 8px;
    padding: 1rem;
    background: var(--code-bg);
  }

  .card-top {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 0.5rem;
  }

  .badge {
    display: inline-block;
    padding: 0.2rem 0.6rem;
    border-radius: 999px;
    font-size: 0.75rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.03em;
  }

  .badge-warning { background: #fef3c7; color: #b45309; }
  .badge-timeout { background: #fde68a; color: #92400e; }
  .badge-suspension { background: #fed7aa; color: #9a3412; }
  .badge-ban { background: #fecaca; color: #b91c1c; }
  .badge-removed { background: #e5e7eb; color: #4b5563; }

  .timestamp {
    font-size: 0.8125rem;
    color: var(--text);
  }

  .reason {
    margin: 0 0 0.5rem;
    font-size: 0.9375rem;
    color: var(--text);
  }

  .card-meta {
    display: flex;
    gap: 1.25rem;
    font-size: 0.8125rem;
    color: var(--text);
  }

  .card-meta strong {
    font-weight: 500;
  }

  .empty {
    text-align: center;
    color: var(--text);
    padding: 2rem;
  }

  .placeholder {
    text-align: center;
    color: var(--text);
    padding: 2rem;
  }

  .reports-list {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .reported-post {
    background: var(--bg);
    border: 1px solid var(--border);
    border-radius: 4px;
    padding: 0.5rem 0.75rem;
    margin: 0.5rem 0;
    font-size: 0.875rem;
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

  .btn-primary {
    background: var(--accent);
    color: white;
    border-color: var(--accent);
  }

  .btn-warning {
    background: #f59e0b;
    color: white;
    border-color: #d97706;
  }

  .btn-danger {
    background: #ef4444;
    color: white;
    border-color: #dc2626;
  }

  .note-form {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    margin-bottom: 1rem;
  }

  .note-form input,
  .note-form textarea {
    padding: 0.5rem 0.75rem;
    border: 1px solid var(--border);
    border-radius: 6px;
    font-size: 0.875rem;
    background: var(--bg);
    color: var(--text);
  }

  .note-form textarea {
    min-height: 80px;
    resize: vertical;
  }

  .notes-list {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .error-msg {
    color: #ef4444;
    text-align: center;
    padding: 1rem;
  }
</style>

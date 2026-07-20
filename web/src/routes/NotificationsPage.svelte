<script>
  import { onMount } from 'svelte';
  import api from '../lib/api/client';

  export let goTo;

  let notifications = [];
  let loading = true;

  onMount(load);

  async function load() {
    loading = true;
    try {
      const data = await api.listNotifications();
      notifications = data.notifications || data || [];
    } catch (e) {
      notifications = [];
    } finally {
      loading = false;
    }
  }

  async function markRead(id) {
    try {
      await api.markNotificationRead(id);
      notifications = notifications.map(n =>
        n.id === id ? { ...n, read: true } : n
      );
    } catch (e) {
      console.error(e);
    }
  }

  async function markAllRead() {
    try {
      await api.markAllNotificationsRead();
      notifications = notifications.map(n => ({ ...n, read: true }));
    } catch (e) {
      console.error(e);
    }
  }

  function timeAgo(dateStr) {
    const now = Date.now();
    const then = new Date(dateStr).getTime();
    const diff = now - then;
    const mins = Math.floor(diff / 60000);
    if (mins < 1) return 'just now';
    if (mins < 60) return `${mins}m ago`;
    const hours = Math.floor(mins / 60);
    if (hours < 24) return `${hours}h ago`;
    const days = Math.floor(hours / 24);
    return `${days}d ago`;
  }

  function iconFor(type) {
    const icons = {
      like: '♥',
      repost: '♺',
      follow: '⊕',
      mention: '@',
      reply: '↩',
      warning: '⚠',
      system: '◆'
    };
    return icons[type] || '•';
  }
</script>

<div class="page">
  <div class="page-header">
    <h2>Notifications</h2>
    <button class="btn btn-sm" on:click={markAllRead}>Mark all read</button>
  </div>

  {#if loading}
    <p class="placeholder">Loading notifications...</p>
  {:else if notifications.length === 0}
    <p class="placeholder">No notifications yet.</p>
  {:else}
    <div class="notif-list">
      {#each notifications as notif}
        <div class="notif-card" class:unread={!notif.read} on:click={() => markRead(notif.id)}>
          <span class="notif-icon">{iconFor(notif.type)}</span>
          <div class="notif-body">
            <p>{@html notif.body || notif.message}</p>
            <span class="notif-time">{timeAgo(notif.created_at)}</span>
          </div>
        </div>
      {/each}
    </div>
  {/if}
</div>

<style>
  .page {
    max-width: 600px;
    margin: 0 auto;
    padding: 1rem;
  }

  .page-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 1rem;
  }

  .page-header h2 {
    margin: 0;
  }

  .placeholder {
    text-align: center;
    color: var(--text);
    opacity: 0.6;
    padding: 2rem;
  }

  .notif-list {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .notif-card {
    display: flex;
    align-items: flex-start;
    gap: 0.75rem;
    padding: 0.75rem;
    border-radius: 6px;
    cursor: pointer;
    transition: background 0.15s;
  }

  .notif-card:hover {
    background: var(--hover, #f0f0f0);
  }

  .notif-card.unread {
    background: var(--hover, #f5f5ff);
  }

  .notif-icon {
    font-size: 1.25rem;
    width: 2rem;
    height: 2rem;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 50%;
    background: var(--bg);
    flex-shrink: 0;
  }

  .notif-body {
    flex: 1;
    min-width: 0;
  }

  .notif-body p {
    margin: 0;
    font-size: 0.9rem;
    line-height: 1.4;
  }

  .notif-time {
    font-size: 0.75rem;
    opacity: 0.5;
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

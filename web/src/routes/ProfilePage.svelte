<script>
  import { onMount } from 'svelte';
  import { currentUser } from '../lib/stores/auth';
  import api from '../lib/api/client';

  export let goTo;
  export let user;

  let profile = null;
  let userPosts = [];
  let userAchievements = [];
  let loading = false;
  let tab = 'posts';

  $: username = user?.username || '';

  onMount(async () => {
    if (!username) return;
    loading = true;
    try {
      const data = await api.get(`/api/v1/users/${encodeURIComponent(username)}`);
      profile = data.user || data;
      const posts = await api.get(`/api/v1/users/${encodeURIComponent(username)}/posts`);
      userPosts = posts.posts || posts || [];
      const ach = await api.listAchievements();
      userAchievements = ach.achievements || ach || [];
    } catch (e) {
      // ignore
    } finally {
      loading = false;
    }
  });

  async function handleMessage() {
    goTo('messages', { username });
  }

  function formatTime(dateStr) {
    const d = new Date(dateStr);
    return d.toLocaleDateString();
  }
</script>

<div class="profile-page">
  {#if loading}
    <p class="placeholder">Loading profile...</p>
  {:else if profile}
    <div class="profile-header">
      <div class="profile-avatar">{profile.username?.charAt(0).toUpperCase() || '?'}</div>
      <div class="profile-info">
        <h2>{profile.display_name || profile.username}</h2>
        <p class="username">@{profile.username}</p>
        {#if profile.bio}
          <p class="bio">{profile.bio}</p>
        {/if}
        <div class="stats">
          <span><strong>{profile.post_count || 0}</strong> posts</span>
          <span><strong>{profile.follower_count || 0}</strong> followers</span>
          <span><strong>{profile.following_count || 0}</strong> following</span>
        </div>
        <div class="badges">
          {#each userAchievements.filter(a => a.unlocked) as ach}
            <span class="badge" title={ach.name}>{ach.icon || '🏅'}</span>
          {/each}
        </div>
      </div>
      <div class="profile-actions">
        <button class="msg-btn" on:click={handleMessage}>Message</button>
      </div>
    </div>

    <div class="tabs">
      <button class="tab" class:active={tab === 'posts'} on:click={() => tab = 'posts'}>Posts</button>
      <button class="tab" class:active={tab === 'achievements'} on:click={() => tab = 'achievements'}>Achievements</button>
    </div>

    {#if tab === 'posts'}
      <div class="post-list">
        {#each userPosts as post}
          <div class="post-card">
            <div class="post-time">{formatTime(post.created_at)}</div>
            <div class="post-body">{post.body}</div>
            <div class="post-stats">
              <span>▲ {post.upvotes || 0}</span>
              <span>♥ {post.likes || 0}</span>
              <span>💬 {post.engagement_count || 0}</span>
            </div>
          </div>
        {/each}
        {#if userPosts.length === 0}
          <p class="empty">No posts yet.</p>
        {/if}
      </div>
    {:else}
      <div class="achievement-grid">
        {#each userAchievements as ach}
          <div class="achievement-card" class:locked={!ach.unlocked}>
            <span class="ach-icon">{ach.icon || '🏅'}</span>
            <span class="ach-name">{ach.name}</span>
            {#if ach.description}
              <span class="ach-desc">{ach.description}</span>
            {/if}
          </div>
        {/each}
        {#if userAchievements.length === 0}
          <p class="empty">No achievements yet.</p>
        {/if}
      </div>
    {/if}
  {:else}
    <p class="placeholder">User not found.</p>
  {/if}
</div>

<style>
  .profile-page {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }
  .profile-header {
    display: flex;
    gap: 1rem;
    align-items: flex-start;
    background: var(--code-bg);
    padding: 1.5rem;
    border-radius: 12px;
  }
  .profile-avatar {
    width: 64px;
    height: 64px;
    border-radius: 50%;
    background: var(--accent);
    color: white;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 1.5rem;
    font-weight: 700;
    flex-shrink: 0;
  }
  .profile-info {
    flex: 1;
  }
  .profile-info h2 {
    margin: 0;
    font-size: 1.25rem;
  }
  .username {
    margin: 0;
    color: var(--text);
    font-size: 0.875rem;
  }
  .bio {
    margin: 0.5rem 0;
    font-size: 0.9rem;
    color: var(--text-h);
  }
  .stats {
    display: flex;
    gap: 1rem;
    font-size: 0.875rem;
    color: var(--text);
  }
  .badges {
    display: flex;
    gap: 0.25rem;
    margin-top: 0.5rem;
  }
  .badge {
    font-size: 1.25rem;
    cursor: help;
  }
  .profile-actions {
    display: flex;
    gap: 0.5rem;
  }
  .msg-btn {
    padding: 0.5rem 1rem;
    border: 1px solid var(--accent);
    border-radius: 6px;
    background: none;
    color: var(--accent);
    cursor: pointer;
    font-size: 0.875rem;
  }
  .msg-btn:hover {
    background: var(--accent);
    color: white;
  }
  .tabs {
    display: flex;
    gap: 0.5rem;
  }
  .tab {
    padding: 0.5rem 1rem;
    border: 1px solid var(--border);
    border-radius: 6px;
    background: var(--bg);
    color: var(--text);
    cursor: pointer;
    font-size: 0.875rem;
  }
  .tab.active {
    background: var(--accent);
    color: white;
    border-color: var(--accent);
  }
  .post-list {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }
  .post-card {
    background: var(--code-bg);
    padding: 1rem;
    border-radius: 8px;
  }
  .post-time {
    font-size: 0.8rem;
    color: var(--text);
    margin-bottom: 0.375rem;
  }
  .post-body {
    color: var(--text-h);
    margin-bottom: 0.5rem;
  }
  .post-stats {
    display: flex;
    gap: 1rem;
    font-size: 0.8rem;
    color: var(--text);
  }
  .achievement-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
    gap: 0.75rem;
  }
  .achievement-card {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.25rem;
    padding: 1rem;
    background: var(--code-bg);
    border-radius: 8px;
    text-align: center;
  }
  .achievement-card.locked {
    opacity: 0.4;
  }
  .ach-icon {
    font-size: 2rem;
  }
  .ach-name {
    font-weight: 600;
    font-size: 0.875rem;
  }
  .ach-desc {
    font-size: 0.75rem;
    color: var(--text);
  }
  .placeholder, .empty {
    text-align: center;
    padding: 2rem;
    color: var(--text);
  }
</style>

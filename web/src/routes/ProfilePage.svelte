<script>
  import { onMount } from 'svelte';
  import { currentUser, logout } from '../lib/stores/auth';
  import api from '../lib/api/client';

  export let goTo;
  export let user;

  let profile = null;
  let userPosts = [];
  let userAchievements = [];
  let loading = false;
  let tab = 'posts';
  let editing = false;
  let editDisplayName = '';
  let editBio = '';
  let isFollowing = false;
  let isBlocked = false;

  $: username = user?.username || '';
  $: isOwnProfile = $currentUser && profile && $currentUser.username === profile.username;

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
      isFollowing = profile.is_following || false;
      isBlocked = profile.is_blocked || false;
      editDisplayName = profile.display_name || '';
      editBio = profile.bio || '';
    } catch (e) {
      // ignore
    } finally {
      loading = false;
    }
  });

  async function handleMessage() {
    goTo('messages', { username });
  }

  async function toggleFollow() {
    try {
      if (isFollowing) {
        await api.unfollowUser(profile.username);
        isFollowing = false;
        profile.follower_count = (profile.follower_count || 1) - 1;
      } else {
        await api.followUser(profile.username);
        isFollowing = true;
        profile.follower_count = (profile.follower_count || 0) + 1;
      }
    } catch (e) {
      console.error(e);
    }
  }

  async function toggleBlock() {
    try {
      if (isBlocked) {
        await api.unblockUser(profile.username);
        isBlocked = false;
      } else {
        await api.blockUser(profile.username);
        isBlocked = true;
      }
    } catch (e) {
      console.error(e);
    }
  }

  async function saveProfile() {
    try {
      const updated = await api.updateProfile({ display_name: editDisplayName, bio: editBio });
      profile = { ...profile, ...updated };
      editing = false;
    } catch (e) {
      console.error(e);
    }
  }

  async function handleAvatarUpload(e) {
    const file = e.target.files[0];
    if (!file) return;
    try {
      const res = await api.uploadAvatar(file);
      profile = { ...profile, avatar_url: res.url };
    } catch (err) {
      console.error(err);
    }
  }

  async function handleBannerUpload(e) {
    const file = e.target.files[0];
    if (!file) return;
    try {
      const res = await api.uploadBanner(file);
      profile = { ...profile, banner_url: res.url };
    } catch (err) {
      console.error(err);
    }
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
    <div class="profile-banner" style={profile.banner_url ? `background-image: url(${profile.banner_url})` : ''}>
      {#if isOwnProfile}
        <label class="banner-upload">
          <input type="file" accept="image/*" on:change={handleBannerUpload} />
          Change banner
        </label>
      {/if}
    </div>
    <div class="profile-header">
      <div class="avatar-section">
        {#if profile.avatar_url}
          <img src={profile.avatar_url} alt="" class="avatar" />
        {:else}
          <div class="profile-avatar">{profile.username?.charAt(0).toUpperCase() || '?'}</div>
        {/if}
        {#if isOwnProfile}
          <label class="avatar-upload">
            <input type="file" accept="image/*" on:change={handleAvatarUpload} />
            Change
          </label>
        {/if}
      </div>
      <div class="profile-info">
        {#if editing}
          <input bind:value={editDisplayName} placeholder="Display name" class="edit-input" />
          <textarea bind:value={editBio} placeholder="Bio" class="edit-textarea"></textarea>
          <div class="edit-actions">
            <button class="btn" on:click={saveProfile}>Save</button>
            <button class="btn" on:click={() => { editing = false; }}>Cancel</button>
          </div>
        {:else}
          <h2>{profile.display_name || profile.username}</h2>
          <p class="username">@{profile.username}</p>
          {#if profile.bio}
            <div class="bio">{@html profile.bio}</div>
          {/if}
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
        {#if isOwnProfile}
          <button class="btn" on:click={() => { editing = !editing; }}>Edit profile</button>
          <button class="btn" on:click={() => { logout(); goTo('login'); }}>Log out</button>
        {:else}
          <button class="msg-btn" on:click={handleMessage}>Message</button>
          <button class="btn" class:active={isFollowing} on:click={toggleFollow}>
            {isFollowing ? 'Following' : 'Follow'}
          </button>
          <button class="btn btn-danger" class:active={isBlocked} on:click={toggleBlock}>
            {isBlocked ? 'Blocked' : 'Block'}
          </button>
        {/if}
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
  .profile-banner {
    height: 200px;
    background: var(--code-bg);
    background-size: cover;
    background-position: center;
    border-radius: 8px;
    margin-bottom: -3rem;
    position: relative;
    display: flex;
    align-items: flex-end;
    justify-content: flex-end;
    padding: 0.5rem;
  }

  .banner-upload {
    font-size: 0.75rem;
    background: rgba(0,0,0,0.5);
    color: white;
    padding: 0.25rem 0.5rem;
    border-radius: 4px;
    cursor: pointer;
  }

  .banner-upload input {
    display: none;
  }

  .avatar-section {
    position: relative;
    display: flex;
    flex-direction: column;
    align-items: center;
  }

  .avatar {
    width: 80px;
    height: 80px;
    border-radius: 50%;
    object-fit: cover;
    border: 3px solid var(--bg);
  }

  .avatar-upload {
    font-size: 0.75rem;
    color: var(--accent);
    cursor: pointer;
    margin-top: 0.25rem;
  }

  .avatar-upload input {
    display: none;
  }

  .edit-input {
    display: block;
    width: 100%;
    padding: 0.5rem;
    margin-bottom: 0.5rem;
    border: 1px solid var(--border);
    border-radius: 6px;
    font-size: 0.9rem;
    background: var(--bg);
    color: var(--text);
    box-sizing: border-box;
  }

  .edit-textarea {
    display: block;
    width: 100%;
    padding: 0.5rem;
    margin-bottom: 0.5rem;
    border: 1px solid var(--border);
    border-radius: 6px;
    font-size: 0.9rem;
    background: var(--bg);
    color: var(--text);
    min-height: 80px;
    resize: vertical;
    box-sizing: border-box;
  }

  .edit-actions {
    display: flex;
    gap: 0.5rem;
    margin-top: 0.5rem;
  }

  .bio {
    line-height: 1.5;
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

  .btn.active {
    background: var(--accent);
    color: white;
    border-color: var(--accent);
  }

  .btn-danger {
    color: #ef4444;
    border-color: #ef4444;
  }

  .btn-danger.active {
    background: #ef4444;
    color: white;
  }

  .placeholder, .empty {
    text-align: center;
    padding: 2rem;
    color: var(--text);
  }
</style>

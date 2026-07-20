<script>
  import { onMount } from 'svelte';
  import { currentUser } from '../lib/stores/auth';
  import { feedPosts, feedLoading, feedError, currentMood, loadFeed, setMood } from '../lib/stores/feed';
  import api from '../lib/api/client';

  export let goTo;

  let postBody = '';
  let mood = 'all';
  let postType = 'post';
  let submitting = false;
  let openMenu = null;
  let showReportModal = false;
  let reportPostId = null;
  let reportReason = '';
  let savedPosts = new Set();
  let hiddenPosts = new Set();

  const moods = [
    { id: 'all', label: 'All' },
    { id: 'serious', label: 'Serious' },
    { id: 'wholesome', label: 'Wholesome' },
    { id: 'humor', label: 'Humor' },
    { id: 'support', label: 'Support' },
    { id: 'meta', label: 'Meta' },
  ];

  $: mood = $currentMood;

  async function handleMoodChange(m) {
    setMood(m);
  }

  async function handleSubmit() {
    if (!postBody.trim()) return;
    submitting = true;
    try {
      await api.createPost({ body: postBody, mood, type: postType });
      postBody = '';
      loadFeed(mood);
    } catch (e) {
      // ignore
    } finally {
      submitting = false;
    }
  }

  async function handleInteract(postId, type, metadata) {
    try {
      await api.interact(postId, type, metadata);
      loadFeed(mood);
    } catch (e) {
      // ignore
    }
  }

  function formatTime(dateStr) {
    const d = new Date(dateStr);
    const now = new Date();
    const diff = now - d;
    if (diff < 60000) return 'just now';
    if (diff < 3600000) return Math.floor(diff / 60000) + 'm';
    if (diff < 86400000) return Math.floor(diff / 3600000) + 'h';
    return d.toLocaleDateString();
  }

  function toggleMenu(postId) {
    openMenu = openMenu === postId ? null : postId;
  }

  async function handleSave(postId) {
    try {
      if (savedPosts.has(postId)) {
        await api.unsavePost(postId);
        savedPosts.delete(postId);
      } else {
        await api.savePost(postId);
        savedPosts.add(postId);
      }
    } catch (e) {
      console.error(e);
    }
    openMenu = null;
  }

  async function handleHide(postId) {
    try {
      if (hiddenPosts.has(postId)) {
        await api.unhidePost(postId);
        hiddenPosts.delete(postId);
      } else {
        await api.hidePost(postId);
        hiddenPosts.add(postId);
      }
    } catch (e) {
      console.error(e);
    }
    openMenu = null;
  }

  function openReport(postId) {
    reportPostId = postId;
    reportReason = '';
    showReportModal = true;
    openMenu = null;
  }

  async function submitReport() {
    if (!reportReason.trim()) return;
    try {
      await api.reportPost(reportPostId, reportReason);
      showReportModal = false;
      reportPostId = null;
      reportReason = '';
    } catch (e) {
      console.error(e);
    }
  }

  onMount(() => {
    loadFeed(mood);
  });
</script>

<div class="feed-page">
  <div class="compose-card">
    <textarea
      bind:value={postBody}
      placeholder="What's on your mind?"
      rows="3"
    ></textarea>
    <div class="compose-actions">
      <select bind:value={postType}>
        <option value="post">Post</option>
        <option value="question">Question</option>
        <option value="suggestion">Suggestion</option>
      </select>
      <button type="submit" on:click={handleSubmit} disabled={submitting || !postBody.trim()}>
        {submitting ? 'Posting...' : 'Post'}
      </button>
    </div>
  </div>

  <div class="mood-bar">
    {#each moods as m}
      <button
        class="mood-chip"
        class:active={mood === m.id}
        on:click={() => handleMoodChange(m.id)}
      >{m.label}</button>
    {/each}
  </div>

  {#if $feedLoading}
    <p class="placeholder">Loading feed...</p>
  {:else if $feedError}
    <p class="error">{$feedError}</p>
  {#if showReportModal}
    <div class="modal-overlay" on:click={() => { showReportModal = false; }} role="presentation">
      <div class="modal" on:click|stopPropagation>
        <h3>Report post</h3>
        <textarea bind:value={reportReason} placeholder="Why are you reporting this post?" rows="3"></textarea>
        <div class="modal-actions">
          <button class="btn" on:click={() => { showReportModal = false; }}>Cancel</button>
          <button class="btn btn-danger" on:click={submitReport} disabled={!reportReason.trim()}>Submit report</button>
        </div>
      </div>
    </div>
  {/if}

  {:else if $feedPosts.length === 0}
    <p class="placeholder">No posts yet. Follow some users to see their posts.</p>
  {:else}
    <div class="post-list">
      {#each $feedPosts as post}
        <div class="post-card" class:hidden={hiddenPosts.has(post.id)}>
          <div class="post-header">
            <span class="post-author" on:click={() => goTo('profile', { username: post.author || $currentUser?.username })}>
              {post.author_username || post.author || 'unknown'}
            </span>
            <span class="post-mood" class:mood-serious={post.mood === 'serious'} class:mood-wholesome={post.mood === 'wholesome'} class:mood-humor={post.mood === 'humor'} class:mood-support={post.mood === 'support'}>
              {post.mood}
            </span>
            <span class="post-indicators">
              {#if post.locked}
                <span class="indicator locked" title="Locked">&#x1f512;</span>
              {/if}
              {#if post.sticky}
                <span class="indicator sticky" title="Pinned">&#x1f4cc;</span>
              {/if}
              {#if post.edited}
                <span class="indicator edited" title="Edited">edited</span>
              {/if}
            </span>
            <span class="post-time">{formatTime(post.created_at)}</span>
            <div class="post-menu-wrap">
              <button class="menu-trigger" on:click={() => toggleMenu(post.id)}>&#x22ef;</button>
              {#if openMenu === post.id}
                <div class="dropdown-menu">
                  <button on:click={() => handleSave(post.id)}>
                    {savedPosts.has(post.id) ? 'Unsave' : 'Save'}
                  </button>
                  <button on:click={() => handleHide(post.id)}>
                    {hiddenPosts.has(post.id) ? 'Unhide' : 'Hide'}
                  </button>
                  <button on:click={() => openReport(post.id)}>Report</button>
                </div>
              {/if}
            </div>
          </div>
          <div class="post-body">{post.body}</div>
          {#if post.tags && post.tags.length > 0}
            <div class="post-tags">
              {#each post.tags as tag}
                <span class="tag">{tag}</span>
              {/each}
            </div>
          {/if}
          <div class="post-actions">
            <button class="action-btn" on:click={() => handleInteract(post.id, 'upvote')}>
              ▲ {post.upvotes || 0}
            </button>
            <button class="action-btn" on:click={() => handleInteract(post.id, 'downvote')}>
              ▼ {post.downvotes || 0}
            </button>
            <button class="action-btn" on:click={() => handleInteract(post.id, 'like')}>
              ♥ {post.likes || 0}
            </button>
            <button class="action-btn" on:click={() => handleInteract(post.id, 'engage')}>
              💬 {post.engagement_count || 0}
            </button>
            <button
              class="action-btn save-btn"
              on:click={() => handleInteract(post.id, 'save')}
            >🔖</button>
          </div>
        </div>
      {/each}
    </div>
  {/if}
</div>

<style>
  .feed-page {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }
  .compose-card {
    background: var(--code-bg);
    border-radius: 8px;
    padding: 1rem;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }
  textarea {
    width: 100%;
    resize: vertical;
    padding: 0.5rem;
    border: 1px solid var(--border);
    border-radius: 6px;
    font-size: 1rem;
    background: var(--bg);
    color: var(--text-h);
    font-family: inherit;
  }
  textarea:focus {
    outline: 2px solid var(--accent);
    outline-offset: -1px;
  }
  .compose-actions {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 0.5rem;
  }
  select {
    padding: 0.375rem 0.5rem;
    border: 1px solid var(--border);
    border-radius: 6px;
    background: var(--bg);
    color: var(--text-h);
    font-size: 0.875rem;
  }
  .compose-actions button {
    padding: 0.5rem 1rem;
    background: var(--accent);
    color: white;
    border: none;
    border-radius: 6px;
    cursor: pointer;
    font-size: 0.875rem;
  }
  .compose-actions button:disabled {
    opacity: 0.5;
  }
  .mood-bar {
    display: flex;
    gap: 0.5rem;
    flex-wrap: wrap;
  }
  .mood-chip {
    padding: 0.375rem 0.75rem;
    border: 1px solid var(--border);
    border-radius: 20px;
    background: var(--bg);
    color: var(--text);
    cursor: pointer;
    font-size: 0.8rem;
    transition: all 0.2s;
  }
  .mood-chip.active {
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
    border-radius: 8px;
    padding: 1rem;
  }
  .post-card.hidden {
    opacity: 0.4;
  }
  .post-header {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin-bottom: 0.5rem;
    font-size: 0.875rem;
  }
  .post-author {
    font-weight: 600;
    color: var(--accent);
    cursor: pointer;
  }
  .post-author:hover {
    text-decoration: underline;
  }
  .post-mood {
    font-size: 0.7rem;
    padding: 0.125rem 0.375rem;
    border-radius: 4px;
    background: var(--border);
  }
  .mood-serious { background: #e0e7ff; color: #4338ca; }
  .mood-wholesome { background: #d1fae5; color: #059669; }
  .mood-humor { background: #fef3c7; color: #d97706; }
  .mood-support { background: #fce7f3; color: #db2777; }
  .post-time {
    color: var(--text);
    font-size: 0.8rem;
  }

  .post-indicators {
    display: inline-flex;
    gap: 0.25rem;
    margin: 0 0.5rem;
  }

  .indicator {
    font-size: 0.75rem;
    color: var(--text);
    opacity: 0.7;
  }

  .indicator.edited {
    font-style: italic;
    font-size: 0.6875rem;
    color: var(--text);
    opacity: 0.5;
  }

  .post-menu-wrap {
    position: relative;
    margin-left: auto;
  }

  .menu-trigger {
    background: none;
    border: none;
    cursor: pointer;
    font-size: 1.25rem;
    color: var(--text);
    opacity: 0.6;
    padding: 0 0.25rem;
  }

  .menu-trigger:hover {
    opacity: 1;
  }

  .dropdown-menu {
    position: absolute;
    right: 0;
    top: 100%;
    background: var(--bg);
    border: 1px solid var(--border);
    border-radius: 6px;
    box-shadow: 0 4px 12px rgba(0,0,0,0.1);
    z-index: 10;
    min-width: 120px;
    display: flex;
    flex-direction: column;
  }

  .dropdown-menu button {
    background: none;
    border: none;
    padding: 0.5rem 0.75rem;
    text-align: left;
    cursor: pointer;
    font-size: 0.875rem;
    color: var(--text);
  }

  .dropdown-menu button:hover {
    background: var(--hover);
  }

  .modal-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0,0,0,0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 100;
  }

  .modal {
    background: var(--bg);
    border: 1px solid var(--border);
    border-radius: 8px;
    padding: 1.5rem;
    width: 90%;
    max-width: 400px;
  }

  .modal h3 {
    margin: 0 0 0.75rem 0;
  }

  .modal textarea {
    width: 100%;
    padding: 0.5rem;
    border: 1px solid var(--border);
    border-radius: 6px;
    font-size: 0.875rem;
    resize: vertical;
    background: var(--bg);
    color: var(--text);
    box-sizing: border-box;
  }

  .modal-actions {
    display: flex;
    gap: 0.5rem;
    justify-content: flex-end;
    margin-top: 0.75rem;
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

  .btn-danger {
    background: #ef4444;
    color: white;
    border-color: #dc2626;
  }
  .post-body {
    margin-bottom: 0.5rem;
    line-height: 1.5;
    color: var(--text-h);
    word-wrap: break-word;
  }
  .post-tags {
    display: flex;
    gap: 0.375rem;
    flex-wrap: wrap;
    margin-bottom: 0.5rem;
  }
  .tag {
    padding: 0.125rem 0.375rem;
    background: var(--accent-bg);
    color: var(--accent);
    border-radius: 4px;
    font-size: 0.75rem;
  }
  .post-actions {
    display: flex;
    gap: 0.5rem;
    flex-wrap: wrap;
  }
  .action-btn {
    padding: 0.25rem 0.5rem;
    border: none;
    background: none;
    color: var(--text);
    cursor: pointer;
    font-size: 0.8rem;
    border-radius: 4px;
    transition: background 0.2s;
  }
  .action-btn:hover {
    background: var(--accent-bg);
    color: var(--accent);
  }
  .save-btn {
    margin-left: auto;
  }
  .placeholder, .error {
    text-align: center;
    padding: 2rem;
    color: var(--text);
  }
  .error {
    color: #dc2626;
  }
</style>

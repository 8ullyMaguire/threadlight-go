<script>
  import { createEventDispatcher, onMount } from 'svelte';
  import { currentUser } from '../lib/stores/auth';
  import api from '../lib/api/client';

  export let goTo;

  const dispatch = createEventDispatcher();

  let step = 1;
  let loading = false;
  let selectedTags = [];
  let availableTags = [];
  let suggestedUsers = [];
  let followedUsers = [];
  let tagSearch = '';
  let displayName = '';
  let bio = '';

  onMount(async () => {
    try {
      const progress = await api.getOnboardingProgress();
      if (progress.stage) step = progress.stage;
    } catch (e) {
      // ignore
    }
  });

  async function searchTags() {
    if (!tagSearch.trim()) return;
    const results = await api.searchTags(tagSearch);
    availableTags = results.tags || results || [];
  }

  function toggleTag(tag) {
    if (selectedTags.find(t => t.id === tag.id)) {
      selectedTags = selectedTags.filter(t => t.id !== tag.id);
    } else {
      selectedTags = [...selectedTags, tag];
    }
  }

  async function nextStep() {
    loading = true;
    try {
      if (step === 1) {
        await api.pickOnboardingTags(selectedTags.map(t => t.id));
        const users = await api.getSuggestedUsers();
        suggestedUsers = users.users || users || [];
      } else if (step === 2) {
        await api.advanceOnboarding(3);
      } else if (step === 3) {
        await api.advanceOnboarding(4);
        dispatch('complete');
        return;
      }
      step++;
      await api.advanceOnboarding(step);
    } catch (e) {
      // ignore
    } finally {
      loading = false;
    }
  }

  function toggleFollow(user) {
    if (followedUsers.find(u => u.id === user.id)) {
      followedUsers = followedUsers.filter(u => u.id !== user.id);
    } else {
      followedUsers = [...followedUsers, user];
    }
  }
</script>

<div class="onboarding">
  <div class="step-indicator">
    {#each [1, 2, 3, 4] as s}
      <span class="step-dot" class:active={step >= s} class:done={step > s}>{s}</span>
    {/each}
  </div>

  {#if step === 1}
    <div class="step-content">
      <h2>Pick your interests</h2>
      <p>Select tags that match what you'd like to see.</p>
      <div class="tag-search">
        <input
          type="text"
          bind:value={tagSearch}
          placeholder="Search tags..."
          on:input={searchTags}
        />
      </div>
      <div class="tag-grid">
        {#each availableTags as tag}
          <button
            class="tag-option"
            class:selected={selectedTags.find(t => t.id === tag.id)}
            on:click={() => toggleTag(tag)}
          >{tag.name}</button>
        {/each}
      </div>
      {#if selectedTags.length > 0}
        <div class="selected-tags">
          {#each selectedTags as tag}
            <span class="selected-tag">{tag.name} <button class="remove-tag" on:click={() => toggleTag(tag)}>×</button></span>
          {/each}
        </div>
      {/if}
    </div>
  {:else if step === 2}
    <div class="step-content">
      <h2>Follow some people</h2>
      <p>Find users with similar interests.</p>
      <div class="user-list">
        {#each suggestedUsers as user}
          <div class="user-card">
            <span class="user-name">{user.username}</span>
            <button
              class="follow-btn"
              class:following={followedUsers.find(u => u.id === user.id)}
              on:click={() => toggleFollow(user)}
            >
              {followedUsers.find(u => u.id === user.id) ? 'Following' : 'Follow'}
            </button>
          </div>
        {/each}
      </div>
    </div>
  {:else if step === 3}
    <div class="step-content">
      <h2>Set up your profile</h2>
      <p>Tell people a bit about yourself.</p>
      <label>
        Display Name
        <input type="text" bind:value={displayName} placeholder="Your name" />
      </label>
      <label>
        Bio
        <textarea bind:value={bio} placeholder="A short bio..." rows="3"></textarea>
      </label>
    </div>
  {:else if step === 4}
    <div class="step-content">
      <h2>You're all set!</h2>
      <p>Welcome to Polaris. Let's see what's in your feed.</p>
    </div>
  {/if}

  <div class="step-actions">
    {#if step > 1 && step < 5}
      <button class="secondary" on:click={() => { step--; }} disabled={loading}>Back</button>
    {/if}
    <button class="primary" on:click={nextStep} disabled={loading}>
      {loading ? 'Loading...' : step === 4 ? 'Finish' : 'Next'}
    </button>
  </div>
</div>

<style>
  .onboarding {
    max-width: 500px;
    margin: 2rem auto;
    padding: 2rem;
    background: var(--code-bg);
    border-radius: 12px;
  }
  .step-indicator {
    display: flex;
    justify-content: center;
    gap: 1rem;
    margin-bottom: 2rem;
  }
  .step-dot {
    width: 32px;
    height: 32px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    border: 2px solid var(--border);
    color: var(--text);
    font-weight: 600;
    font-size: 0.875rem;
  }
  .step-dot.active {
    border-color: var(--accent);
    background: var(--accent);
    color: white;
  }
  .step-dot.done {
    border-color: var(--accent);
    color: var(--accent);
  }
  .step-content {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    margin-bottom: 2rem;
  }
  .step-content h2 {
    margin: 0;
    font-size: 1.25rem;
  }
  .step-content p {
    margin: 0;
    color: var(--text);
  }
  .tag-search input {
    width: 100%;
    padding: 0.5rem;
    border: 1px solid var(--border);
    border-radius: 6px;
    background: var(--bg);
    color: var(--text-h);
    font-size: 0.9rem;
  }
  .tag-grid {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
  }
  .tag-option {
    padding: 0.375rem 0.75rem;
    border: 1px solid var(--border);
    border-radius: 20px;
    background: var(--bg);
    color: var(--text);
    cursor: pointer;
    font-size: 0.8rem;
    transition: all 0.2s;
  }
  .tag-option.selected {
    background: var(--accent);
    color: white;
    border-color: var(--accent);
  }
  .selected-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 0.375rem;
  }
  .selected-tag {
    display: flex;
    align-items: center;
    gap: 0.25rem;
    padding: 0.25rem 0.5rem;
    background: var(--accent-bg);
    color: var(--accent);
    border-radius: 20px;
    font-size: 0.8rem;
  }
  .remove-tag {
    border: none;
    background: none;
    color: var(--accent);
    cursor: pointer;
    font-size: 1.1rem;
    line-height: 1;
    padding: 0;
  }
  .user-list {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }
  .user-card {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0.75rem;
    background: var(--bg);
    border-radius: 6px;
  }
  .user-name {
    font-weight: 500;
  }
  .follow-btn {
    padding: 0.375rem 0.75rem;
    border: 1px solid var(--accent);
    border-radius: 6px;
    background: none;
    color: var(--accent);
    cursor: pointer;
    font-size: 0.8rem;
  }
  .follow-btn.following {
    background: var(--accent);
    color: white;
  }
  label {
    display: flex;
    flex-direction: column;
    gap: 0.375rem;
    font-size: 0.875rem;
    color: var(--text);
  }
  label input, label textarea {
    padding: 0.5rem;
    border: 1px solid var(--border);
    border-radius: 6px;
    background: var(--bg);
    color: var(--text-h);
    font-family: inherit;
  }
  .step-actions {
    display: flex;
    justify-content: space-between;
    gap: 1rem;
  }
  .step-actions button {
    padding: 0.5rem 1.5rem;
    border-radius: 6px;
    font-size: 0.875rem;
    cursor: pointer;
  }
  .primary {
    background: var(--accent);
    color: white;
    border: none;
    margin-left: auto;
  }
  .primary:disabled {
    opacity: 0.5;
  }
  .secondary {
    border: 1px solid var(--border);
    background: var(--bg);
    color: var(--text);
  }
</style>

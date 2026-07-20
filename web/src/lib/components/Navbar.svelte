<script>
  import { currentUser, isAuthenticated, logout } from '../stores/auth';
  import { unreadCount } from '../stores/notifications';

  export let goTo;
  export let current;

  const pages = [
    { id: 'feed', label: 'Feed', icon: 'home' },
    { id: 'search', label: 'Search', icon: 'search' },
    { id: 'communities', label: 'Communities', icon: 'users' },
    { id: 'collections', label: 'Collections', icon: 'bookmark' },
    { id: 'leaderboard', label: 'Leaderboard', icon: 'trophy' },
    { id: 'profile', label: 'Profile', icon: 'user' },
    { id: 'mod', label: 'Mod Center', icon: 'shield' },
    { id: 'admin', label: 'Admin', icon: 'settings' },
  ];

  function handleLogout() {
    logout();
    goTo('login');
  }
</script>

<nav class="navbar">
  <div class="nav-brand" on:click={() => goTo('feed')} on:keydown={() => {}} role="button" tabindex="0">
    <span class="logo">Threadlight</span>
  </div>
  <div class="nav-links">
    {#each pages as p}
      <button
        class="nav-link"
        class:active={current === p.id}
        on:click={() => goTo(p.id)}
      >
        <span class="nav-icon">{p.icon}</span>
        <span class="nav-label">{p.label}</span>
      </button>
    {/each}
  </div>
  <div class="nav-actions">
    {#if $unreadCount > 0}
      <span class="badge">{$unreadCount}</span>
    {/if}
    <button class="logout-btn" on:click={handleLogout}>Log out</button>
  </div>
</nav>

<style>
  .navbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0.5rem 1rem;
    border-bottom: 1px solid var(--border);
    background: var(--bg);
    position: sticky;
    top: 0;
    z-index: 100;
  }
  .nav-brand {
    cursor: pointer;
    font-family: var(--heading);
    font-size: 1.5rem;
    font-weight: 600;
    color: var(--accent);
  }
  .nav-links {
    display: flex;
    gap: 0.25rem;
  }
  .nav-link {
    display: flex;
    align-items: center;
    gap: 0.25rem;
    padding: 0.5rem 0.75rem;
    border: none;
    background: none;
    color: var(--text);
    cursor: pointer;
    border-radius: 6px;
    font-size: 0.875rem;
    transition: background 0.2s, color 0.2s;
  }
  .nav-link:hover {
    background: var(--accent-bg);
    color: var(--accent);
  }
  .nav-link.active {
    background: var(--accent-bg);
    color: var(--accent);
    font-weight: 500;
  }
  .nav-icon {
    font-size: 1rem;
  }
  .nav-label {
    display: none;
  }
  @media (min-width: 640px) {
    .nav-label {
      display: inline;
    }
  }
  .nav-actions {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }
  .badge {
    background: var(--accent);
    color: white;
    border-radius: 50%;
    width: 20px;
    height: 20px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 0.75rem;
    font-weight: 600;
  }
  .logout-btn {
    padding: 0.375rem 0.75rem;
    border: 1px solid var(--border);
    background: none;
    color: var(--text);
    border-radius: 6px;
    cursor: pointer;
    font-size: 0.875rem;
    transition: background 0.2s, color 0.2s;
  }
  .logout-btn:hover {
    background: var(--code-bg);
    color: var(--text-h);
  }
</style>

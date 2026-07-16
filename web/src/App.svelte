<script>
  import { onMount } from 'svelte';
  import { currentUser, isAuthenticated, checkAuth } from './lib/stores/auth';
  import LoginPage from './routes/LoginPage.svelte';
  import RegisterPage from './routes/RegisterPage.svelte';
  import FeedPage from './routes/FeedPage.svelte';
  import OnboardingWizard from './routes/OnboardingWizard.svelte';
  import Navbar from './lib/components/Navbar.svelte';
  import ModCenter from './routes/ModCenter.svelte';
  import ProfilePage from './routes/ProfilePage.svelte';
  import SearchPage from './routes/SearchPage.svelte';
  import CommunitiesPage from './routes/CommunitiesPage.svelte';
  import CollectionsPage from './routes/CollectionsPage.svelte';
  import SettingsPage from './routes/SettingsPage.svelte';
  import LeaderboardPage from './routes/LeaderboardPage.svelte';
  import AdminPage from './routes/AdminPage.svelte';
  import TrendingSidebar from './lib/components/TrendingSidebar.svelte';

  let page = 'feed';
  let params = {};

  function navigate(e) {
    const { detail } = e;
    page = detail.page || 'feed';
    params = detail.params || {};
  }

  function goTo(p, p2) {
    page = p;
    params = p2 || {};
  }

  $: onboardingIncomplete = $currentUser && $currentUser.onboarding_stage != null && $currentUser.onboarding_stage < 4;

  onMount(() => {
    checkAuth();
    window.addEventListener('navigate', navigate);
    return () => window.removeEventListener('navigate', navigate);
  });
</script>

{#if !$isAuthenticated}
  {#if page === 'register'}
    <RegisterPage on:register={() => goTo('feed')} on:nav={(e) => goTo(e.detail.page)} />
  {:else}
    <LoginPage on:login={() => goTo('feed')} on:nav={(e) => goTo(e.detail.page)} />
  {/if}
{:else if onboardingIncomplete}
  <OnboardingWizard on:complete={() => goTo('feed')} />
{:else}
  <Navbar {goTo} current={page} />
  <div class="app-layout">
    <main class="main-content">
      {#if page === 'feed'}
        <FeedPage {goTo} />
      {:else if page === 'search'}
        <SearchPage {goTo} />
      {:else if page === 'communities'}
        <CommunitiesPage {goTo} />
      {:else if page === 'collections'}
        <CollectionsPage {goTo} />
      {:else if page === 'profile'}
        <ProfilePage user={$currentUser} {goTo} />
      {:else if page === 'mod'}
        <ModCenter {goTo} />
      {:else if page === 'settings'}
        <SettingsPage />
      {:else if page === 'leaderboard'}
        <LeaderboardPage />
      {:else if page === 'admin'}
        <AdminPage />
      {:else}
        <FeedPage {goTo} />
      {/if}
    </main>
    <aside class="sidebar">
      <TrendingSidebar {goTo} />
    </aside>
  </div>
{/if}

<style>
  .app-layout {
    display: flex;
    max-width: 1200px;
    margin: 0 auto;
    padding: 1rem;
    gap: 1rem;
  }
  .main-content {
    flex: 1;
    min-width: 0;
  }
  .sidebar {
    width: 300px;
    flex-shrink: 0;
  }
  @media (max-width: 768px) {
    .app-layout {
      flex-direction: column;
    }
    .sidebar {
      width: 100%;
    }
  }
</style>

<script>
  import api from '../lib/api/client';
  let version = '0.1';

  api.get('/api/v1/about').then(d => { if (d) version = d.version || '0.1'; }).catch(() => {});

  import { onMount } from 'svelte';

  export let goTo;

  function nav(page) {
    window.dispatchEvent(new CustomEvent('navigate', { detail: { page } }));
  }
</script>

<footer class="footer">
  <div class="footer-inner">
    <div class="footer-section">
      <h4>Discover</h4>
      <button class="link" on:click={() => nav('feed')}>Feed</button>
      <button class="link" on:click={() => nav('communities')}>Communities</button>
      <button class="link" on:click={() => nav('search')}>Explore</button>
      <button class="link" on:click={() => nav('leaderboard')}>Leaderboard</button>
    </div>
    <div class="footer-section">
      <h4>Account</h4>
      <button class="link" on:click={() => nav('profile')}>Profile</button>
      <button class="link" on:click={() => nav('collections')}>Collections</button>
      <button class="link" on:click={() => nav('settings')}>Settings</button>
      <button class="link" on:click={() => nav('mod')}>Mod Center</button>
    </div>
    <div class="footer-section">
      <h4>Support</h4>
      <button class="link" on:click={() => nav('about')}>About</button>
      <a class="link" href="/donate" target="_blank" rel="noopener">Donate</a>
    </div>
    <div class="footer-section version">
      <p>Polaris v{version}</p>
      <p class="muted">Built for positive connection</p>
    </div>
  </div>
</footer>

<style>
  .footer {
    background: #f9fafb;
    border-top: 1px solid var(--border);
    margin-top: 2rem;
    padding: 2rem 1rem 1rem;
  }
  .footer-inner {
    max-width: 1200px;
    margin: 0 auto;
    display: flex;
    gap: 3rem;
    flex-wrap: wrap;
  }
  .footer-section {
    display: flex;
    flex-direction: column;
    gap: 0.4rem;
  }
  .footer-section h4 {
    margin: 0 0 0.3rem;
    font-size: 0.85rem;
    text-transform: uppercase;
    color: var(--muted);
    letter-spacing: 0.05em;
  }
  .link {
    background: none;
    border: none;
    color: var(--text);
    cursor: pointer;
    font-size: 0.9rem;
    text-align: left;
    padding: 0;
  }
  .link:hover { color: var(--primary); text-decoration: underline; }
  a.link { color: var(--text); text-decoration: none; }
  a.link:hover { color: var(--primary); text-decoration: underline; }
  .version { margin-left: auto; text-align: right; }
  .version p { margin: 0; }
  .muted { color: var(--muted); font-size: 0.8rem; }
  @media (max-width: 600px) {
    .footer-inner { flex-direction: column; gap: 1.5rem; }
    .version { margin-left: 0; text-align: left; }
  }
</style>

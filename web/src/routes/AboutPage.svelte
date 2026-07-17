<script>
  import api from '../lib/api/client';
  let data = null;
  let loading = true;
  let error = null;

  api.get('/api/v1/about').then(d => { data = d; loading = false; }).catch(e => { error = e.message; loading = false; });
</script>

<div class="about-page">
  {#if loading}
    <p class="muted">Loading...</p>
  {:else if error}
    <p class="error">{error}</p>
  {:else if data}
    <div class="hero">
      <h1>{data.instance_name}</h1>
      <p class="subtitle">{data.short_description}</p>
    </div>

    <div class="section">
      <h2>About</h2>
      <p>{data.description}</p>
    </div>

    <div class="section">
      <h2>Version</h2>
      <p><strong>v{data.version}</strong></p>
    </div>

    {#if data.admin_users && data.admin_users.length > 0}
    <div class="section">
      <h2>Administrators</h2>
      <ul>
        {#each data.admin_users as admin}
          <li>{admin}</li>
        {/each}
      </ul>
    </div>
    {/if}

    {#if data.admin_contact_email}
    <div class="section">
      <h2>Contact</h2>
      <a href="mailto:{data.admin_contact_email}">{data.admin_contact_email}</a>
    </div>
    {/if}

    <div class="section links">
      {#if data.privacy_policy_url}
        <a href={data.privacy_policy_url} target="_blank" rel="noopener">Privacy Policy</a>
      {/if}
      {#if data.terms_url}
        <a href={data.terms_url} target="_blank" rel="noopener">Terms of Service</a>
      {/if}
      {#if data.code_of_conduct_url}
        <a href={data.code_of_conduct_url} target="_blank" rel="noopener">Code of Conduct</a>
      {/if}
      {#if data.defederation_policy_url}
        <a href={data.defederation_policy_url} target="_blank" rel="noopener">Defederation Policy</a>
      {/if}
    </div>

    {#if data.donation_url}
    <div class="section donate">
      <h2>Support Us</h2>
      <p>{data.donate_text}</p>
      <a href={data.donation_url} target="_blank" rel="noopener" class="btn">Donate</a>
    </div>
    {/if}
  {/if}
</div>

<style>
  .about-page { max-width: 700px; margin: 0 auto; padding: 2rem 1rem; }
  .hero { text-align: center; margin-bottom: 2rem; }
  .hero h1 { font-size: 2rem; margin: 0; }
  .subtitle { color: var(--muted); font-size: 1.2rem; }
  .section { margin-bottom: 2rem; }
  .section h2 { font-size: 1.2rem; border-bottom: 1px solid var(--border); padding-bottom: 0.3rem; }
  .links { display: flex; gap: 1rem; flex-wrap: wrap; }
  .links a { color: var(--primary); }
  .donate { background: #f0fdf4; padding: 1.5rem; border-radius: var(--radius); border: 1px solid #bbf7d0; }
  .donate h2 { border: none; }
  .btn { display: inline-block; background: var(--primary); color: white; padding: 0.6rem 1.5rem; border-radius: var(--radius); text-decoration: none; margin-top: 0.5rem; }
  .btn:hover { opacity: 0.9; }
  .error { color: var(--danger); }
  .muted { color: var(--muted); }
</style>

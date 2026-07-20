<script>
  import { createEventDispatcher } from 'svelte';
  import { login } from '../lib/stores/auth';
  import api from '../lib/api/client';

  const dispatch = createEventDispatcher();

  let mode = 'login';
  let email = '';
  let password = '';
  let error = '';
  let submitting = false;
  let message = '';

  async function handleLogin() {
    error = '';
    submitting = true;
    try {
      await login(email, password);
      dispatch('login');
    } catch (e) {
      error = e.message || 'Login failed';
    } finally {
      submitting = false;
    }
  }

  async function handleForgotPassword() {
    error = '';
    message = '';
    submitting = true;
    try {
      await api.forgotPassword(email);
      message = 'If that email is registered, a reset link has been sent.';
    } catch (e) {
      error = e.message || 'Something went wrong';
    } finally {
      submitting = false;
    }
  }

  async function handleResendVerification() {
    error = '';
    message = '';
    submitting = true;
    try {
      await api.resendVerification();
      message = 'Verification email sent. Check your inbox.';
    } catch (e) {
      error = e.message || 'Something went wrong';
    } finally {
      submitting = false;
    }
  }
</script>

<div class="auth-page">
  {#if mode === 'login'}
    <form class="auth-card" on:submit|preventDefault={handleLogin}>
      <h1>Welcome back</h1>
      <p class="subtitle">Sign in to Threadlight</p>
      {#if error}
        <div class="error">{error}</div>
      {/if}
      <label>
        Email
        <input type="email" bind:value={email} required placeholder="you@example.com" />
      </label>
      <label>
        Password
        <input type="password" bind:value={password} required placeholder="&#xb7;&#xb7;&#xb7;&#xb7;&#xb7;&#xb7;&#xb7;&#xb7;" />
      </label>
      <button type="submit" disabled={submitting}>
        {submitting ? 'Signing in...' : 'Sign in'}
      </button>
      <p class="switch">
        Don't have an account?
        <button type="button" class="link" on:click={() => dispatch('nav', { page: 'register' })}>Sign up</button>
      </p>
      <p class="switch">
        <button type="button" class="link" on:click={() => { mode = 'forgot'; error = ''; message = ''; }}>Forgot password?</button>
      </p>
      <p class="switch">
        <button type="button" class="link" on:click={() => { mode = 'resend'; error = ''; message = ''; }}>Resend verification email</button>
      </p>
    </form>
  {:else if mode === 'forgot'}
    <form class="auth-card" on:submit|preventDefault={handleForgotPassword}>
      <h1>Reset password</h1>
      <p class="subtitle">Enter your email to receive a reset link</p>
      {#if error}
        <div class="error">{error}</div>
      {/if}
      {#if message}
        <div class="success">{message}</div>
      {/if}
      <label>
        Email
        <input type="email" bind:value={email} required placeholder="you@example.com" />
      </label>
      <button type="submit" disabled={submitting}>
        {submitting ? 'Sending...' : 'Send reset link'}
      </button>
      <p class="switch">
        <button type="button" class="link" on:click={() => { mode = 'login'; error = ''; message = ''; }}>Back to sign in</button>
      </p>
    </form>
  {:else if mode === 'resend'}
    <form class="auth-card" on:submit|preventDefault={handleResendVerification}>
      <h1>Resend verification</h1>
      <p class="subtitle">Enter your email to receive a new verification link</p>
      {#if error}
        <div class="error">{error}</div>
      {/if}
      {#if message}
        <div class="success">{message}</div>
      {/if}
      <label>
        Email
        <input type="email" bind:value={email} required placeholder="you@example.com" />
      </label>
      <button type="submit" disabled={submitting}>
        {submitting ? 'Sending...' : 'Resend'}
      </button>
      <p class="switch">
        <button type="button" class="link" on:click={() => { mode = 'login'; error = ''; message = ''; }}>Back to sign in</button>
      </p>
    </form>
  {/if}
</div>

<style>
  .auth-page {
    display: flex;
    justify-content: center;
    align-items: center;
    min-height: 100svh;
    padding: 1rem;
  }
  .auth-card {
    width: 100%;
    max-width: 400px;
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }
  h1 {
    margin: 0;
  }
  .subtitle {
    color: var(--text);
    margin-bottom: 0.5rem;
  }
  .error {
    padding: 0.5rem;
    background: #fef2f2;
    color: #dc2626;
    border-radius: 6px;
    font-size: 0.875rem;
  }
  .success {
    padding: 0.5rem;
    background: #f0fdf4;
    color: #16a34a;
    border-radius: 6px;
    font-size: 0.875rem;
  }
  label {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
    font-size: 0.875rem;
    color: var(--text-h);
  }
  input {
    padding: 0.5rem 0.75rem;
    border: 1px solid var(--border);
    border-radius: 6px;
    font-size: 1rem;
    background: var(--bg);
    color: var(--text-h);
  }
  input:focus {
    outline: 2px solid var(--accent);
    outline-offset: -1px;
  }
  button[type="submit"] {
    padding: 0.625rem;
    background: var(--accent);
    color: white;
    border: none;
    border-radius: 6px;
    font-size: 1rem;
    cursor: pointer;
    transition: opacity 0.2s;
  }
  button[type="submit"]:disabled {
    opacity: 0.6;
  }
  button[type="submit"]:hover:not(:disabled) {
    opacity: 0.9;
  }
  .switch {
    text-align: center;
    font-size: 0.875rem;
  }
  .link {
    background: none;
    border: none;
    color: var(--accent);
    cursor: pointer;
    font-size: inherit;
    padding: 0;
    text-decoration: underline;
  }
</style>

# FRONTEND PATTERNS (Svelte 4 + Vite SPA)

## API Client (web/src/lib/api/client.js)
```js
class ApiClient {
  constructor() { this.token = null; }
  setToken(t) { this.token = t; localStorage.setItem('polaris_token', t); }
  clearToken() { this.token = null; localStorage.removeItem('polaris_token'); }
  loadToken() { this.token = localStorage.getItem('polaris_token'); return this.token; }

  async request(method, path, body = null) {
    const headers = { 'Content-Type': 'application/json' };
    if (this.token) headers['Authorization'] = `Bearer ${this.token}`;
    const url = path.startsWith('/') ? `http://localhost:8080${path}` : path;
    const res = await fetch(url, { method, headers, body: body ? JSON.stringify(body) : null });
    if (res.status === 204) return null;
    const data = await res.json();
    if (!res.ok) throw new Error(data.error || data.message || 'Request failed');
    return data;
  }

  get(p) { return this.request('GET', p); }
  post(p, b) { return this.request('POST', p, b); }
  put(p, b) { return this.request('PUT', p, b); }
  patch(p, b) { return this.request('PATCH', p, b); }
  del(p) { return this.request('DELETE', p); }

  // Auth
  login(e, p) { return this.post('/api/v1/auth/login', { email: e, password: p }); }
  register(u, e, p, ic) { return this.post('/api/v1/auth/register', { username: u, email: e, password: p, invite_code: ic }); }
  requestPasswordReset(e) { return this.post('/api/v1/auth/password-reset', { email: e }); }
  confirmPasswordReset(t, p) { return this.post('/api/v1/auth/password-reset/confirm', { token: t, password: p }); }

  // Content
  listPosts(l, o) { return this.get(`/api/v1/posts?limit=${l}&offset=${o}`); }
  createPost(d) { return this.post('/api/v1/posts', d); }
  editPost(id, body) { return this.patch(`/api/v1/posts/${id}`, { body }); }
  lockPost(id) { return this.post(`/api/v1/posts/${id}/lock`); }
  hidePost(id) { return this.post(`/api/v1/posts/${id}/hide`); }

  // Communities
  listCommunities(s) { return this.get(`/api/v1/communities?sort=${s}`); }
  getCommunity(s) { return this.get(`/api/v1/communities/${s}`); }
  joinCommunity(s) { return this.post(`/api/v1/communities/${s}/join`); }

  // Circles
  getCircleSuggestions() { return this.get('/api/v1/circles/suggestions'); }
  acceptCircle(i) { return this.post(`/api/v1/circles/${i}/accept`); }

  // Blocking
  blockUser(i) { return this.post(`/api/v1/users/${i}/block`); }
  unblockUser(i) { return this.post(`/api/v1/users/${i}/unblock`); }
  getBlockedUsers() { return this.get('/api/v1/blocks'); }

  // Reports
  createReport(pid, cat, r) { return this.post('/api/v1/reports', { post_id: pid, category: cat, reason: r }); }
  getReports() { return this.get('/api/v1/mod/reports'); }

  // Notifications
  getNotifications(l, o) { return this.get(`/api/v1/notifications?limit=${l}&offset=${o}`); }
  markAllRead() { return this.post('/api/v1/notifications/read-all'); }
  getUnreadCount() { return this.get('/api/v1/notifications/unread-count'); }

  // Profile
  updateProfile(d) { return this.patch('/api/v1/users/me', d); }
  getUserStats(id) { return this.get(`/api/v1/users/${id}/stats`); }

  // Moderation
  flagContent(pid, r) { return this.post('/api/v1/mod/flag', { post_id: pid, reason: r }); }
  getModLog() { return this.get('/api/v1/mod/log'); }
  createModNote(uid, n) { return this.post('/api/v1/mod/notes', { user_id: uid, note: n }); }

  // Admin
  listEmoji() { return this.get('/api/v1/admin/emoji'); }
  createEmoji(s, u, a) { return this.post('/api/v1/admin/emoji', { shortcode: s, image_url: u, alt_text: a }); }

  // Misc
  search(q) { return this.get(`/api/v1/search?q=${encodeURIComponent(q)}`); }
  getFeed(m, l, o) { let p=`limit=${l}&offset=${o}`; if(m&&m!=='all') p+=`&mood=${m}`; return this.get(`/api/v1/feeds?${p}`); }
  claimDaily() { return this.post('/api/v1/credits/claim-daily'); }
  getLeaderboard(t, p) { return this.get(`/api/v1/leaderboard/${t}?period=${p}`); }
}
export const api = new ApiClient();
export default api;
```

## Stores (web/src/lib/stores/)
```js
// auth.js
import { writable } from 'svelte/store';
import api from '../api/client';
export const currentUser = writable(null);
export const isAuthenticated = writable(false);
export const onboardingStage = writable(4);

export async function login(email, password) {
  const resp = await api.login(email, password);
  api.setToken(resp.token);
  currentUser.set(resp.user);
  isAuthenticated.set(true);
  onboardingStage.set(resp.user.onboarding_stage || 4);
  return resp;
}

export async function register(username, email, password, inviteCode) {
  const resp = await api.register(username, email, password, inviteCode);
  api.setToken(resp.token);
  currentUser.set(resp.user);
  isAuthenticated.set(true);
  onboardingStage.set(resp.user.onboarding_stage || 4);
  return resp;
}

export function logout() { api.clearToken(); currentUser.set(null); isAuthenticated.set(false); onboardingStage.set(4); }

export function checkAuth() {
  const token = api.loadToken();
  if (token) { isAuthenticated.set(true); api.session().then(data => currentUser.set(data.user || data)).catch(() => logout()); }
}

// feed.js
export const feedPosts = writable([]);
export const feedLoading = writable(false);
export const feedError = writable(null);
export const currentMood = writable('all');

export async function loadFeed(mood = 'all', limit = 20, offset = 0) {
  feedLoading.set(true); feedError.set(null);
  try { const data = await api.getFeed(mood, limit, offset); feedPosts.set(data.posts || data || []); }
  catch(e) { feedError.set(e.message); } finally { feedLoading.set(false); }
}

// notifications.js
export const notifications = writable([]);
export const unreadCount = writable(0);
export function addNotification(n) { notifications.update(ns => [n,...ns]); unreadCount.update(c=>c+1); }
export function markAllRead() { unreadCount.set(0); }
```

## App.svelte Routing Pattern
```svelte
<script>
  let page = 'feed'; let params = {};
  function goTo(p, p2) { page = p; params = p2 || {}; }
  onMount(() => { checkAuth(); window.addEventListener('navigate', e => { page = e.detail.page; params = e.detail.params || {}; }); });
</script>

{#if !$isAuthenticated}
  {#if page === 'register'}<RegisterPage on:register={()=>goTo('feed')}/>{:else}<LoginPage on:login={()=>goTo('feed')}/>{/if}
{:else if $currentUser && $currentUser.onboarding_stage != null && $currentUser.onboarding_stage < 4}
  <OnboardingWizard on:complete={()=>goTo('feed')}/>
{:else}
  <Navbar {goTo} current={page}/>
  <div class="app-layout"><main>{#if page==='feed'}<FeedPage {goTo}/>
    {:else if page==='search'}<SearchPage {goTo}/>
    {:else if page==='notifications'}<NotificationsPage/>
    {:else if page==='blocked'}<BlockedUsersPage/>
    {:else if page==='profile'}<ProfilePage user={$currentUser} {goTo}/>
    {:else if page==='settings'}<SettingsPage/>
    {:else if page==='mod'}<ModCenter {goTo}/>
    {:else if page==='leaderboard'}<LeaderboardPage/>
    {:else if page==='admin'}<AdminPage/>
    {:else}<FeedPage {goTo}/>
  {/if}</main><aside><TrendingSidebar {goTo}/></aside></div>
{/if}
```

## CSS Variables (in app.css)
```css
:root { --primary: #4f46e5; --bg: #ffffff; --text: #1f2937; --muted: #9ca3af; --border: #e5e7eb; --success: #10b981; --danger: #ef4444; --radius: 8px; }
```

## Key Patterns
- Components use `export let` for props
- `createEventDispatcher` for child→parent events (e.g., `dispatch('login', resp)`)
- Navigation: `window.dispatchEvent(new CustomEvent('navigate', { detail: { page: '...' } }))`
- Vite proxy in vite.config.js: `server: { proxy: { '/api': 'http://localhost:8080' } }`
- All API calls in try/catch, display error via feedError store or local state

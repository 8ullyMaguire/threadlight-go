# Polaris — Frontend Reference

## Tech
- Svelte 4 + Vite (SPA, no SvelteKit)
- Vanilla CSS with CSS custom properties
- Fetch-based API client (no Axios)

## Startup
```bash
cd web && npm install && npm run dev  # → http://localhost:5173
```

## Directory
```
web/src/
├── App.svelte                   # Root: auth guard → onboarding → 14 page routes
├── app.css                      # Global styles + CSS variables
├── main.js                      # Mount point
├── lib/
│   ├── api/
│   │   └── client.js            # ApiClient: ~95 API methods
│   ├── stores/
│   │   ├── auth.js              # currentUser, isAuthenticated, onboardingStage
│   │   ├── feed.js              # feedPosts, feedLoading, currentMood
│   │   └── notifications.js     # notifications, unreadCount
│   └── components/
│       ├── Navbar.svelte        # Top nav: logo, search, links, bell, avatar
│       └── TrendingSidebar.svelte  # Trending topics widget
└── routes/                      # 14 page components
    ├── LoginPage.svelte         # Login + password reset + email verify
    ├── RegisterPage.svelte      # Registration with invite code
    ├── FeedPage.svelte          # Main feed + mood picker + post creation
    ├── SearchPage.svelte        # Full-text search with results
    ├── CommunitiesPage.svelte   # Browse + manage communities
    ├── CollectionsPage.svelte   # Bookmark collections
    ├── ProfilePage.svelte       # User profile + avatar + stats + block
    ├── NotificationsPage.svelte # Notification history
    ├── BlockedUsersPage.svelte  # Blocked user management
    ├── ModCenter.svelte         # Mod log + reports + mod notes
    ├── SettingsPage.svelte      # Theme, filters, account, delete
    ├── LeaderboardPage.svelte   # Multi-period leaderboards
    ├── AdminPage.svelte         # Health, emoji, pages
    └── OnboardingWizard.svelte  # 4-stage setup wizard
```

## Routing
Simple page-variable routing in App.svelte via custom events:
```js
window.dispatchEvent(new CustomEvent('navigate', { detail: { page: 'feed', params: {} } }));
```
Pages: feed, login, register, search, communities, collections, profile, notifications, blocked, mod, settings, leaderboard, admin (+ onboarding when stage < 4).

## API Client Usage
```js
import api from '../lib/api/client';
api.setToken(token);          // from login response
api.loadToken();              // from localStorage
const posts = await api.listPosts(20, 0);
await api.blockUser(userId);
const unread = await api.getUnreadCount();
```

Full API client has methods for all ~95 endpoints. See AI_03_API_REFERENCE.md.

## Stores
```js
import { currentUser, isAuthenticated } from '../lib/stores/auth';
import { feedPosts, feedLoading, feedError, currentMood } from '../lib/stores/feed';
import { notifications, unreadCount } from '../lib/stores/notifications';
```
Reactive access: `{$currentUser?.username}` `{#if $feedLoading}...{/if}`

## CSS Variables
```css
--primary: #4f46e5;  --bg: #ffffff;  --text: #1f2937;
--muted: #9ca3af;    --border: #e5e7eb;
--success: #10b981;  --danger: #ef4444;  --radius: 8px;
```

## Key Patterns
- `export let` for props, `createEventDispatcher` for component events
- `$storeName` auto-subscription in templates
- All API calls in try/catch blocks
- Navigation via `navigate` custom event on window
- Vite proxies `/api` → `http://localhost:8080` (configured in vite.config.js)
- JWT stored in localStorage, managed by auth store

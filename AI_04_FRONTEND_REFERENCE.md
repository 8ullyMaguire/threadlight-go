# Polaris — Frontend Reference

## Tech
- Svelte 5 + Vite (SPA, no SvelteKit)
- Vanilla CSS (no Tailwind/Bootstrap)
- Fetch-based API client (no Axios/ky)

## Startup
```bash
cd web && npm install && npm run dev
# Opens at http://localhost:5173, proxies API to :8080
```

## Directory
```
web/
├── src/
│   ├── App.svelte          # Root: auth guard → onboarding → main app
│   ├── app.css             # Global styles + CSS variables
│   ├── main.js             # Mount point
│   ├── lib/
│   │   ├── api/
│   │   │   └── client.js   # ApiClient class: all endpoint methods
│   │   ├── stores/
│   │   │   ├── auth.js     # currentUser, isAuthenticated, onboardingStage
│   │   │   ├── feed.js     # feedPosts, feedLoading, currentMood
│   │   │   └── notifications.js  # notifications, unreadCount
│   │   └── components/
│   │       ├── Navbar.svelte       # Top nav with links + search + notifications
│   │       ├── TrendingSidebar.svelte  # Trending topics widget
│   │       ├── FeedItem.svelte     # Single post card
│   │       ├── MoodPicker.svelte   # Mood filter pills
│   │       └── CommunityCard.svelte # Community display card
│   └── routes/         # 12 page components (see below)
├── index.html
├── vite.config.js
└── package.json
```

## Routing (App.svelte)
Simple page variable routing (no router library):
```js
// Navigate from any component:
window.dispatchEvent(new CustomEvent('navigate', {
  detail: { page: 'feed', params: {} }
}));
```

Pages mapped in App.svelte:
| Variable | Page Component | Auth Required |
|----------|---------------|---------------|
| 'feed' | FeedPage.svelte | Yes |
| 'login' | LoginPage.svelte | No |
| 'register' | RegisterPage.svelte | No |
| 'search' | SearchPage.svelte | Yes |
| 'communities' | CommunitiesPage.svelte | Yes |
| 'collections' | CollectionsPage.svelte | Yes |
| 'profile' | ProfilePage.svelte | Yes |
| 'mod' | ModCenter.svelte | Yes |
| 'settings' | SettingsPage.svelte | Yes |
| 'leaderboard' | LeaderboardPage.svelte | Yes |
| 'admin' | AdminPage.svelte | Yes |
| (onboarding) | OnboardingWizard.svelte | Yes (stage < 4) |

## API Client Usage
```js
import api from '../lib/api/client';

// Token management (handled by stores)
api.setToken(token);
api.loadToken();   // from localStorage
api.clearToken();

// Calls
const posts = await api.listPosts(20, 0);
const community = await api.getCommunity('photography');
const notes = await api.getPostNotes(postId);
```

## Stores
```js
import { currentUser, isAuthenticated } from '../lib/stores/auth';
import { feedPosts, feedLoading, feedError, currentMood } from '../lib/stores/feed';
import { notifications, unreadCount } from '../lib/stores/notifications';

// Reactive access in templates:
// {$currentUser?.username}
// {#if $feedLoading}Loading...{/if}
```

## CSS Variables (defined in app.css)
```css
--primary: #4f46e5;
--bg: #ffffff;
--text: #1f2937;
--muted: #9ca3af;
--border: #e5e7eb;
--success: #10b981;
--danger: #ef4444;
--radius: 8px;
```

## Key Patterns
- Components use `export let` for props, `createEventDispatcher` for events
- Stores use `$storeName` for auto-subscription
- All API calls wrapped in try/catch with error display
- Navigation via custom events, not URL routing
- Vite proxy in vite.config.js forwards `/api` to the Go server

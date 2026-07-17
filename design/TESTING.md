## Comprehensive Testing Strategy for Social Platform (Go/Gin + Postgres + Svelte)

### 1. Backend Testing (Go & Gin)

#### A. Unit Tests (Business Logic)
- **Tools:** Standard `testing` package, `testify` (for assertions/mocks), `gomock` or `mockery`.
- **What to test:** Service layer functions (e.g., `CreatePost`, `FollowUser`, `GetTimeline`) in isolation.
- **How:** Mock all external dependencies (database repositories, cache, 3rd-party APIs).
- **Go Pattern:** Use interfaces for your repositories. Pass mocks into your services during tests.

```go
// Example: Mocking the DB repo
func TestCreatePost_InvalidContent(t *testing.T) {
    mockRepo := new(MockPostRepository)
    service := NewPostService(mockRepo)
    err := service.CreatePost(ctx, "user1", "") // Empty content
    assert.Error(t, err)
}
```

#### B. Integration Tests (API Endpoints & DB)
- **Tools:** `net/http/httptest` (Gin's built-in test context), `testcontainers-go` (for spinning up real Postgres in Docker), or `pgx` with a dedicated test database.
- **Strategy:** 
  1. Spin up a temporary Postgres container (or use a local test DB).
  2. Run migrations on it.
  3. Seed essential test data (e.g., a test user).
  4. Fire HTTP requests at your Gin router using `httptest.NewRecorder()`.
  5. Assert HTTP status codes, JSON response bodies, and verify the actual DB state changed correctly.

```go
func TestPostComment_Integration(t *testing.T) {
    // Setup test DB container
    db := setupTestDB(t) 
    router := setupRouter(db)
    
    // Create a user and auth token for the test
    token := createTestUserAndLogin(router, t)
    
    // Perform request
    req := httptest.NewRequest("POST", "/api/posts/1/comments", strings.NewReader(`{"text":"Hello"}`))
    req.Header.Set("Authorization", "Bearer "+token)
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    assert.Equal(t, 201, w.Code)
    // Assert the comment actually exists in the test DB
    var count int
    db.QueryRow("SELECT COUNT(*) FROM comments WHERE post_id = 1").Scan(&count)
    assert.Equal(t, 1, count)
}
```

---

### 2. Database Testing (Postgres)

Social platforms are DB-heavy (timelines, feeds, followers). Don't just test CRUD; test **behavior**.

- **Migration Tests:** Test your `up` and `down` migrations in a CI environment to ensure they are idempotent.
- **Query Performance:** Write tests that run `EXPLAIN ANALYZE` on your heaviest queries (e.g., fetching a user's feed with joins). Assert that the query uses the correct indexes (Index Only Scans) and doesn't do Sequential Scans on large tables.
- **Concurrency/Locking Tests:** Test race conditions. For example, run 100 concurrent goroutines trying to "Like" the same post. Use `pgx` batch operations and test for `deadlock detected` errors to ensure your transaction isolation levels (`READ COMMITTED` / `SERIALIZABLE`) are correctly handling contention.

---

### 3. Frontend Testing (Svelte)

Svelte testing focuses on components and user interactions.

#### A. Unit / Component Tests
- **Tools:** `Vitest` (for unit logic) + `@testing-library/svelte` (for component testing).
- **What to test:** 
  - Pure functions (date formatting, string truncation).
  - Component rendering (does the "Post" component render the avatar and text?).
  - Reactive declarations (does `$: filteredPosts = posts.filter(...)` work correctly when input changes).

```javascript
import { render, screen } from '@testing-library/svelte';
import PostCard from './PostCard.svelte';

test('displays post content', () => {
  render(PostCard, { props: { content: 'Hello World', username: 'John' } });
  expect(screen.getByText('Hello World')).toBeInTheDocument();
});
```

#### B. Store/State Testing
- Social apps rely heavily on stores (writable/readable).
- Test that dispatching actions (e.g., liking a post) correctly updates the store state *before* the API call optimistically updates the UI.

---

### 4. End-to-End (E2E) Testing (Critical for Social UX)

This simulates real user flows across the entire stack.

- **Tool:** **Playwright** (preferred) or Cypress.
- **Critical Flows to test:**
  1. **Auth Flow:** Sign up -> Email verification -> Login -> Logout.
  2. **Social Graph:** User A follows User B -> Check if A sees B's posts in their feed.
  3. **Real-time/WebSocket:** If you have chat/notifications, test socket connections.
  4. **File Upload:** Profile picture upload and compression.
  5. **Infinite Scroll:** Scroll to the bottom of the feed and verify new posts load via the API.

```javascript
// Playwright Example
test('User can like a post', async ({ page }) => {
  await page.goto('/');
  await page.click('text=Login');
  // ... fill credentials
  const likeButton = page.locator('[data-testid="like-button"]').first();
  await expect(likeButton).toHaveText('0');
  await likeButton.click();
  await expect(likeButton).toHaveText('1');
});
```
- **Visual Regression:** Use Playwright's `screenshot` feature to catch accidental CSS breaks across different screen sizes.

---

### 5. Performance & Load Testing (Crucial for Social)

Social platforms have viral spikes. You must test under load.

- **Tool:** **k6** (open-source).
- **Scenario:** Simulate 1,000 concurrent users performing a "Feed Refresh" every 5 seconds.
- **Metrics to track:** P95 latency, Error rate, Rate limiting behavior.
- **Where to test:** Hit your **read-heavy** endpoints (GET `/api/feed`) and **write-heavy** endpoints (POST `/api/posts`).
- **Goal:** Ensure your Gin middleware (timeouts, graceful shutdown) and Postgres connection pool handle the surge without crashing.

---

### 6. Security Testing

- **SQL Injection:** Run a DAST (Dynamic Analysis) tool like `ZAP` against your staging environment. Since you use GORM or `pgx`, ensure raw SQL queries are parameterized.
- **JWT/Authorization:** Write integration tests specifically for Gin middleware. Try accessing `/api/admin` with a standard user token – it must return `403 Forbidden`.
- **Rate Limiting:** Write a test that fires 101 requests to the login endpoint in 1 minute and asserts the 100th request returns `429 Too Many Requests`.

---

### 7. The CI/CD Pipeline (Forgejo-Specific)

Automate all of the above in your pipeline. Forgejo Actions is compatible with GitHub Actions syntax. Place this in `.forgejo/workflows/ci.yml`.

#### Forgejo-Specific Setup:
1. **Install and register the runner:** `forgejo-runner register` and `forgejo-runner daemon` on your server.
2. **Store secrets** in your Forgejo repo: Settings → Secrets (e.g., `DOCKER_PASSWORD`, `DEPLOY_SSH_KEY`).
3. **Forgejo Container Registry** is built-in at `your-forgejo-domain.com/owner/repo` - use it instead of Docker Hub.

```yaml
name: Full Stack CI/CD

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  # ---------- JOB 1: Backend Tests (Go + Postgres) ----------
  backend-test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16-alpine
        env:
          POSTGRES_USER: testuser
          POSTGRES_PASSWORD: testpass
          POSTGRES_DB: social_test
        ports:
          - 5432:5432
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'
          cache: true

      - name: Cache Go modules
        uses: actions/cache@v4
        with:
          path: ~/go/pkg/mod
          key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}

      - name: Run Go Unit + Integration Tests
        env:
          TEST_DB_HOST: localhost
          TEST_DB_PORT: 5432
          TEST_DB_USER: testuser
          TEST_DB_PASS: testpass
          TEST_DB_NAME: social_test
        run: |
          go test -v -race -coverprofile=coverage.out ./...
          go tool cover -html=coverage.out -o coverage.html

      - name: Upload coverage report
        uses: actions/upload-artifact@v4
        with:
          name: backend-coverage
          path: coverage.html

  # ---------- JOB 2: Frontend Tests (Svelte) ----------
  frontend-test:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'
          cache-dependency-path: frontend/package-lock.json

      - name: Cache Node modules
        uses: actions/cache@v4
        with:
          path: frontend/node_modules
          key: ${{ runner.os }}-node-${{ hashFiles('frontend/package-lock.json') }}

      - name: Install dependencies
        run: npm ci
        working-directory: ./frontend

      - name: Run Svelte Unit/Component tests
        run: npm run test:unit
        working-directory: ./frontend

      - name: Build Svelte app
        run: npm run build
        working-directory: ./frontend

  # ---------- JOB 3: Full E2E (Playwright) ----------
  e2e-test:
    runs-on: ubuntu-latest
    needs: [backend-test, frontend-test]
    services:
      postgres:
        image: postgres:16-alpine
        env:
          POSTGRES_USER: testuser
          POSTGRES_PASSWORD: testpass
          POSTGRES_DB: social_e2e
        ports:
          - 5432:5432
        options: --health-cmd pg_isready --health-interval 10s --health-timeout 5s --health-retries 5

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Set up Node
        uses: actions/setup-node@v4
        with:
          node-version: '20'

      - name: Build Go backend
        run: |
          go build -o server ./cmd/api
          chmod +x server

      - name: Build Svelte frontend
        run: |
          cd frontend
          npm ci
          npm run build
          cd ..

      - name: Start Backend Server
        env:
          DB_HOST: localhost
          DB_PORT: 5432
          DB_USER: testuser
          DB_PASS: testpass
          DB_NAME: social_e2e
          JWT_SECRET: testsecret
          STATIC_DIR: ./frontend/dist
        run: |
          ./server &
          sleep 5

      - name: Install Playwright
        run: |
          cd frontend
          npx playwright install --with-deps

      - name: Run Playwright E2E tests
        run: |
          cd frontend
          npm run test:e2e
        env:
          BASE_URL: http://localhost:8080

      - name: Upload Playwright report on failure
        uses: actions/upload-artifact@v4
        if: failure()
        with:
          name: playwright-report
          path: frontend/test-results/

  # ---------- JOB 4: Deploy to Production (Forgejo) ----------
  deploy:
    runs-on: ubuntu-latest
    needs: [e2e-test]
    if: github.ref == 'refs/heads/main'
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      # Forgejo has a built-in container registry
      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Log in to Forgejo Container Registry
        uses: docker/login-action@v3
        with:
          registry: your-forgejo-domain.com  # Replace with your Forgejo domain
          username: ${{ secrets.DOCKER_USERNAME }}
          password: ${{ secrets.DOCKER_PASSWORD }}

      - name: Build and push Backend Docker image
        uses: docker/build-push-action@v5
        with:
          context: .
          file: ./Dockerfile.backend
          push: true
          tags: your-forgejo-domain.com/your-org/your-repo/backend:latest

      - name: Build and push Frontend Docker image
        uses: docker/build-push-action@v5
        with:
          context: ./frontend
          file: ./frontend/Dockerfile
          push: true
          tags: your-forgejo-domain.com/your-org/your-repo/frontend:latest

      # Alternative: Deploy via SSH (non-Docker)
      - name: Deploy via SSH
        if: false  # Set to true if using SSH instead of Docker
        uses: appleboy/ssh-action@v1
        with:
          host: ${{ secrets.PRODUCTION_HOST }}
          username: ${{ secrets.SSH_USER }}
          key: ${{ secrets.SSH_PRIVATE_KEY }}
          script: |
            cd /var/www/social-app
            git pull origin main
            go build -o server ./cmd/api
            sudo systemctl restart social-backend
            cd frontend && npm ci && npm run build && sudo systemctl restart nginx
```

#### Forgejo-Specific Notes for the Pipeline:

| Aspect | Forgejo Implementation |
| :--- | :--- |
| **Runner Setup** | Install `forgejo-runner` on your server. Register with: `forgejo-runner register` and run with `forgejo-runner daemon`. Use the `docker` executor for isolation. |
| **Secrets Management** | Store secrets (DB passwords, JWT tokens, SSH keys) in repo Settings → Secrets. Reference with `${{ secrets.NAME }}`. |
| **Container Registry** | Forgejo includes a built-in OCI registry. Authenticate with your Forgejo username and a token/password. Images are stored at `your-forgejo-domain.com/owner/repo/image:tag`. |
| **Cache Storage** | Forgejo runner supports `actions/cache@v4`. Caches are stored locally on the runner server by default. |
| **Debugging** | Set `ACTIONS_RUNNER_DEBUG: true` in repo variables for verbose logs. |
| **Resource Limits** | In the runner config (`config.yml`), set `mem_limit` and `cpu_shares` to prevent CI from overloading your self-hosted server. |

---

### Summary Checklist for Your Stack (With Forgejo Updates)

| Layer | Tool | Target |
| :--- | :--- | :--- |
| **Go Logic** | `testify` + Mocks | Service functions, validators, utils |
| **Gin API** | `httptest` + Testcontainers | HTTP handlers, middleware, routing |
| **Postgres** | `go-test` + `EXPLAIN` | Migrations, Index usage, Transaction locks |
| **Svelte Components** | Vitest + Testing Library | UI state, reactivity, DOM rendering |
| **Full Stack** | Playwright | Critical user journeys (Auth, Feed, Interactions) |
| **Performance** | k6 | `/feed` and `/post` endpoints under high concurrency |
| **Security** | OWASP ZAP / custom tests | AuthZ, Rate limiting, CORS, SQL injection |
| **CI/CD** | Forgejo Actions (`.forgejo/workflows/ci.yml`) | Automate tests + deploy to Forgejo Container Registry or via SSH |

---

**Pro Tip for Forgejo:** Since you're self-hosting, be careful with parallel test jobs that spin up Postgres containers. If your runner has limited RAM, set `max-parallel: 2` in your job configuration or use `testcontainers` with resource limits to avoid OOM errors.

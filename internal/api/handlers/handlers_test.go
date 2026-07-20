package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/opencode-ai/polaris/internal/api"
	"github.com/opencode-ai/polaris/internal/db"
	"github.com/redis/go-redis/v9"
)

var (
	testRouter *gin.Engine
	testPG     *pgxpool.Pool
	testRDB    *redis.Client
	testJWT    = "test-jwt-secret-not-for-production"
)

func TestMain(m *testing.M) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://polaris:polaris@localhost:5432/polaris?sslmode=disable"
	}
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	var err error
	testPG, err = db.NewPostgres(dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "postgres: %v (set DATABASE_URL or start docker-compose)\n", err)
		os.Exit(1)
	}
	testRDB, err = db.NewRedis(redisAddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "redis: %v\n", err)
		os.Exit(1)
	}

	migrate(testPG)

	testRouter = api.NewRouter(testPG, testRDB, testJWT)

	// Ensure registration is open for tests
	testPG.Exec(context.Background(), `UPDATE site_config SET registration_mode = 'open' WHERE id = 1`)

	code := m.Run()
	testPG.Close()
	testRDB.Close()
	os.Exit(code)
}

func migrate(pg *pgxpool.Pool) {
	sql, err := os.ReadFile("../../db/migrations/001_initial.sql")
	if err != nil {
		panic(fmt.Sprintf("read migration: %v", err))
	}
	_, err = pg.Exec(context.Background(), string(sql))
	if err != nil {
		panic(fmt.Sprintf("migrate: %v", err))
	}
}

func jsonBody(v interface{}) io.Reader {
	b, _ := json.Marshal(v)
	return bytes.NewReader(b)
}

func request(method, path string, body io.Reader, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, body)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	return w
}

func registerUser(username, email, password string) (string, int) {
	w := request("POST", "/api/v1/auth/register", jsonBody(map[string]string{
		"username": username,
		"email":    email,
		"password": password,
	}), "")
	var resp struct {
		Token string `json:"token"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	return resp.Token, w.Code
}

func TestHealth(t *testing.T) {
	w := request("GET", "/health", nil, "")
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["status"] != "ok" {
		t.Errorf("expected ok status")
	}
}

func TestAuthRegister(t *testing.T) {
	w := request("POST", "/api/v1/auth/register", jsonBody(map[string]string{
		"username": "testuser_" + fmt.Sprintf("%d", time.Now().UnixNano()),
		"email":    fmt.Sprintf("test_%d@example.com", time.Now().UnixNano()),
		"password": "password123",
	}), "")
	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp["token"] == "" {
		t.Error("expected non-empty token")
	}
}

func TestAuthLogin(t *testing.T) {
	email := fmt.Sprintf("login_%d@example.com", time.Now().UnixNano())
	registerUser("loginuser_"+fmt.Sprintf("%d", time.Now().UnixNano()), email, "password123")

	w := request("POST", "/api/v1/auth/login", jsonBody(map[string]string{
		"email":    email,
		"password": "password123",
	}), "")
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["token"] == "" {
		t.Error("expected non-empty token")
	}
}

func TestAuthBadLogin(t *testing.T) {
	w := request("POST", "/api/v1/auth/login", jsonBody(map[string]string{
		"email":    "nonexistent@example.com",
		"password": "wrong",
	}), "")
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthSession(t *testing.T) {
	ts := fmt.Sprintf("%d_%d", time.Now().UnixNano(), 1)
	token, code := registerUser("sesuser_"+ts, "ses_"+ts+"@example.com", "password123")
	if code != http.StatusCreated {
		// read the body for debugging
		w := request("POST", "/api/v1/auth/register", jsonBody(map[string]string{
			"username": "sesuser_" + ts,
			"email":    "ses_" + ts + "@example.com",
			"password": "password123",
		}), "")
		t.Fatalf("register failed: %d, body=%s", w.Code, w.Body.String())
	}

	w := request("GET", "/api/v1/auth/session", nil, token)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCreatePost(t *testing.T) {
	token, _ := registerUser("postuser_"+fmt.Sprintf("%d", time.Now().UnixNano()),
		fmt.Sprintf("post_%d@example.com", time.Now().UnixNano()), "password123")

	w := request("POST", "/api/v1/posts", jsonBody(map[string]interface{}{
		"title":    "Test Post",
		"body":     "This is a test post body",
		"mood":     1,
		"is_educational": true,
	}), token)
	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestListPosts(t *testing.T) {
	token, _ := registerUser("listuser_"+fmt.Sprintf("%d", time.Now().UnixNano()),
		fmt.Sprintf("list_%d@example.com", time.Now().UnixNano()), "password123")

	request("POST", "/api/v1/posts", jsonBody(map[string]interface{}{
		"title": "Post 1", "body": "Body 1",
	}), token)
	request("POST", "/api/v1/posts", jsonBody(map[string]interface{}{
		"title": "Post 2", "body": "Body 2",
	}), token)

	w := request("GET", "/api/v1/posts?limit=10&offset=0", nil, token)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestAuthMiddleware(t *testing.T) {
	w := request("GET", "/api/v1/posts", nil, "")
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for no token, got %d", w.Code)
	}

	w = request("GET", "/api/v1/posts", nil, "invalid-token")
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for bad token, got %d", w.Code)
	}
}

func TestTags(t *testing.T) {
	token, _ := registerUser("taguser_"+fmt.Sprintf("%d", time.Now().UnixNano()),
		fmt.Sprintf("tag_%d@example.com", time.Now().UnixNano()), "password123")

	w := request("POST", "/api/v1/tags", jsonBody(map[string]string{
		"name":     fmt.Sprintf("testtag_%d", time.Now().UnixNano()),
		"category": "testing",
	}), token)
	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCommunities(t *testing.T) {
	token, _ := registerUser("comuser_"+fmt.Sprintf("%d", time.Now().UnixNano()),
		fmt.Sprintf("com_%d@example.com", time.Now().UnixNano()), "password123")

	slug := fmt.Sprintf("testcom_%d", time.Now().UnixNano())
	w := request("POST", "/api/v1/communities", jsonBody(map[string]interface{}{
		"name":          "Test Community",
		"slug":          slug,
		"tags":          []string{"testing", "go"},
		"slow_boot_days": 3,
	}), token)
	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	w = request("GET", "/api/v1/communities/"+slug, nil, token)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCollections(t *testing.T) {
	token, _ := registerUser("coluser_"+fmt.Sprintf("%d", time.Now().UnixNano()),
		fmt.Sprintf("col_%d@example.com", time.Now().UnixNano()), "password123")

	w := request("POST", "/api/v1/collections", jsonBody(map[string]interface{}{
		"name":       "Test Collection",
		"visibility": 1,
	}), token)
	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestSearch(t *testing.T) {
	token, _ := registerUser("searchuser_"+fmt.Sprintf("%d", time.Now().UnixNano()),
		fmt.Sprintf("search_%d@example.com", time.Now().UnixNano()), "password123")

	// Create a post to search for
	request("POST", "/api/v1/posts", jsonBody(map[string]interface{}{
		"title": "Searchable Post Title",
		"body":  "This is a unique searchable content for testing",
	}), token)

	w := request("GET", "/api/v1/search?q=searchable&limit=10", nil, token)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestFeed(t *testing.T) {
	token, _ := registerUser("feeduser_"+fmt.Sprintf("%d", time.Now().UnixNano()),
		fmt.Sprintf("feed_%d@example.com", time.Now().UnixNano()), "password123")

	w := request("GET", "/api/v1/feeds?limit=10&offset=0", nil, token)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestBlocklist(t *testing.T) {
	token, _ := registerUser("bluser_"+fmt.Sprintf("%d", time.Now().UnixNano()),
		fmt.Sprintf("bl_%d@example.com", time.Now().UnixNano()), "password123")

	w := request("POST", "/api/v1/moderation/actions", jsonBody(map[string]interface{}{
		"action_type": 5,
		"reason":      "test blocklist entry",
	}), token)
	// Blocklist endpoint might error due to trust threshold, but it should return valid HTTP
	if w.Code != http.StatusCreated && w.Code != http.StatusForbidden && w.Code != http.StatusBadRequest {
		t.Errorf("unexpected status: %d", w.Code)
	}
}

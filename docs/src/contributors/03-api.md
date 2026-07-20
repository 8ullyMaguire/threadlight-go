# API Reference

The Threadlight API is a RESTful JSON API. All endpoints are prefixed with `/api/v1/`.

## Authentication

Most endpoints require authentication. To authenticate, include a JWT token in the `Authorization` header:

```
Authorization: Bearer <your-token>
```

To get a token, use the login endpoint.

## Endpoints

### Authentication

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/auth/register` | Create a new account |
| POST | `/api/v1/auth/login` | Log in and get a token |
| POST | `/api/v1/auth/refresh` | Refresh an expiring token |

### Users

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/users/me` | Get the current user's profile |
| GET | `/api/v1/users/:id` | Get a user's public profile |
| PUT | `/api/v1/users/me` | Update your profile |
| POST | `/api/v1/users/me/trust/:id` | Trust or untrust a user |

### Posts

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/posts` | List posts (feed) |
| POST | `/api/v1/posts` | Create a post |
| GET | `/api/v1/posts/:id` | Get a single post |
| PUT | `/api/v1/posts/:id` | Update a post |

### Communities

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/communities` | List communities |
| POST | `/api/v1/communities` | Create a community |
| GET | `/api/v1/communities/:id` | Get community details |
| POST | `/api/v1/communities/:id/join` | Join a community |

### Community Notes

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/posts/:id/notes` | Add a community note |
| GET | `/api/v1/posts/:id/notes` | Get notes for a post |

### Moderation

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/mod/report` | Report a post or user |
| GET | `/api/v1/mod/jury` | Get current jury cases |
| POST | `/api/v1/mod/jury/:id/vote` | Vote on a jury case |

### About / Instance Info

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/about` | Get server information, stats, and philosophy |

### Tags

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/tags` | List trending tags |
| GET | `/api/v1/tags/:tag` | Get posts for a tag |

### Feeds

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/feeds` | List your custom feeds |
| POST | `/api/v1/feeds` | Create a custom feed |
| GET | `/api/v1/feeds/:id` | Get posts from a custom feed |

## Response Format

All responses are JSON. Successful responses look like:

```json
{
  "data": { ... },
  "success": true
}
```

Error responses look like:

```json
{
  "error": "Description of what went wrong",
  "success": false
}
```

## Pagination

List endpoints support pagination with `page` and `limit` query parameters:

```
GET /api/v1/posts?page=1&limit=20
```

Responses include pagination metadata:

```json
{
  "data": [ ... ],
  "page": 1,
  "limit": 20,
  "total": 150,
  "success": true
}
```

## Rate Limiting

API requests are rate-limited per user. The limits are:

- 60 requests per minute for authenticated users
- 10 requests per minute for unauthenticated users

Exceeding the limit returns a `429 Too Many Requests` response.

# WASM Plugin API

This page describes how to write a WASM plugin for Threadlight's feed marketplace. Plugins let developers add custom ranking and filtering logic to user feeds.

## Overview

Threadlight supports three types of plugins that run inside a WebAssembly sandbox (powered by [wazero](https://wazero.io/)):

- **Rank plugins** — score posts so the best ones appear higher in the feed
- **Filter plugins** — remove or hide posts that match certain criteria
- **Hybrid plugins** — do both filtering and ranking in one pass

Plugins are compiled to WASI (WebAssembly System Interface) and uploaded to the server. They have no access to the network, filesystem, or any system resources outside the sandbox. The only data they receive is the post input you give them.

## Plugin Types

### Rank Plugins (type 0)

Rank plugins assign a score to each post. The feed system sorts posts by this score (higher = more likely to appear).

Use a rank plugin when you want to surface the best content, like something that boosts posts from people you trust or penalizes posts with certain patterns.

The plugin receives an array of post objects and must return an array of scored results in the same order.

### Filter Plugins (type 1)

Filter plugins decide which posts to keep and which to hide. They return a subset of the posts they received.

Use a filter plugin when you want to hide content, like something that removes posts containing certain keywords or from accounts that match a pattern.

### Hybrid Plugins (type 2)

Hybrid plugins do both. They filter out some posts and score the remaining ones. The feed system first applies filtering, then ranks the survivors.

Use a hybrid plugin when you need both behaviors in a single pass (for example, remove short posts and score the rest by sentiment).

## Exported Functions

Your WASM module must export one or more of the following functions depending on the plugin type.

### `rank_posts`

```
func rank_posts(input: []byte) -> []byte
```

Required for: rank and hybrid plugins.

**Input:** JSON array of post objects. Each post object contains:

```json
{
  "id": 12345,
  "author_id": 6789,
  "content": "Post text here...",
  "created_at": "2026-07-20T12:00:00Z",
  "tags": ["gardening", "cooking"],
  "reaction_counts": {
    "insightful": 5,
    "funny": 2,
    "misleading": 0
  },
  "author_trust_level": 2,
  "community_id": 42
}
```

**Output:** JSON array of scored post objects:

```json
[
  {"post_id": 12345, "score": 0.85},
  {"post_id": 12346, "score": 0.42}
]
```

Scores should be floats between 0.0 and 1.0. The feed system sorts by score descending and interleaves with the default algorithm.

### `filter_post`

```
func filter_post(input: []byte) -> []byte
```

Required for: filter and hybrid plugins.

**Input:** Same JSON array of post objects as `rank_posts`.

**Output:** JSON array of post IDs to keep:

```json
[12345, 12347, 12349]
```

Posts whose IDs are not in the returned array are removed from the feed. The plugin decides which posts to include based on its own logic.

### `plugin_metadata`

```
func plugin_metadata() -> []byte
```

Optional but recommended for all plugin types.

**Output:** JSON object with metadata:

```json
{
  "name": "My Cool Plugin",
  "version": "1.0.0",
  "description": "Ranks posts by gardening relevance",
  "plugin_type": "rank",
  "author": "your-username"
}
```

This metadata is used for display purposes in the plugin marketplace.

## Input and Output Format

All data passes through WASM linear memory. The host (Threadlight) writes the input JSON to memory, calls the exported function, and reads the output JSON from memory.

The exact memory layout follows the WASI convention:

1. The host allocates memory and writes the input JSON string.
2. The host passes two `i32` arguments: pointer and length of the input.
3. The plugin reads the input, processes it, and writes the output JSON to memory.
4. The plugin returns two `i32` values: pointer and length of the output.

If you are using the TinyGo or Rust WASI SDK, this is handled for you by the standard I/O wrappers.

### Input Field Reference

| Field | Type | Description |
|-------|------|-------------|
| `id` | int64 | Post ID |
| `author_id` | int64 | ID of the user who created the post |
| `content` | string | The post text (Markdown) |
| `created_at` | string | ISO 8601 timestamp |
| `tags` | array of strings | Tags applied to the post |
| `reaction_counts` | object | Map of reaction type to count |
| `author_trust_level` | int | Trust level of the author (0–4) |
| `community_id` | int64 | Community the post belongs to, or 0 |

## Sandbox Restrictions

Plugins run in a strict sandbox with the following limitations:

- **No network access.** You cannot make HTTP requests, connect to databases, or reach any external service.
- **No filesystem access.** You cannot read or write files on the server.
- **No system calls.** You cannot spawn processes, read environment variables, or access system time beyond basic WASI clocks.
- **Memory limit.** The sandbox is limited to 1 MB of linear memory (configurable by the server admin).
- **Execution timeout.** Each plugin call is limited to 10 milliseconds of CPU time. If your plugin takes longer, it is killed and the feed falls back to the default algorithm for that call.

These restrictions are enforced by the wazero runtime with a custom host environment that provides no host functions beyond what WASI requires for basic I/O.

## How to Write a Plugin

### Using Go (with TinyGo)

TinyGo compiles Go code to WASI. Here is a minimal rank plugin:

```go
package main

import (
	"encoding/json"
	"strings"
)

type Post struct {
	ID      int64  `json:"id"`
	Content string `json:"content"`
	// Include other fields as needed
}

type ScoredPost struct {
	PostID int64   `json:"post_id"`
	Score  float64 `json:"score"`
}

func main() {}

//export rank_posts
func RankPosts(inputPtr, inputLen uint32) (resultPtr, resultLen uint32) {
	input := readMemory(inputPtr, inputLen)

	var posts []Post
	json.Unmarshal(input, &posts)

	var results []ScoredPost
	for _, p := range posts {
		score := 0.5
		if strings.Contains(p.Content, "gardening") {
			score = 0.9
		}
		results = append(results, ScoredPost{PostID: p.ID, Score: score})
	}

	output, _ := json.Marshal(results)
	return writeMemory(output)
}
```

Compile with:

```bash
tinygo build -o plugin.wasm -target=wasi -no-debug .
```

### Using Rust

```rust
use serde::{Deserialize, Serialize};

#[derive(Deserialize)]
struct Post {
    id: i64,
    content: String,
}

#[derive(Serialize)]
struct ScoredPost {
    post_id: i64,
    score: f64,
}

#[no_mangle]
pub extern "C" fn rank_posts(ptr: i32, len: i32) -> (i32, i32) {
    let input = unsafe {
        std::slice::from_raw_parts(ptr as *const u8, len as usize)
    };
    let posts: Vec<Post> = serde_json::from_slice(input).unwrap();
    let results: Vec<ScoredPost> = posts.iter().map(|p| {
        ScoredPost {
            post_id: p.id,
            score: if p.content.contains("gardening") { 0.9 } else { 0.5 },
        }
    }).collect();
    let output = serde_json::to_vec(&results).unwrap();
    (output.as_ptr() as i32, output.len() as i32)
}
```

Compile with:

```bash
cargo build --target wasm32-wasi --release
```

The compiled `.wasm` file is what you will upload to the marketplace.

### Required Metadata Fields

When you upload a plugin, you must provide the following fields:

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Plugin name, max 100 characters |
| `description` | No | Short description of what the plugin does |
| `version` | No | Semantic version string, defaults to 1.0.0 |
| `plugin_type` | Yes | 0 = rank, 1 = filter, 2 = hybrid |
| `price_credits` | No | Cost in credits (0 = free) |
| `wasm` | Yes | The compiled `.wasm` binary file (max 10 MB) |

## Upload and Review Process

1. **Compile your plugin** to a `.wasm` file using one of the SDK languages above.
2. **Upload** the `.wasm` file through the feed plugins section in the UI or via the API.
3. **Submit for review.** New plugins are marked as `reviewed = false` by default.
4. **Admin review.** A server administrator will review your plugin. They can see the WASM SHA-256 hash and the metadata. Since plugins have no network or filesystem access, the primary concern is correctness and performance.
5. **Publishing.** Once the admin marks it as reviewed, the plugin appears in the marketplace for other users to install.

## Installation and Usage

After a plugin is published, users can:

- **Install** the plugin from the marketplace (one click)
- **Configure** it with a JSON config blob (optional, passed to the plugin at runtime)
- **Enable or disable** it from their installed plugins list
- **Rate and review** it (1–5 stars, with an optional text review)

Users can install multiple plugins. The feed system runs all installed rank plugins and averages their scores, then applies all filter plugins in sequence.

## Performance Considerations

- Keep your plugin fast. The 10ms timeout is per-call, and a single feed render may invoke it multiple times.
- Avoid complex regex or heavy string processing in the hot path.
- If your plugin needs to do heavy analysis, consider precomputing data and embedding it in the WASM binary at compile time.
- Test your plugin locally before uploading. You can use the API's execute endpoint to test with sample post IDs.

## Security

The WASM sandbox is designed to be safe even for untrusted plugins:

- No network access means a plugin cannot exfiltrate data.
- No filesystem access means it cannot read server files or write malware.
- Memory limits prevent resource exhaustion.
- Timeout prevents infinite loops.
- The SHA-256 hash of every uploaded binary is stored, so the admin can verify integrity.

If you find a way to escape the sandbox, please report it to the server administrator immediately.

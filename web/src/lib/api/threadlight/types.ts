export interface User {
  id: number
  username: string
  display_name: string | null
  email: string | null
  avatar_url: string | null
  banner_url: string | null
  bio: string | null
  role: number
  trust_level: number
  credits: number
  post_count: number
  comment_count: number
  like_count: number
  is_private: boolean
  is_deleted: boolean
  created_at: string
  updated_at: string | null
}

export interface Post {
  id: number
  author_id: number
  title: string
  body: string
  content_type: number
  mood: number
  is_educational: boolean
  is_entertaining: boolean
  is_nsfw: boolean
  content_warning: string | null
  interaction_count: number
  cumulative_interactions: number
  status: number
  scheduled_at: string | null
  created_at: string | null
  updated_at: string | null
  archived_at: string | null
  edited_at: string | null
  locked: boolean
  sticky: boolean
  sticky_at: string | null
  language: string
  is_ai_generated: boolean
  license: string
  cross_post_root_id: number | null
  moved_from_community_id: number | null
  is_deleted: boolean
}

export interface PostResponse {
  post: Post
  message: string
}

export interface CreatePostRequest {
  title: string
  body: string
  content_type: number
  mood: number
  is_educational: boolean
  is_entertaining: boolean
  is_nsfw: boolean
  content_warning: string
}

export interface UpdatePostRequest {
  title: string
  body: string
  content_type: number
  mood: number
  is_educational: boolean
  is_entertaining: boolean
  is_nsfw: boolean
  content_warning: string
}

export interface Community {
  id: number
  name: string
  slug: string
  description: string
  icon_url: string | null
  banner_url: string | null
  primary_tag_id: number | null
  owner_id: number
  member_count: number
  post_count: number
  is_private: boolean
  is_moderated: boolean
  slow_boot_days: number
  created_at: string
  updated_at: string | null
}

export interface CommunityMember {
  id: number
  community_id: number
  user_id: number
  role: number
  status: number
  joined_at: string
}

export interface CreateCommunityRequest {
  name: string
  description: string
  icon_url: string
  banner_url: string
  is_private: boolean
  primary_tag_id: number
}

export interface Tag {
  id: number
  name: string
  description: string
  category: string
  is_wiki: boolean
  created_by: number
  created_at: string
  total_votes?: number
  vote_score?: number
  user_vote?: number
}

export interface PostTag {
  post_id: number
  tag_id: number
  tagged_by: number
  created_at: string
}

export interface TagVote {
  id: number
  tag_id: number
  user_id: number
  vote: number
  created_at: string
}

export interface TagVoteRequest {
  vote: number
}

export interface Notification {
  id: number
  user_id: number
  notification_type: number
  actor_id: number | null
  post_id: number | null
  body: string
  is_read: boolean
  created_at: string
}

export interface UserInvite {
  id: number
  inviter_id: number
  code: string
  used_by: number | null
  used_at: string | null
  created_at: string
}

export interface PostReportItem {
  id: number
  post_id: number
  reporter_id: number
  category: number
  reason: string
  status: number
  resolved_by: number | null
  created_at: string
  resolved_at: string | null
}

export interface BlockedUser {
  id: number
  blocker_id: number
  blocked_id: number
  created_at: string
}

export interface Interaction {
  id: number
  user_id: number
  post_id: number
  interaction_type: number
  metadata: Record<string, unknown> | null
  created_at: string
}

export interface FeedSource {
  id: number
  feed_id: number
  community_id: number | null
  tag_id: number | null
  user_id: number | null
  created_at: string
}

export interface CustomFeed {
  id: number
  user_id: number
  name: string
  description: string
  icon_url: string | null
  created_at: string
  updated_at: string | null
}

export interface Collection {
  id: number
  name: string
  description: string | null
  user_id: number
  post_count: number
  is_public: boolean
  created_at: string
  updated_at: string | null
}

export interface Circle {
  id: number
  name: string
  description: string | null
  owner_id: number
  member_count: number
  created_at: string
  updated_at: string | null
}

export interface Block {
  id: number
  user_id: number
  blocked_user_id: number
  created_at: string
}

export interface Report {
  id: number
  reporter_id: number
  target_type: string
  target_id: number
  reason: string
  status: number
  created_at: string
}

export interface CreateReportRequest {
  target_type: string
  target_id: number
  reason: string
}

export interface SiteConfig {
  title: string
  description: string
  logo_url: string | null
  invite_only: boolean
  registration_open: boolean
  max_upload_size: number
  default_language: string
  federation_enabled: boolean
  theme: string
}

export interface AboutResponse {
  title: string
  description: string
  version: string
  source_url: string
  registration_open: boolean
  invite_only: boolean
  max_upload_size: number
  default_language: string
  post_count: number
  user_count: number
  community_count: number
}

export interface AuthResponse {
  user: User
  token: string
  message: string
}

export interface SearchResults {
  posts: Post[]
  users: User[]
  communities: Community[]
  tags: Tag[]
}

// ── New Search API types ─────────────────────────────────────

export interface SearchPostsParams {
  q: string
  author?: string
  community?: string
  tags?: string
  date_from?: string
  date_to?: string
  mood?: string
  content_type?: string
  is_educational?: string
  is_nsfw?: string
  sort?: string
  page?: number
  limit?: number
}

export interface SearchUsersParams {
  q: string
  page?: number
  limit?: number
}

export interface SearchCommunitiesParams {
  q: string
  page?: number
  limit?: number
}

export interface SearchResultAuthor {
  id: number
  name: string
  display_name?: string | null
  avatar_url?: string | null
}

export interface SearchResultCommunity {
  id: number
  name: string
  slug: string
  icon_url?: string | null
}

export interface SearchResultTag {
  id: number
  name: string
}

export interface SearchResultItem {
  id: number
  type: 'post' | 'user' | 'community'
  /** Title (post title, user/community name) */
  title?: string
  /** Body / description snippet */
  body?: string
  /** Highlighted snippet with <mark> tags */
  snippet?: string
  author?: SearchResultAuthor
  community?: SearchResultCommunity
  tags?: SearchResultTag[]
  score?: number
  comment_count?: number
  created_at?: string
  avatar_url?: string | null
  banner_url?: string | null
  member_count?: number
  post_count?: number
  is_nsfw?: boolean
  is_educational?: boolean
  mood?: number
  content_type?: number
}

export interface SearchFacets {
  authors?: { name: string; count: number }[]
  communities?: { name: string; count: number }[]
  tags?: { name: string; count: number }[]
  moods?: { name: string; count: number }[]
}

export interface SearchResponse {
  results: SearchResultItem[]
  total: number
  page: number
  limit: number
  facets?: SearchFacets
}

export interface SearchSuggestResponse {
  suggestions: string[]
}

export interface RegisterRequest {
  username: string
  email: string
  password: string
  invite_code?: string
}

export interface LoginRequest {
  email: string
  password: string
}

export interface ForgotPasswordRequest {
  email: string
}

export interface ResetPasswordRequest {
  token: string
  password: string
}

export interface UpdateSettingsRequest {
  display_name?: string
  bio?: string
  avatar_url?: string
  banner_url?: string
  is_private?: boolean
  theme?: string
}

export interface UserCredits {
  credits: number
  lifetime_credits: number
}

export interface UserBlocklist {
  blocked_users: User[]
}

export interface UnreadCountResponse {
  count: number
}

export interface UserProfile {
  user: User
  post_count: number
  comment_count: number
  is_following: boolean
  is_blocked: boolean
}

export interface HealthResponse {
  status: string
  uptime: string
  version: string
}

export interface ErrorResponse {
  error: string
}

export interface MessageResponse {
  message: string
}

// ── User List types ──────────────────────────────────────────
export interface UserList {
  id: number
  name: string
  description: string
  list_type: 'follow' | 'block' | 'algorithmic'
  visibility: 'private' | 'public'
  owner_id: number
  created_at: string
  updated_at: string | null
}

export interface CreateListRequest {
  name: string
  description: string
  list_type: 'follow' | 'block'
  visibility: 'private' | 'public'
}

export interface CreateAlgorithmicListRequest {
  name: string
  description: string
  list_type: 'follow' | 'block'
  visibility: 'private' | 'public'
  scope: string
  criteria: string
}

export interface AddListMemberRequest {
  target_user_id: number
}

export interface AddCollaboratorRequest {
  user_id: number
  role: string
}

export interface UserListMember {
  id: number
  list_id: number
  user_id: number
  created_at: string
}

export interface UserListCollaborator {
  id: number
  list_id: number
  user_id: number
  role: string
  created_at: string
}

// ── Affinity types ──────────────────────────────────────────
export interface AffinityEntry {
  user_b_id: number
  affinity_score: number
  breakdown: string
}

export interface AffinitySimilarResponse {
  user_ids: number[]
}

// ── Feed Plugin types ──────────────────────────────────────
export interface FeedPlugin {
  id: number
  name: string
  description: string
  version: string
  plugin_type: string
  wasm_url: string
  rating: number
  install_count: number
  reviewed: boolean
  user_id: number
  created_at: string
  updated_at: string | null
}

export interface FeedPluginReview {
  id: number
  plugin_id: number
  user_id: number
  rating: number
  review: string
  created_at: string
}

export interface CreatePluginReviewRequest {
  rating: number
  review: string
}

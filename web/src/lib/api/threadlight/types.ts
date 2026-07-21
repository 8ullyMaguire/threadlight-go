export interface User {
  id: number
  username: string
  display_name: string | null
  email: string | null
  avatar_url: string | null
  banner_url: string | null
  bio: string | null
  role?: number
  trust_level: number
  trust_score?: number
  reputation?: number
  credits: number
  post_count?: number
  comment_count?: number
  like_count?: number
  is_private?: boolean
  is_deleted: boolean
  is_active?: boolean
  is_local?: boolean
  onboarding_stage?: number
  email_verified?: boolean
  theme?: string | null
  hide_read_posts?: boolean
  created_at: string
  updated_at?: string | null
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
  tags?: string[]
  icon_url?: string | null
  banner_url?: string | null
  primary_tag_id?: number | null
  owner_id?: number
  created_by?: number
  member_count: number
  post_count?: number
  is_private?: boolean
  invite_only?: boolean
  is_moderated?: boolean
  curator_lock?: boolean
  slow_boot_days: number
  min_trust_score?: number
  forked_from?: number | null
  created_at: string
  updated_at?: string | null
  archived_at?: string | null
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
  instance_name: string
  short_description: string
  description: string
  version: string
  admin_contact_email: string
  privacy_policy_url: string
  terms_url: string
  code_of_conduct_url: string
  defederation_policy_url: string
  donation_url: string
  donate_text: string
  admin_users: string[]
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
  tags?: Record<string, number>
  mood?: Record<number, number>
  content_type?: Record<number, number>
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
  list_type: number | 'follow' | 'block' | 'algorithmic'
  visibility: number | 'private' | 'public'
  owner_id: number
  is_algorithmic?: boolean
  criteria_json?: string | null
  scope?: number
  tag_id?: number | null
  refresh_interval?: string
  last_refreshed_at?: string | null
  created_at: string
  updated_at?: string | null
}

export interface UserListWithMeta extends UserList {
  is_subscribed: boolean
  member_count: number
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
  tag_id?: number | null
  refresh?: string
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
  user_a_id?: number
  user_b_id: number
  affinity_score: number
  recency_factor?: number
  breakdown?: string | null
  computed_at?: string
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
  plugin_type: number | string
  author_id?: number
  wasm_url?: string
  wasm_sha256?: string
  price_credits?: number
  rating: number
  install_count: number
  enabled?: boolean
  reviewed: boolean
  user_id?: number
  created_at: string
  updated_at?: string | null
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

// ── Modlog / Moderation types (Threadlight-native) ──────────

export interface ModAction {
  id: number
  moderator_id: number
  action_type: number
  reason: string
  target_post_id: number | null
  target_comment_id: number | null
  target_user_id: number | null
  target_community_id: number | null
  created_at: string
}

export interface ControversialDecision {
  id: number
  action_id: number
  action: ModAction
  reviews_fair: number
  reviews_unfair: number
  reviews_total: number
  controversy_score: number
  moderator?: User
  community?: Community
  target_user?: User
  target_post?: Post
  created_at: string
}

export interface ControversialDecisionsResponse {
  decisions: ControversialDecision[]
  total: number
  limit: number
  offset: number
}

export interface ReviewVote {
  id: number
  user_id: number
  action_id: number
  vote: number // 1 = fair, -1 = unfair
  trust_penalty: number | null
  created_at: string
}

export interface CastVoteResponse {
  review: ReviewVote
  trust_penalty?: number
  message: string
}

export interface ReviewListResponse {
  reviews: ReviewVote[]
  total: number
}

export interface ModerationActionResponse {
  actions: ModAction[]
  total: number
}

export interface ResolveReportResponse {
  message: string
}

// ── Credit Economy types ─────────────────────────────────────

export interface DailyQuest {
  quest_type: string
  description: string
  completed: boolean
  progress: number
  target: number
  reward: number
}

export interface CreditBalance {
  credits: number
  lifetime_credits: number
  streak: number
  quests: DailyQuest[]
}

export interface CreditTransaction {
  id: number
  user_id: number
  amount: number
  balance_after: number
  reason: string
  reference_type: string | null
  reference_id: number | null
  created_at: string
}

export interface CreditTransactionsResponse {
  transactions: CreditTransaction[]
  total: number
  page: number
  limit: number
}

export interface CreditCosts {
  post_creation: number
  image_upload: number
  search: number
  private_message: number
  reaction: number
}

export interface CommunityBalance {
  community_slug: string
  balance: number
  total_earned: number
  total_spent: number
}

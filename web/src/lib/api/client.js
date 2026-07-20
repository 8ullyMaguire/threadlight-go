const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080';

class ApiClient {
  constructor() {
    this.token = null;
  }

  setToken(token) {
    this.token = token;
    localStorage.setItem('polaris_token', token);
  }

  clearToken() {
    this.token = null;
    localStorage.removeItem('polaris_token');
  }

  loadToken() {
    this.token = localStorage.getItem('polaris_token');
    return this.token;
  }

  async request(method, path, body = null) {
    const headers = { 'Content-Type': 'application/json' };
    if (this.token) {
      headers['Authorization'] = `Bearer ${this.token}`;
    }
    const url = path.startsWith('http') ? path : `${API_BASE}${path}`;
    const opts = { method, headers };
    if (body) opts.body = JSON.stringify(body);
    const res = await fetch(url, opts);
    if (res.status === 204) return null;
    const data = await res.json();
    if (!res.ok) throw new ApiError(res.status, data.error || data.message || 'Request failed');
    return data;
  }

  get(path) { return this.request('GET', path); }
  post(path, body) { return this.request('POST', path, body); }
  put(path, body) { return this.request('PUT', path, body); }
  patch(path, body) { return this.request('PATCH', path, body); }
  del(path) { return this.request('DELETE', path); }

  // Auth
  register(username, email, password, inviteCode) {
    return this.post('/api/v1/auth/register', { username, email, password, invite_code: inviteCode });
  }
  login(email, password) { return this.post('/api/v1/auth/login', { email, password }); }
  logout() { return this.post('/api/v1/auth/logout'); }
  session() { return this.get('/api/v1/auth/session'); }
  forgotPassword(email) { return this.post('/api/v1/auth/forgot-password', { email }); }
  resetPassword(token, password) { return this.post('/api/v1/auth/reset-password', { token, password }); }
  resendVerification() { return this.post('/api/v1/auth/resend-verification'); }
  verifyEmail(token) { return this.post('/api/v1/auth/verify-email', { token }); }

  // Posts
  createPost(data) { return this.post('/api/v1/posts', data); }
  getPost(id) { return this.get(`/api/v1/posts/${id}`); }
  listPosts(limit = 20, offset = 0) { return this.get(`/api/v1/posts?limit=${limit}&offset=${offset}`); }
  interact(postId, type, metadata) { return this.post(`/api/v1/posts/${postId}/interact`, { interaction_type: type, metadata }); }
  deletePost(id) { return this.del(`/api/v1/posts/${id}`); }
  hidePost(id) { return this.post(`/api/v1/posts/${id}/hide`); }
  unhidePost(id) { return this.post(`/api/v1/posts/${id}/unhide`); }
  savePost(id) { return this.post(`/api/v1/posts/${id}/save`); }
  unsavePost(id) { return this.post(`/api/v1/posts/${id}/unsave`); }
  reportPost(id, reason) { return this.post(`/api/v1/posts/${id}/report`, { reason }); }

  // Tags
  createTag(name, category) { return this.post('/api/v1/tags', { name, category }); }
  searchTags(q) { return this.get(`/api/v1/tags/search?q=${encodeURIComponent(q)}`); }

  // Feed
  getFeed(mood, limit, offset) {
    let params = `limit=${limit}&offset=${offset}`;
    if (mood && mood !== 'all') params += `&mood=${mood}`;
    return this.get(`/api/v1/feeds?${params}`);
  }

  // Communities
  createCommunity(data) { return this.post('/api/v1/communities', data); }
  listCommunities(sort = 'active_weekly') { return this.get(`/api/v1/communities?sort=${sort}`); }
  getCommunity(slug) { return this.get(`/api/v1/communities/${slug}`); }
  joinCommunity(slug) { return this.post(`/api/v1/communities/${slug}/join`); }
  leaveCommunity(slug) { return this.post(`/api/v1/communities/${slug}/leave`); }

  // Circles
  getCircleSuggestions() { return this.get('/api/v1/circles/suggestions'); }
  acceptCircle(id) { return this.post(`/api/v1/circles/${id}/accept`); }
  dismissCircle(id) { return this.post(`/api/v1/circles/${id}/dismiss`); }

  // Collections
  createCollection(data) { return this.post('/api/v1/collections', data); }
  listCollections() { return this.get('/api/v1/collections'); }
  addToCollection(collectionId, postId) {
    return this.post(`/api/v1/collections/${collectionId}/posts`, { post_id: postId });
  }
  removeFromCollection(collectionId, postId) {
    return this.del(`/api/v1/collections/${collectionId}/posts/${postId}`);
  }

  // Notes
  createNote(postId, body) { return this.post('/api/v1/notes', { post_id: postId, body }); }
  voteOnNote(noteId, helpful) { return this.post(`/api/v1/notes/${noteId}/vote`, { helpful }); }
  getPostNotes(postId) { return this.get(`/api/v1/notes/post/${postId}`); }

  // Search
  search(q, limit = 20) { return this.get(`/api/v1/search?q=${encodeURIComponent(q)}&limit=${limit}`); }

  // Moderation
  flagContent(postId, reason) { return this.post('/api/v1/mod/flag', { post_id: postId, reason }); }
  getModLog() { return this.get('/api/v1/mod/log'); }
  getReports() { return this.get('/api/v1/mod/reports'); }
  resolveReport(id, action) { return this.post(`/api/v1/mod/reports/${id}/resolve`, { action }); }
  createModNote(userId, body) { return this.post('/api/v1/mod/notes', { user_id: userId, body }); }
  getModNotes(userId) { return this.get(`/api/v1/mod/notes/${userId}`); }

  // Credits
  claimDaily() { return this.post('/api/v1/credits/claim-daily'); }
  getBalance() { return this.get('/api/v1/credits/balance'); }
  getTransactionHistory() { return this.get('/api/v1/credits/transactions'); }

  // Leaderboard
  getLeaderboard(type, period = 'weekly') {
    return this.get(`/api/v1/leaderboard/${type}?period=${period}`);
  }

  // Onboarding
  getOnboardingProgress() { return this.get('/api/v1/onboarding/progress'); }
  pickOnboardingTags(tagIds) { return this.post('/api/v1/onboarding/tags', { tag_ids: tagIds }); }
  getSuggestedUsers() { return this.get('/api/v1/onboarding/suggested-users'); }
  advanceOnboarding(stage) { return this.post('/api/v1/onboarding/stage', { stage }); }

  // Filters
  listFilters() { return this.get('/api/v1/filters'); }
  createFilter(data) { return this.post('/api/v1/filters', data); }
  deleteFilter(id) { return this.del(`/api/v1/filters/${id}`); }

  // Achievements
  listAchievements() { return this.get('/api/v1/achievements'); }

  // Config + Invites
  getConfig() { return this.get('/api/v1/admin/config'); }
  updateConfig(data) { return this.put('/api/v1/admin/config', data); }
  generateInvite() { return this.post('/api/v1/invites', {}); }
  listInvites() { return this.get('/api/v1/invites'); }
  getInviteLimit() { return this.get('/api/v1/invites/limit'); }

  // Users / Social
  getUser(username) { return this.get(`/api/v1/users/${encodeURIComponent(username)}`); }
  getUserPosts(username, limit = 20, offset = 0) { return this.get(`/api/v1/users/${encodeURIComponent(username)}/posts?limit=${limit}&offset=${offset}`); }
  getFollowers(username) { return this.get(`/api/v1/users/${encodeURIComponent(username)}/followers`); }
  getFollowing(username) { return this.get(`/api/v1/users/${encodeURIComponent(username)}/following`); }
  followUser(username) { return this.post(`/api/v1/users/${encodeURIComponent(username)}/follow`); }
  unfollowUser(username) { return this.post(`/api/v1/users/${encodeURIComponent(username)}/unfollow`); }
  blockUser(username) { return this.post(`/api/v1/users/${encodeURIComponent(username)}/block`); }
  unblockUser(username) { return this.post(`/api/v1/users/${encodeURIComponent(username)}/unblock`); }
  getBlockedUsers() { return this.get('/api/v1/users/blocked'); }
  updateProfile(data) { return this.patch('/api/v1/users/me', data); }

  // Notifications
  listNotifications() { return this.get('/api/v1/notifications'); }
  markNotificationRead(id) { return this.post(`/api/v1/notifications/${id}/read`); }
  markAllNotificationsRead() { return this.post('/api/v1/notifications/read-all'); }

  // Admin
  adminListUsers() { return this.get('/api/v1/admin/users'); }
  adminSuspendUser(id, reason) { return this.post(`/api/v1/admin/users/${id}/suspend`, { reason }); }
  adminBanUser(id, reason) { return this.post(`/api/v1/admin/users/${id}/ban`, { reason }); }
  adminWarnUser(id, reason) { return this.post(`/api/v1/admin/users/${id}/warn`, { reason }); }
  adminDeletePost(id) { return this.del(`/api/v1/admin/posts/${id}`); }

  // Referrals
  getReferralCode() { return this.get('/api/v1/referrals/code'); }
  getReferrals() { return this.get('/api/v1/referrals'); }

  // Health
  health() { return this.get('/health'); }

  // File uploads
  uploadAvatar(file) { return this.upload('/api/v1/users/me/avatar', file); }
  uploadBanner(file) { return this.upload('/api/v1/users/me/banner', file); }

  async upload(path, file) {
    const form = new FormData();
    form.append('file', file);
    const headers = {};
    if (this.token) headers['Authorization'] = `Bearer ${this.token}`;
    const url = `${API_BASE}${path}`;
    const res = await fetch(url, { method: 'POST', headers, body: form });
    const data = await res.json();
    if (!res.ok) throw new ApiError(res.status, data.error || 'Upload failed');
    return data;
  }
}

class ApiError extends Error {
  constructor(status, message) {
    super(message);
    this.status = status;
  }
}

export const api = new ApiClient();
export default api;

import { BaseClient } from '$lib/api/base'
import type * as Lemmy from 'lemmy-js-client'
import { instance } from '$lib/app/instance.svelte'
import { instanceToURL } from '$lib/app/util.svelte'
import type { MyUserInfo } from '$lib/api/types'
import {
  type AboutResponse,
  type UserProfile,
  type PostResponse,
  type Community,
  type Tag,
  type TagVote,
  type TagVoteRequest,
  type SearchResults,
  type User,
  type Post,
  type SearchResponse,
  type SearchPostsParams,
  type SearchUsersParams,
  type SearchCommunitiesParams,
  type SearchSuggestResponse,
} from './types'

function mapUser(u: User): Lemmy.PersonView {
  return {
    person: {
      id: u.id,
      name: u.display_name ?? u.username,
      display_name: u.display_name ?? undefined,
      avatar: u.avatar_url ?? undefined,
      banner: u.banner_url ?? undefined,
      bio: u.bio ?? undefined,
      published: u.created_at,
      updated: u.updated_at ?? undefined,
      actor_id: `${instanceToURL(instance.data)}/u/${u.username}`,
      local: true,
      deleted: u.is_deleted,
      matrix_user_id: undefined,
      admin: u.role >= 100,
      bot_account: false,
      ban_expires: undefined,
      banned: false,
      instance_id: 1,
    },
    counts: {
      id: u.id,
      person_id: u.id,
      post_count: u.post_count,
      comment_count: u.comment_count,
      post_score: u.like_count,
      comment_score: 0,
    },
  } as Lemmy.PersonView
}

function mapPost(p: Post): Lemmy.PostView {
  return {
    post: {
      id: p.id,
      name: p.title,
      body: p.body ?? undefined,
      url: undefined,
      embed_title: undefined,
      embed_description: undefined,
      embed_video_url: undefined,
      thumbnail_url: undefined,
      nsfw: p.is_nsfw,
      published: p.created_at ?? new Date().toISOString(),
      updated: p.updated_at ?? undefined,
      deleted: p.is_deleted,
      locked: p.locked,
      stickied: p.sticky,
      language_id: 1,
      featured_community: false,
      featured_local: false,
      community_id: 1,
      creator_id: p.author_id,
      ap_id: '',
      local: true,
      embed_html: undefined,
    },
    creator_banned_from_community: false,
    counts: {
      id: p.id,
      post_id: p.id,
      comments: 0,
      score: p.interaction_count,
      upvotes: p.cumulative_interactions,
      downvotes: 0,
      featured_community: false,
      featured_local: false,
      hot_rank: 0,
      hot_rank_active: 0,
      newest_comment_time: p.created_at ?? undefined,
      newest_comment_time_necro: p.created_at ?? undefined,
      published: p.created_at ?? new Date().toISOString(),
      read_comments: 0,
      score_time: 0,
      upvotes_time: 0,
      downvotes_time: 0,
    },
    creator: undefined as unknown as Lemmy.Person,
    community: undefined as unknown as Lemmy.Community,
  } as Lemmy.PostView
}

function mapCommunity(c: Community): Lemmy.CommunityView {
  return {
    community: {
      id: c.id,
      name: c.slug,
      title: c.name,
      description: c.description ?? undefined,
      icon: c.icon_url ?? undefined,
      banner: c.banner_url ?? undefined,
      removed: false,
      published: c.created_at,
      updated: c.updated_at ?? undefined,
      deleted: false,
      nsfw: c.is_private,
      actor_id: '',
      local: true,
      hidden: false,
      posting_restricted_to_mods: false,
      instance_id: 1,
      only_followers_can_vote: false,
    },
    counts: {
      id: c.id,
      community_id: c.id,
      subscribers: c.member_count,
      posts: c.post_count,
      comments: 0,
      published: c.created_at,
      users_active_day: 0,
      users_active_half_year: 0,
      users_active_month: 0,
      users_active_week: 0,
      hot_rank: 0,
      subscribers_local: 0,
    },
    subscribed: 'NotSubscribed' as Lemmy.SubscribedType,
    blocked: false,
  } as Lemmy.CommunityView
}

function mapNotification(n: Notification): Lemmy.Notification {
  return {
    id: n.id,
    person_id: n.user_id,
    published: n.created_at,
    read: n.is_read,
    updated: undefined,
    type: n.type as Lemmy.NotificationType,
    title: n.title,
    subtitle: undefined,
    body: n.body,
    link: n.link ?? undefined,
  } as Lemmy.Notification
}

export class ThreadlightClient extends BaseClient {
  #baseUrl: string = ''

  constructor(
    baseUrl: string,
    args: {
      fetchFunction?: (
        input: RequestInfo | URL,
        init?: RequestInit | undefined,
      ) => Promise<Response>
      headers?: Record<string, string>
    },
  ) {
    super()
    this.#baseUrl = baseUrl
  }

  async #api<T>(
    path: string,
    options: {
      method?: string
      body?: unknown
      params?: Record<string, string>
      allowNonOk?: boolean
    } = {},
  ): Promise<T> {
    let url = `${this.#baseUrl.replace(/\/$/, '')}/api/v1${path}`

    if (options.params) {
      const searchParams = new URLSearchParams()
      for (const [key, value] of Object.entries(options.params)) {
        if (value !== undefined && value !== '') {
          searchParams.set(key, value)
        }
      }
      const qs = searchParams.toString()
      if (qs) url += `?${qs}`
    }

    const init: RequestInit = {
      method: options.method ?? 'GET',
      credentials: 'include',
      headers: {
        'Content-Type': 'application/json',
        Accept: 'application/json',
      },
    }

    if (options.body && options.method !== 'GET') {
      init.body = JSON.stringify(options.body)
    }

    const res = await fetch(url, init)
    if (!res.ok && !options.allowNonOk) {
      const errBody = await res.text().catch(() => '')
      throw new Error(errBody || `HTTP ${res.status}`)
    }
    const contentType = res.headers.get('content-type') ?? ''
    if (contentType.includes('application/json')) {
      return (await res.json()) as T
    }
    return { message: await res.text() } as T
  }

  async getSite(): Promise<Lemmy.GetSiteResponse> {
    let about: AboutResponse | undefined
    try {
      about = await this.#api<AboutResponse>('/about')
    } catch {
      // Backend not reachable; will use fallback
    }

    if (!about) {
      return {
        site_view: {
          site: {
            id: 0,
            name: 'Threadlight',
            sidebar: '',
            published: new Date().toISOString(),
            updated: undefined,
            icon: undefined,
            banner: undefined,
            actor_id: '',
            last_refreshed_at: new Date().toISOString(),
            inbox_url: '',
            public_key: '',
            private_key: undefined,
            instance_id: 1,
            description: '',
            sidebar_html: undefined,
          },
          counts: {
            comments: 0,
            communities: 0,
            posts: 0,
            users: 0,
            users_active_day: 0,
            users_active_half_year: 0,
            users_active_month: 0,
            users_active_week: 0,
            id: 0,
            site_id: 0,
          },
        },
        admins: [],
        all_languages: [],
        custom_emojis: [],
        discussion_languages: [],
        taglines: [],
        version: '0.19.4',
        my_user: undefined,
      } as Lemmy.GetSiteResponse
    }

    let myUser: MyUserInfo | undefined
    try {
      const sessionData = await this.#api<{ user: User }>('/auth/session')
      if (sessionData?.user) {
        const userView = mapUser(sessionData.user)
        myUser = {
          local_user_view: userView,
          follows: [],
          moderates: [],
          community_blocks: [],
          person_blocks: [],
          discussion_languages: [1],
        }
      }
    } catch {
      // Not logged in
    }

    return {
      site_view: {
        site: {
          id: 1,
          name: about.title,
          sidebar: about.description,
          published: new Date().toISOString(),
          updated: undefined,
          icon: undefined,
          banner: undefined,
          actor_id: '',
          last_refreshed_at: new Date().toISOString(),
          inbox_url: '',
          public_key: '',
          private_key: undefined,
          instance_id: 1,
          description: about.description,
          sidebar_html: undefined,
        },
        counts: {
          comments: 0,
          communities: about.community_count,
          posts: about.post_count,
          users: about.user_count,
          users_active_day: 0,
          users_active_half_year: 0,
          users_active_month: 0,
          users_active_week: 0,
          id: 1,
          site_id: 1,
        },
      },
      admins: [],
      all_languages: [],
      custom_emojis: [],
      discussion_languages: [],
      taglines: [],
      version: about.version || '0.19.4',
      my_user: myUser,
    } as Lemmy.GetSiteResponse
  }

  async getLoggedInUser(): Promise<{ user: User } | undefined> {
    try {
      return await this.#api<{ user: User }>('/auth/session')
    } catch {
      return undefined
    }
  }

  async login(form: Lemmy.Login): Promise<Lemmy.LoginResponse> {
    const res = await this.#api<{ user: User; message: string } | { error: string }>(
      '/auth/login',
      {
        method: 'POST',
        body: { email: form.username_or_email, password: form.password },
      },
    )
    if ('error' in res) throw new Error(res.error)
    return {
      jwt: 'session',
      registration_created: false,
      verify_email_sent: false,
    } as Lemmy.LoginResponse
  }

  async register(form: Lemmy.Register): Promise<Lemmy.LoginResponse> {
    const body: Record<string, unknown> = {
      username: form.username,
      email: form.email ?? '',
      password: form.password,
    }
    if (form.show_nsfw !== undefined) body['is_nsfw'] = form.show_nsfw
    const res = await this.#api<{ user: User; message: string } | { error: string }>(
      '/auth/register',
      {
        method: 'POST',
        body,
      },
    )
    if ('error' in res) throw new Error(res.error)
    return {
      jwt: 'session',
      registration_created: false,
      verify_email_sent: false,
    } as Lemmy.LoginResponse
  }

  async saveUserSettings(
    form: Lemmy.SaveUserSettings,
  ): Promise<Lemmy.LoginResponse> {
    await this.#api<{ message: string }>('/users/settings', {
      method: 'PUT',
      body: form,
    })
    return { jwt: '' } as Lemmy.LoginResponse
  }

  async getPosts(form: Lemmy.GetPosts): Promise<Lemmy.GetPostsResponse> {
    const params: Record<string, string> = {}
    if (form.limit) params['limit'] = form.limit.toString()
    if (form.page) params['page'] = form.page.toString()
    if (form.community_id) params['community_id'] = form.community_id.toString()
    if (form.sort) {
      if (form.sort === 'Hot') params['sort'] = 'hot'
      else if (form.sort === 'New') params['sort'] = 'new'
      else if (form.sort === 'Top') params['sort'] = 'top'
      else if (form.sort === 'TopDay') params['sort'] = 'top_day'
    }

    const path = form.community_id ? '/posts' : '/feed'
    const data = await this.#api<Post[] | { posts: Post[] }>(path, { params })
    const posts = Array.isArray(data) ? data : data.posts ?? []
    return {
      posts: posts.map(mapPost),
      next_page: undefined,
    } as Lemmy.GetPostsResponse
  }

  async getPost(form: Lemmy.GetPost): Promise<Lemmy.GetPostResponse> {
    const data = await this.#api<Post>(`/posts/${form.id}`)
    return {
      post_view: mapPost(data),
      community_view: undefined as unknown as Lemmy.CommunityView,
      moderators: [],
      cross_posts: [],
    } as Lemmy.GetPostResponse
  }

  async createPost(form: Lemmy.CreatePost): Promise<Lemmy.PostResponse> {
    const data = await this.#api<PostResponse>('/posts', {
      method: 'POST',
      body: {
        title: form.name,
        body: form.body ?? '',
        is_nsfw: form.nsfw ?? false,
      },
    })
    return {
      post_view: mapPost(data.post),
    } as Lemmy.PostResponse
  }

  async editPost(form: Lemmy.EditPost): Promise<Lemmy.PostResponse> {
    const data = await this.#api<PostResponse>(`/posts/${form.post_id}`, {
      method: 'PUT',
      body: {
        title: form.name,
        body: form.body ?? '',
        is_nsfw: form.nsfw ?? false,
      },
    })
    return { post_view: mapPost(data.post) } as Lemmy.PostResponse
  }

  async deletePost(form: Lemmy.DeletePost): Promise<Lemmy.PostResponse> {
    await this.#api(`/posts/${form.post_id}`, { method: 'DELETE' })
    return {} as Lemmy.PostResponse
  }

  async likePost(form: Lemmy.LikePost): Promise<Lemmy.PostResponse> {
    const data = await this.#api<PostResponse>(`/posts/${form.post_id}/like`, {
      method: 'POST',
      body: { score: form.score },
    })
    return { post_view: mapPost(data.post) } as Lemmy.PostResponse
  }

  async savePost(form: Lemmy.SavePost): Promise<Lemmy.PostResponse> {
    const data = await this.#api<PostResponse>(`/posts/${form.post_id}/bookmark`, {
      method: 'POST',
      body: { save: form.save },
    })
    return { post_view: mapPost(data.post) } as Lemmy.PostResponse
  }

  async getCommunities(
    form: Lemmy.ListCommunities,
  ): Promise<Lemmy.ListCommunitiesResponse> {
    const params: Record<string, string> = {}
    if (form.limit) params['limit'] = form.limit.toString()
    if (form.page) params['page'] = form.page.toString()
    const data = await this.#api<Community[]>('/communities', { params })
    return {
      communities: (data ?? []).map(mapCommunity),
    } as Lemmy.ListCommunitiesResponse
  }

  async getCommunity(
    form: Lemmy.GetCommunity,
  ): Promise<Lemmy.GetCommunityResponse> {
    const data = await this.#api<Community>(`/communities/${form.id}`)
    return {
      community_view: mapCommunity(data),
      moderators: [],
      online: 0,
      discussion_languages: [],
    } as Lemmy.GetCommunityResponse
  }

  async followCommunity(
    form: Lemmy.FollowCommunity,
  ): Promise<Lemmy.CommunityResponse> {
    const endpoint = form.follow ? 'join' : 'leave'
    await this.#api(`/communities/${form.community_id}/${endpoint}`, {
      method: 'POST',
    })
    return {} as Lemmy.CommunityResponse
  }

  async getPerson(form: Lemmy.GetPersonDetails): Promise<Lemmy.GetPersonDetailsResponse> {
    const data = await this.#api<UserProfile>(`/users/${form.person_id}`)
    return {
      person_view: mapUser(data.user),
      comments: [],
      posts: [],
      moderates: [],
    } as Lemmy.GetPersonDetailsResponse
  }

  async getPersonDetails(
    form: Lemmy.GetPersonDetails,
  ): Promise<Lemmy.GetPersonDetailsResponse> {
    return this.getPerson(form)
  }

  async search(form: Lemmy.Search): Promise<Lemmy.SearchResponse> {
    const params: Record<string, string> = { q: form.q }
    if (form.limit) params['limit'] = form.limit.toString()
    if (form.page) params['page'] = form.page.toString()
    const data = await this.#api<SearchResults>('/search', { params })
    return {
      posts: (data.posts ?? []).map(mapPost),
      comments: [],
      communities: (data.communities ?? []).map(mapCommunity),
      users: (data.users ?? []).map(mapUser),
      type_: form.type_,
    } as Lemmy.SearchResponse
  }

  // ── New search endpoints ──────────────────────────────────

  async searchPosts(params: SearchPostsParams): Promise<SearchResponse> {
    const apiParams: Record<string, string> = { q: params.q }
    if (params.author) apiParams['author'] = params.author
    if (params.community) apiParams['community'] = params.community
    if (params.tags) apiParams['tags'] = params.tags
    if (params.date_from) apiParams['date_from'] = params.date_from
    if (params.date_to) apiParams['date_to'] = params.date_to
    if (params.mood) apiParams['mood'] = params.mood
    if (params.content_type) apiParams['content_type'] = params.content_type
    if (params.is_educational !== undefined) apiParams['is_educational'] = params.is_educational
    if (params.is_nsfw !== undefined) apiParams['is_nsfw'] = params.is_nsfw
    if (params.sort) apiParams['sort'] = params.sort
    if (params.page) apiParams['page'] = params.page.toString()
    if (params.limit) apiParams['limit'] = params.limit.toString()
    return this.#api<SearchResponse>('/search/posts', { params: apiParams })
  }

  async searchUsers(params: SearchUsersParams): Promise<SearchResponse> {
    const apiParams: Record<string, string> = { q: params.q }
    if (params.page) apiParams['page'] = params.page.toString()
    if (params.limit) apiParams['limit'] = params.limit.toString()
    return this.#api<SearchResponse>('/search/users', { params: apiParams })
  }

  async searchCommunities(params: SearchCommunitiesParams): Promise<SearchResponse> {
    const apiParams: Record<string, string> = { q: params.q }
    if (params.page) apiParams['page'] = params.page.toString()
    if (params.limit) apiParams['limit'] = params.limit.toString()
    return this.#api<SearchResponse>('/search/communities', { params: apiParams })
  }

  async searchSuggest(q: string): Promise<SearchSuggestResponse> {
    return this.#api<SearchSuggestResponse>('/search/suggest', {
      params: { q },
    })
  }

  async getReplies(form: Lemmy.GetReplies): Promise<Lemmy.GetRepliesResponse> {
    return { replies: [] } as Lemmy.GetRepliesResponse
  }

  async getPersonMentions(
    form: Lemmy.GetPersonMentions,
  ): Promise<Lemmy.GetPersonMentionsResponse> {
    return { mentions: [] } as Lemmy.GetPersonMentionsResponse
  }

  async markAllAsRead(): Promise<Lemmy.GetRepliesResponse> {
    await this.#api('/notifications/read-all', { method: 'PUT' })
    return { replies: [] } as Lemmy.GetRepliesResponse
  }

  async markPersonMentionAsRead(
    form: Lemmy.MarkPersonMentionAsRead,
  ): Promise<Lemmy.PersonMentionResponse> {
    await this.#api(`/notifications/${form.person_mention_id}/read`, {
      method: 'PUT',
    })
    return {} as Lemmy.PersonMentionResponse
  }

  async markCommentReplyAsRead(
    form: Lemmy.MarkCommentReplyAsRead,
  ): Promise<Lemmy.CommentReplyResponse> {
    await this.#api(`/notifications/${form.comment_reply_id}/read`, {
      method: 'PUT',
    })
    return {} as Lemmy.CommentReplyResponse
  }

  async getUnreadCount(): Promise<Lemmy.GetUnreadCountResponse> {
    const data = await this.#api<{ count: number }>('/notifications/unread-count')
    return {
      replies: 0,
      mentions: 0,
      private_messages: 0,
      notifications: data.count,
    } as Lemmy.GetUnreadCountResponse
  }

  async getReportCount(form: Lemmy.GetReportCount): Promise<Lemmy.GetReportCountResponse> {
    return { community_id: undefined, comment_reports: 0, post_reports: 0, private_message_reports: 0 } as Lemmy.GetReportCountResponse
  }

  // Lemmy-specific stubs
  async getComments(form: Lemmy.GetComments): Promise<Lemmy.GetCommentsResponse> {
    return { comments: [] } as Lemmy.GetCommentsResponse
  }

  async createComment(form: Lemmy.CreateComment): Promise<Lemmy.CommentResponse> {
    return {
      comment_view: {
        comment: {
          id: 0,
          creator_id: 0,
          post_id: form.post_id,
          content: form.content || '',
          published: new Date().toISOString(),
          updated: undefined,
          deleted: false,
          removed: false,
          distinguished: false,
          path: '',
          local: true,
          ap_id: '',
        },
        creator: {
          id: 0,
          name: '',
          display_name: undefined,
          avatar: undefined,
          banned: false,
          published: new Date().toISOString(),
          actor_id: '',
          local: true,
          deleted: false,
          admin: false,
          bot_account: false,
          instance_id: 1,
        },
        post: {
          id: form.post_id,
          name: '',
          creator_id: 0,
          published: new Date().toISOString(),
          updated: undefined,
          nsfw: false,
          removed: false,
          deleted: false,
          local: true,
          locked: false,
          embed_title: undefined,
          embed_description: undefined,
          embed_video_url: undefined,
          thumbnail_url: undefined,
          ap_id: '',
          language_id: 0,
          featured_community: false,
          featured_local: false,
          url: undefined,
          body: undefined,
          alt_text: undefined,
        },
        community: {
          id: 0,
          name: '',
          title: '',
          description: undefined,
          removed: false,
          deleted: false,
          published: new Date().toISOString(),
          updated: undefined,
          actor_id: '',
          local: true,
          nsfw: false,
          posting_restricted_to_mods: false,
          instance_id: 1,
          icon: undefined,
          banner: undefined,
          hidden: false,
        },
        counts: {
          id: 0,
          comment_id: 0,
          score: 0,
          upvotes: 0,
          downvotes: 0,
          child_count: 0,
          hot_rank: 0,
          hot_rank_active: 0,
          reports: 0,
          score_tally: 0,
        },
        saved: false,
        creator_banned_from_community: false,
        creator_is_moderator: false,
        creator_blocked: false,
        subscribed: false,
      },
    }
  }

  async editComment(form: Lemmy.EditComment): Promise<Lemmy.CommentResponse> {
    return this.createComment(form, undefined)
  }

  async deleteComment(form: Lemmy.DeleteComment): Promise<Lemmy.CommentResponse> {
    return {
      comment_view: {
        comment: {
          id: form.comment_id,
          creator_id: 0,
          post_id: 0,
          content: '',
          published: new Date().toISOString(),
          updated: undefined,
          deleted: true,
          removed: false,
          distinguished: false,
          path: '',
          local: true,
          ap_id: '',
        },
        creator: {
          id: 0,
          name: '',
          display_name: undefined,
          avatar: undefined,
          banned: false,
          published: new Date().toISOString(),
          actor_id: '',
          local: true,
          deleted: false,
          admin: false,
          bot_account: false,
          instance_id: 1,
        },
        post: {
          id: 0,
          name: '',
          creator_id: 0,
          published: new Date().toISOString(),
          updated: undefined,
          nsfw: false,
          removed: false,
          deleted: false,
          local: true,
          locked: false,
          embed_title: undefined,
          embed_description: undefined,
          embed_video_url: undefined,
          thumbnail_url: undefined,
          ap_id: '',
          language_id: 0,
          featured_community: false,
          featured_local: false,
          url: undefined,
          body: undefined,
          alt_text: undefined,
        },
        community: {
          id: 0,
          name: '',
          title: '',
          description: undefined,
          removed: false,
          deleted: false,
          published: new Date().toISOString(),
          updated: undefined,
          actor_id: '',
          local: true,
          nsfw: false,
          posting_restricted_to_mods: false,
          instance_id: 1,
          icon: undefined,
          banner: undefined,
          hidden: false,
        },
        counts: {
          id: 0,
          comment_id: form.comment_id,
          score: 0,
          upvotes: 0,
          downvotes: 0,
          child_count: 0,
          hot_rank: 0,
          hot_rank_active: 0,
          reports: 0,
          score_tally: 0,
        },
        saved: false,
        creator_banned_from_community: false,
        creator_is_moderator: false,
        creator_blocked: false,
        subscribed: false,
      },
    }
  }

  async likeComment(form: Lemmy.LikeComment): Promise<Lemmy.CommentResponse> {
    return this.deleteComment({ comment_id: form.comment_id })
  }

  async createCommunity(form: Lemmy.CreateCommunity): Promise<Lemmy.CommunityResponse> {
    const data = await this.#api<Community>('/communities', {
      method: 'POST',
      body: {
        name: form.name,
        description: form.description ?? '',
        icon_url: form.icon ?? '',
        banner_url: form.banner ?? '',
        is_private: form.nsfw ?? false,
      },
    })
    return { community_view: mapCommunity(data) } as Lemmy.CommunityResponse
  }

  async editCommunity(form: Lemmy.EditCommunity): Promise<Lemmy.CommunityResponse> {
    const data = await this.#api<Community>(`/communities/${form.community_id}`, {
      method: 'PUT',
      body: {
        name: form.title,
        description: form.description ?? '',
        icon_url: form.icon ?? '',
        banner_url: form.banner ?? '',
      },
    })
    return { community_view: mapCommunity(data) } as Lemmy.CommunityResponse
  }

  async deleteCommunity(form: Lemmy.DeleteCommunity): Promise<Lemmy.CommunityResponse> {
    await this.#api(`/communities/${form.community_id}`, { method: 'DELETE' })
    return {} as Lemmy.CommunityResponse
  }

  async banFromCommunity(form: Lemmy.BanFromCommunity): Promise<Lemmy.BanFromCommunityResponse> {
    return { banned: true } as Lemmy.BanFromCommunityResponse
  }

  async addModToCommunity(form: Lemmy.AddModToCommunity): Promise<Lemmy.AddModToCommunityResponse> {
    return { moderators: [] } as Lemmy.AddModToCommunityResponse
  }

  async transferCommunity(form: Lemmy.TransferCommunity): Promise<Lemmy.GetCommunityResponse> {
    return {} as Lemmy.GetCommunityResponse
  }

  async blockCommunity(form: Lemmy.BlockCommunity): Promise<Lemmy.BlockCommunityResponse> {
    return { blocked: true, community_view: undefined as unknown as Lemmy.CommunityView } as Lemmy.BlockCommunityResponse
  }

  async hideCommunity(form: Lemmy.HideCommunity): Promise<Lemmy.SuccessResponse> {
    return { success: true } as Lemmy.SuccessResponse
  }

  async blockPerson(form: Lemmy.BlockPerson): Promise<Lemmy.BlockPersonResponse> {
    return { blocked: true, person_view: undefined as unknown as Lemmy.PersonView } as Lemmy.BlockPersonResponse
  }

  async resolveObject(form: Lemmy.ResolveObject): Promise<Lemmy.ResolveObjectResponse> {
    return { comment: undefined, community: undefined, post: undefined, person: undefined } as Lemmy.ResolveObjectResponse
  }

  async markPrivateMessageAsRead(form: Lemmy.MarkPrivateMessageAsRead): Promise<Lemmy.PrivateMessageResponse> {
    return {} as Lemmy.PrivateMessageResponse
  }

  async createPrivateMessage(form: Lemmy.CreatePrivateMessage): Promise<Lemmy.PrivateMessageResponse> {
    return {} as Lemmy.PrivateMessageResponse
  }

  async listPrivateMessages(form: Lemmy.ListPrivateMessages): Promise<Lemmy.ListPrivateMessagesResponse> {
    return { private_messages: [] } as Lemmy.ListPrivateMessagesResponse
  }

  async transferSite(form: Lemmy.TransferSite): Promise<Lemmy.GetSiteResponse> {
    return {} as Lemmy.GetSiteResponse
  }

  async getSiteMetadata(form: Lemmy.GetSiteMetadata): Promise<Lemmy.GetSiteMetadataResponse> {
    return { metadata: { description: undefined, image: undefined, title: undefined } } as Lemmy.GetSiteMetadataResponse
  }

  async getModlog(form: Lemmy.GetModlog): Promise<Lemmy.GetModlogResponse> {
    return {
      added: [],
      added_to_community: [],
      admin_purged_comments: [],
      admin_purged_communities: [],
      admin_purged_posts: [],
      admin_purged_persons: [],
      banned: [],
      banned_from_community: [],
      featured_posts: [],
      locked_posts: [],
      mod_remove_comments: [],
      mod_remove_communities: [],
      mod_remove_posts: [],
      mod_remove_persons: [],
      mod_transfer_community: [],
      mod_transfer_site: [],
      removed_posts: [],
      restored_posts: [],
      mod_change_password: [],
      mod_remove_post_origin: [],
    } as Lemmy.GetModlogResponse
  }

  async listPostReports(form: Lemmy.ListPostReports): Promise<Lemmy.ListPostReportsResponse> {
    return { post_reports: [] } as Lemmy.ListPostReportsResponse
  }

  async listCommentReports(form: Lemmy.ListCommentReports): Promise<Lemmy.ListCommentReportsResponse> {
    return { comment_reports: [] } as Lemmy.ListCommentReportsResponse
  }

  async resolvePostReport(form: Lemmy.ResolvePostReport): Promise<Lemmy.ResolvePostReportResponse> {
    return { post_report_view: undefined as unknown as Lemmy.PostReportView } as Lemmy.ResolvePostReportResponse
  }

  async resolveCommentReport(form: Lemmy.ResolveCommentReport): Promise<Lemmy.ResolveCommentReportResponse> {
    return { comment_report_view: undefined as unknown as Lemmy.CommentReportView } as Lemmy.ResolveCommentReportResponse
  }

  async getUnregenerateKeys(): Promise<Lemmy.SuccessResponse> {
    return { success: true } as Lemmy.SuccessResponse
  }

  async createSite(form: Lemmy.CreateSite): Promise<Lemmy.SiteResponse> {
    return {} as Lemmy.SiteResponse
  }

  async listMedia(): Promise<Lemmy.ListMediaResponse> {
    return { media: [] } as Lemmy.ListMediaResponse
  }

  async purgePerson(form: Lemmy.PurgePerson): Promise<Lemmy.SuccessResponse> {
    return { success: true } as Lemmy.SuccessResponse
  }

  async purgeCommunity(form: Lemmy.PurgeCommunity): Promise<Lemmy.SuccessResponse> {
    return { success: true } as Lemmy.SuccessResponse
  }

  async purgePost(form: Lemmy.PurgePost): Promise<Lemmy.SuccessResponse> {
    return { success: true } as Lemmy.SuccessResponse
  }

  async purgeComment(form: Lemmy.PurgeComment): Promise<Lemmy.SuccessResponse> {
    return { success: true } as Lemmy.SuccessResponse
  }

  async getBannedPersons(): Promise<Lemmy.BannedPersonsResponse> {
    return { banned: [] } as Lemmy.BannedPersonsResponse
  }

  async addAdmin(form: Lemmy.AddAdmin): Promise<Lemmy.AddAdminResponse> {
    return { admins: [] } as Lemmy.AddAdminResponse
  }

  async getUnreadRegistrationApplicationCount(): Promise<Lemmy.GetUnreadRegistrationApplicationCountResponse> {
    return { registration_applications: 0 } as Lemmy.GetUnreadRegistrationApplicationCountResponse
  }

  async listRegistrationApplications(form: Lemmy.ListRegistrationApplications): Promise<Lemmy.ListRegistrationApplicationsResponse> {
    return { registration_applications: [] } as Lemmy.ListRegistrationApplicationsResponse
  }

  async approveRegistrationApplication(form: Lemmy.ApproveRegistrationApplication): Promise<Lemmy.RegistrationApplicationResponse> {
    return { registration_application: undefined as unknown as Lemmy.RegistrationApplicationView } as Lemmy.RegistrationApplicationResponse
  }

  async getTags(): Promise<Lemmy.GetTagsResponse> {
    const tags = await this.#api<Tag[]>('/tags')
    return {
      tags: (tags ?? []).map((t: Tag) => ({
        id: t.id,
        name: t.name,
        description: t.description,
        category: t.category,
        total_votes: t.total_votes ?? 0,
        vote_score: t.vote_score ?? 0,
        user_vote: t.user_vote ?? 0,
      })),
    } as unknown as Lemmy.GetTagsResponse
  }

  async voteOnTag(tagId: number, vote: number): Promise<Tag> {
    return this.#api<Tag>(`/tags/${tagId}/vote`, {
      method: 'POST',
      body: { vote } as TagVoteRequest,
    })
  }

  async removeTagVote(tagId: number): Promise<{ message: string }> {
    return this.#api<{ message: string }>(`/tags/${tagId}/vote`, {
      method: 'DELETE',
    })
  }

  async searchPostLikes(form: { post_id: number; page?: number; limit?: number }): Promise<Lemmy.SearchResponse> {
    return { posts: [], comments: [], communities: [], users: [], type_: undefined } as Lemmy.SearchResponse
  }

  async getPersonPostsLiked(form: { person_id: number }): Promise<Lemmy.GetPostsResponse> {
    return { posts: [], next_page: undefined } as Lemmy.GetPostsResponse
  }

  async getPersonPostsSaved(form: { person_id: number }): Promise<Lemmy.GetPostsResponse> {
    return { posts: [], next_page: undefined } as Lemmy.GetPostsResponse
  }

  async featurePost(form: Lemmy.FeaturePost): Promise<Lemmy.PostResponse> {
    return {} as Lemmy.PostResponse
  }

  async lockPost(form: Lemmy.LockPost): Promise<Lemmy.PostResponse> {
    return {} as Lemmy.PostResponse
  }

  async removePost(form: Lemmy.RemovePost): Promise<Lemmy.PostResponse> {
    return {} as Lemmy.PostResponse
  }

  async removeComment(form: Lemmy.RemoveComment): Promise<Lemmy.CommentResponse> {
    return {} as Lemmy.CommentResponse
  }

  async distinguishComment(form: Lemmy.DistinguishComment): Promise<Lemmy.CommentResponse> {
    return {} as Lemmy.CommentResponse
  }

  async distinguishPost(form: Lemmy.DistinguishPost): Promise<Lemmy.PostResponse> {
    return {} as Lemmy.PostResponse
  }

  async createPostReport(form: Lemmy.CreatePostReport): Promise<Lemmy.PostReportResponse> {
    await this.#api('/reports', {
      method: 'POST',
      body: { target_type: 'post', target_id: form.post_id, reason: form.reason },
    })
    return {} as Lemmy.PostReportResponse
  }

  async createCommentReport(form: Lemmy.CreateCommentReport): Promise<Lemmy.CommentReportResponse> {
    await this.#api('/reports', {
      method: 'POST',
      body: { target_type: 'comment', target_id: form.comment_id, reason: form.reason },
    })
    return {} as Lemmy.CommentReportResponse
  }

  async getComment(form: Lemmy.GetComment): Promise<Lemmy.GetCommentResponse> {
    return { comment_view: undefined as unknown as Lemmy.CommentView } as Lemmy.GetCommentResponse
  }

  async getVotes(form: Lemmy.GetVotes): Promise<Lemmy.GetVotesResponse> {
    return { type_: form.type_, comments: [], posts: [] } as unknown as Lemmy.GetVotesResponse
  }

  async getVote(form: { post_id?: number; comment_id?: number }): Promise<Lemmy.GetVotesResponse> {
    return { type_: undefined, comments: [], posts: [] } as unknown as Lemmy.GetVotesResponse
  }

  async createCommentLike(form: Lemmy.LikeComment): Promise<Lemmy.CommentResponse> {
    return this.deleteComment({ comment_id: form.comment_id })
  }

  async getPostLikes(form: { post_id: number }): Promise<Lemmy.GetPostLikesResponse> {
    return { post_likes: [] } as Lemmy.GetPostLikesResponse
  }

  async getCommentLikes(form: { comment_id: number }): Promise<Lemmy.GetCommentLikesResponse> {
    return { comment_likes: [] } as Lemmy.GetCommentLikesResponse
  }

  async saveComment(form: Lemmy.SaveComment): Promise<Lemmy.CommentResponse> {
    return this.createComment({ post_id: 0, content: '' })
  }

  async saveUserSettingsFast(form: Lemmy.SaveUserSettings): Promise<Lemmy.SuccessResponse> {
    await this.saveUserSettings(form)
    return { success: true } as Lemmy.SuccessResponse
  }

  async changePassword(form: Lemmy.ChangePassword): Promise<Lemmy.LoginResponse> {
    return { jwt: '' } as Lemmy.LoginResponse
  }

  async getPersonDetailsFast(form: Lemmy.GetPersonDetails): Promise<Lemmy.GetPersonDetailsResponse> {
    return this.getPerson(form)
  }

  async getSiteFast(): Promise<Lemmy.GetSiteResponse> {
    return this.getSite()
  }

  async getPostsFast(form: Lemmy.GetPosts): Promise<Lemmy.GetPostsResponse> {
    return this.getPosts(form)
  }

  async getCommunityFast(form: Lemmy.GetCommunity): Promise<Lemmy.GetCommunityResponse> {
    return this.getCommunity(form)
  }

  async getUnreadCountFast(): Promise<Lemmy.GetUnreadCountResponse> {
    return this.getUnreadCount()
  }

  async markAllAsReadFast(): Promise<Lemmy.GetRepliesResponse> {
    return this.markAllAsRead()
  }

  async searchFast(form: Lemmy.Search): Promise<Lemmy.SearchResponse> {
    return this.search(form)
  }

  async getSiteMetadataFast(form: Lemmy.GetSiteMetadata): Promise<Lemmy.GetSiteMetadataResponse> {
    return this.getSiteMetadata(form)
  }

  async markCommentReplyAsReadFast(form: Lemmy.MarkCommentReplyAsRead): Promise<Lemmy.CommentReplyResponse> {
    return this.markCommentReplyAsRead(form)
  }

  async markPersonMentionAsReadFast(form: Lemmy.MarkPersonMentionAsRead): Promise<Lemmy.PersonMentionResponse> {
    return this.markPersonMentionAsRead(form)
  }

  async getRepliesFast(form: Lemmy.GetReplies): Promise<Lemmy.GetRepliesResponse> {
    return this.getReplies(form)
  }

  async getPersonMentionsFast(form: Lemmy.GetPersonMentions): Promise<Lemmy.GetPersonMentionsResponse> {
    return this.getPersonMentions(form)
  }

  // Community-specific mod methods
  async listCommunityModerators(form: { community_id: number }): Promise<Lemmy.ListCommunityModeratorsResponse> {
    return { moderators: [] } as unknown as Lemmy.ListCommunityModeratorsResponse
  }

  async getCommunityByName(form: { name: string }): Promise<Lemmy.GetCommunityResponse> {
    const data = await this.#api<Community[]>('/communities', {
      params: { name: form.name },
    })
    const community = (data ?? [])[0]
    if (!community) throw new Error('Community not found')
    return this.getCommunity({ id: community.id })
  }

  async searchByName(form: { q: string }): Promise<Lemmy.SearchResponse> {
    return this.search({ q: form.q })
  }

  async logout(): Promise<Lemmy.LoginResponse> {
    await this.#api('/auth/logout', { method: 'DELETE' })
    return { jwt: '' } as Lemmy.LoginResponse
  }

  async editSite(form: Lemmy.EditSite): Promise<Lemmy.SiteResponse> {
    await this.#api('/admin/config', {
      method: 'PUT',
      body: {
        title: form.name,
        description: form.description ?? form.sidebar ?? '',
        invite_only: form.registration_mode === 'RequireApplication',
        registration_open: form.registration_mode !== 'Closed',
      },
    })
    return {} as Lemmy.SiteResponse
  }

  async generateTotpSecret(): Promise<Lemmy.GenerateTotpSecretResponse> {
    return { totp_secret_url: '', totp_secret: '' } as Lemmy.GenerateTotpSecretResponse
  }

  async listLogins(): Promise<Lemmy.LoginToken[]> {
    return []
  }

  async listAllMedia(form: Lemmy.ListMedia): Promise<Lemmy.ListMediaResponse> {
    return { media: [] } as Lemmy.ListMediaResponse
  }

  async updateTotp(form: Lemmy.UpdateTotp): Promise<Lemmy.UpdateTotpResponse> {
    return { enabled: true } as Lemmy.UpdateTotpResponse
  }

  async removeCommunity(form: Lemmy.RemoveCommunity): Promise<Lemmy.CommunityResponse> {
    return {} as Lemmy.CommunityResponse
  }

  async banPerson(form: Lemmy.BanPerson): Promise<Lemmy.BanPersonResponse> {
    return { banned: true, person_view: undefined as unknown as Lemmy.PersonView } as Lemmy.BanPersonResponse
  }

  async getCaptcha(): Promise<Lemmy.GetCaptchaResponse> {
    return { ok: undefined } as Lemmy.GetCaptchaResponse
  }

  async deleteAccount(form: Lemmy.DeleteAccount): Promise<Lemmy.LoginResponse> {
    return { jwt: '' } as Lemmy.LoginResponse
  }

  async passwordReset(form: Lemmy.PasswordReset): Promise<Lemmy.PasswordResetResponse> {
    await this.#api('/auth/password-reset/forgot', {
      method: 'POST',
      body: { email: form.email },
    })
    return {} as Lemmy.PasswordResetResponse
  }

  async passwordChangeAfterReset(form: Lemmy.PasswordChangeAfterReset): Promise<Lemmy.LoginResponse> {
    await this.#api('/auth/password-reset/reset', {
      method: 'POST',
      body: { token: form.token, password: form.password },
    })
    return { jwt: 'session' } as Lemmy.LoginResponse
  }

  async verifyEmail(form: Lemmy.VerifyEmail): Promise<Lemmy.LoginResponse> {
    return { jwt: '' } as Lemmy.LoginResponse
  }

  async getFederatedInstances(): Promise<Lemmy.GetFederatedInstancesResponse> {
    return { federated_instances: { linked: [], allowed: [], blocked: [] } } as Lemmy.GetFederatedInstancesResponse
  }

  async blockInstance(form: Lemmy.BlockInstance): Promise<Lemmy.BlockInstanceResponse> {
    return { blocked: true, instance_view: undefined as unknown as Lemmy.InstanceView } as Lemmy.BlockInstanceResponse
  }

  async uploadImage(form: FormData): Promise<Lemmy.UploadImageResponse> {
    return { files: [], msg: '' } as Lemmy.UploadImageResponse
  }

  async deleteImage(form: { filename: string }): Promise<Lemmy.SuccessResponse> {
    return { success: true } as Lemmy.SuccessResponse
  }

  async getPostReplies(form: { post_id: number }): Promise<Lemmy.GetCommentsResponse> {
    return { comments: [] } as Lemmy.GetCommentsResponse
  }

  async createCommentReportFast(form: Lemmy.CreateCommentReport): Promise<Lemmy.CommentReportResponse> {
    return this.createCommentReport(form)
  }

  async createCommunityFast(form: Lemmy.CreateCommunity): Promise<Lemmy.CommunityResponse> {
    return this.createCommunity(form)
  }

  async createPostFast(form: Lemmy.CreatePost): Promise<Lemmy.PostResponse> {
    return this.createPost(form)
  }

  async followCommunityFast(form: Lemmy.FollowCommunity): Promise<Lemmy.CommunityResponse> {
    return this.followCommunity(form)
  }

  async getCommunitiesFast(form: Lemmy.ListCommunities): Promise<Lemmy.ListCommunitiesResponse> {
    return this.getCommunities(form)
  }

  async loginFast(form: Lemmy.Login): Promise<Lemmy.LoginResponse> {
    return this.login(form)
  }

  async registerFast(form: Lemmy.Register): Promise<Lemmy.LoginResponse> {
    return this.register(form)
  }

  async changePasswordFast(form: Lemmy.ChangePassword): Promise<Lemmy.LoginResponse> {
    return this.changePassword(form)
  }
}

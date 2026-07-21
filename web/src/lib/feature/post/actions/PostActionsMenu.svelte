<script lang="ts">
  import { PiefedClient } from '$lib/api/piefed/adapter'
  import type { PostView } from '$lib/api/types'
  import { profile } from '$lib/app/auth'

  import { settings } from '$lib/app/settings.svelte'
  import { deleteItem, markAsRead } from '$lib/feature/legacy/contentview'
  import { report } from '$lib/feature/moderation/moderation'
  import { MenuButton } from 'mono-svelte'
  import {
    ArrowTopRightOnSquare,
    Eye,
    EyeSlash,
    Flag,
    PencilSquare,
    Trash,
    XMark,
  } from 'svelte-hero-icons/dist'
  import { type PostFormInit } from '../form/postform.svelte'
  import { hidePost } from '../helpers'

  interface Props {
    post: PostView
    onhide?: (hide: boolean) => void
    editing: boolean
  }

  let { post = $bindable(), onhide, editing = $bindable() }: Props = $props()

  function crosspostB64() {
    return JSON.stringify({
      body: `${
        settings.crosspostOriginalLink
          ? `cross-posted from: ${post.post.ap_id}`
          : ``
      }\n${
        post.post.body ? '>' + post.post.body.split('\n').join('\n> ') : ''
      }`,
      name: post.post.name,
      url: post.post.url,
      nsfw: post.post.nsfw,
    } as PostFormInit)
  }
</script>

{#if profile.current?.user && profile.current?.jwt && profile.current.user.local_user_view.person.id == post.creator.id}
  <MenuButton onclick={() => (editing = true)} icon={PencilSquare}>
    Edit
  </MenuButton>
{/if}
{#if profile.current?.jwt}
  <MenuButton
    onclick={async () => {
      if (profile.current?.jwt)
        post.read = await markAsRead(post.post, !post.read)
    }}
    icon={post.read ? EyeSlash : Eye}
  >
    {post.read
      ?'Mark as unread'
      :'Mark as read'}
  </MenuButton>
{/if}
{#if profile.current?.jwt}
  <MenuButton
    href="/create/post?crosspost={crosspostB64()}"
    icon={ArrowTopRightOnSquare}
  >
    Crosspost
  </MenuButton>
  {#if profile.current.user && post.creator.id == profile.current.user.local_user_view.person.id}
    <MenuButton
      onclick={async () => {
        if (profile.current?.jwt)
          post.post.deleted = await deleteItem(post, !post.post.deleted)
      }}
      color="danger-subtle"
      icon={Trash}
    >
      {post.post.deleted
        ?'Restore'
        :'Delete'}
    </MenuButton>
  {/if}
  {#if profile.current.user?.local_user_view.person.id != post.creator.id}
    {#if !(profile.client instanceof PiefedClient)}
      <MenuButton
        onclick={async () => {
          if (!profile.current?.jwt) return
          const hidden = await hidePost(
            post.post.id,
            !post.hidden,
            profile.current?.jwt,
          )
          post.hidden = hidden
          if (hidden) onhide?.(hidden)
        }}
        color="danger-subtle"
        icon={XMark}
      >
        {post.hidden
          ?'Unhide'
          :'Hide'}
      </MenuButton>
    {/if}
    <MenuButton onclick={() => report(post)} color="danger-subtle" icon={Flag}>
      Report
    </MenuButton>
  {/if}
{/if}

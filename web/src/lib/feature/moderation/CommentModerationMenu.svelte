<script lang="ts">
  import type { CommentView } from '$lib/api/types'
  import { profile } from '$lib/app/auth'

  import { Menu, MenuButton, MenuDivider } from 'mono-svelte'
  import type { Snippet } from 'svelte'
  import {
    ArrowsUpDown,
    Fire,
    Icon,
    Megaphone,
    Newspaper,
    ShieldExclamation,
    Trash,
  } from 'svelte-hero-icons/dist'
  import type { Attachment } from 'svelte/attachments'
  import { isCommentView } from '../legacy/item'
  import { ban, feature, remove, viewVotes } from './moderation'

  interface Props {
    item: CommentView
    target: Snippet<[Attachment]>
  }

  let { item = $bindable(), target, ...rest }: Props = $props()
</script>

<Menu {...rest} placement="bottom" {target}>
  {#if profile.isMod(item.community) || profile.isAdmin}
    <MenuDivider hidden>
      {#if !item.community.local && !profile.isMod(item.community)}
        Moderation (Local only)
      {:else}
        Moderation
      {/if}
    </MenuDivider>
    <MenuButton color="danger-subtle" onclick={() => remove(item)} icon={Trash}>
      {#if isCommentView(item)}
        {item.comment.removed
          ?'Restore'
          :'Remove'}
      {/if}
    </MenuButton>
    {#if profile.current?.user && profile.current.user?.local_user_view.person.id != item.creator.id}
      <!--Comment made by someone else-->
      <MenuButton
        color="danger-subtle"
        onclick={() =>
          ban(item.creator_banned_from_community, item.creator, item.community)}
        icon={ShieldExclamation}
      >
        {item.creator_banned_from_community
          ?'Unban user from community'
          :'Ban user from community'}
      </MenuButton>
    {:else}
      <!--Comment made by self-->
      <MenuButton
        color="success-subtle"
        onclick={async () => {
          if (!profile.current.jwt) return
          item.comment = (
            await feature(
              !item.comment.distinguished,
              item.comment,
              profile.current.jwt,
            )
          ).comment_view.comment
        }}
        icon={Megaphone}
      >
        {item.comment.distinguished
          ?'Unfeature'
          :'Feature'}
      </MenuButton>
    {/if}

    <MenuButton color="success-subtle" href="/modlog?comment={item.comment.id}">
      <Icon src={Newspaper} size="16" micro />
      Comment moderation log
    </MenuButton>
    <MenuButton color="success-subtle" href="/modlog?user={item.creator.id}">
      <Icon src={Newspaper} size="16" micro />
      User moderation log
    </MenuButton>
    <MenuButton color="blue-subtle" onclick={() => viewVotes(item)}>
      <Icon src={ArrowsUpDown} size="16" micro />
      Votes
    </MenuButton>
  {/if}

  {#if profile.isAdmin}
    <MenuDivider showLabel>Administration</MenuDivider>
    <MenuButton
      color="danger-subtle"
      onclick={() => remove(item, true)}
      icon={Fire}
    >
      Purge
    </MenuButton>
  {/if}
</Menu>

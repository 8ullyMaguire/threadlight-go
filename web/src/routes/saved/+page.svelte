<script lang="ts">
  import { page } from '$app/state'
  import type { CommentView, PostView } from '$lib/api/types'

  import { searchParam } from '$lib/app/util.svelte'
  import CommentItem from '$lib/feature/comment/CommentItem.svelte'
  import { PostItem } from '$lib/feature/post'
  import Fixate from '$lib/ui/generic/Fixate.svelte'
  import Placeholder from '$lib/ui/info/Placeholder.svelte'
  import { CommonList, Header, Pageination } from '$lib/ui/layout'
  import { Option, Select } from 'mono-svelte'
  import {
    AdjustmentsHorizontal,
    Bars3,
    Bookmark,
    ChatBubbleOvalLeft,
    Icon,
    PencilSquare,
  } from 'svelte-hero-icons/dist'

  let { data } = $props()

  let type = $state(data.type)!

  const isComment = (item: CommentView | PostView): item is CommentView =>
    'comment' in item
</script>

<svelte:head>
  <title>Saved</title>
</svelte:head>

<Header pageHeader>
  Saved

  {#snippet extended()}
    <div class="flex items-center">
      <Select
        bind:value={type}
        onchange={() => searchParam(page.url, 'type', type, 'page')}
      >
        {#snippet customLabel()}
          <div class="flex items-center gap-0.5">
            <Icon src={AdjustmentsHorizontal} size="15" mini />
            Filter
          </div>
        {/snippet}
        <Option icon={Bars3} value="all">
          All
        </Option>
        <Option icon={PencilSquare} value="posts">
          Posts
        </Option>
        <Option icon={ChatBubbleOvalLeft} value="comments">
          Comments
        </Option>
      </Select>
    </div>
  {/snippet}
</Header>

{#if !data.data || (data.data?.length ?? 0) == 0}
  <Placeholder
    icon={Bookmark}
    title="No saved items"
    description="Save posts or comments, and they'll be here to refer to them later."
    class="my-auto"
  />
{:else}
  <CommonList items={data.data} size="lg">
    {#snippet item(item)}
      {#if isComment(item)}
        <CommentItem comment={item} />
      {:else}
        <PostItem post={item} />
      {/if}
    {/snippet}
  </CommonList>
{/if}
<Fixate placement="bottom">
  <Pageination
    hasMore={data.data.length == 40}
    href={(page) => `?page=${page}`}
    page={data.page}
  />
</Fixate>

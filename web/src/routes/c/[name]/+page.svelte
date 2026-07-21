<script lang="ts">
  import { browser } from '$app/environment'
  import { navigating, page } from '$app/state'
  import { client } from '$lib/api/client.svelte'
  import { site } from '$lib/api/client.svelte'
  import { profile } from '$lib/app/auth'

  import { setSessionStorage } from '$lib/app/session'
  import CommunityHeader from '$lib/feature/community/CommunityHeader.svelte'
  import { resumables } from '$lib/feature/legacy/item'
  import { PostListShell } from '$lib/ui/layout'
  import { Badge, Button, Note, Spinner } from 'mono-svelte'
  import { onDestroy, onMount } from 'svelte'
  import { ArrowRight, CurrencyDollar, Icon, Plus } from 'svelte-hero-icons/dist'

  let { data } = $props()

  let communityBalance = $state<number | null>(null)
  let balanceLoading = $state(false)

  let isModOrCurator = $derived(
    profile.current?.user && data.community?.moderators?.some(
      (m: { moderator?: { id: number }; person?: { id: number } }) =>
        (m.moderator?.id ?? m.person?.id) === profile.current?.user?.local_user_view?.person?.id,
    ),
  )

  onMount(async () => {
    if (browser)
      setSessionStorage('lastSeenCommunity', data.community.community_view)

    resumables.add({
      name: data.community.community_view.community.title,
      type: 'community',
      url: page.url.toString(),
      avatar: data.community.community_view.community.icon,
    })

    if (isModOrCurator && browser) {
      balanceLoading = true
      try {
        const tl = client() as unknown as import('$lib/api/threadlight/adapter').ThreadlightClient
        const bal = await tl.getCommunityBalance(data.community.community_view.community.name)
        communityBalance = bal.balance
      } catch {
        // non-threadlight backend or not available
      } finally {
        balanceLoading = false
      }
    }
  })

  onDestroy(() => {
    if (browser) {
      if (navigating?.to?.route?.id == '/create/post') return

      setSessionStorage('lastSeenCommunity', undefined)
    }
  })
</script>

<svelte:head>
  <title>{data.community.community_view.community.title}</title>

  <meta
    name="og:title"
    content={data.community.community_view.community.title}
  />
  {#if data.community.community_view.community.description}
    <meta
      name="og:description"
      content={data.community.community_view.community.description}
    />
  {/if}
</svelte:head>

<PostListShell
  bind:posts={data.posts}
  bind:cursor={data.next_page}
  bind:client={data.client}
  getParams={data.params}
  params={{
    sort: data.params.sort!,
  }}
>
  {#snippet extended()}
    <CommunityHeader
      bind:community={data.community.community_view.community}
      bind:subscribed={data.community.community_view.subscribed}
      blocked={data.community.community_view.blocked}
      moderators={data.community.moderators}
      counts={data.community.community_view.counts}
      class="w-full relative"
      compact="lg"
      avatarCircle={false}
    />
    {#if isModOrCurator}
      {#if balanceLoading}
        <div class="flex items-center gap-1 text-xs text-zinc-500 px-2 py-1">
          <Spinner width={12} />
          <span>Treasury...</span>
        </div>
      {:else if communityBalance !== null}
        <div class="flex items-center gap-1 text-xs font-medium text-amber-600 dark:text-amber-400 bg-amber-50 dark:bg-amber-950/20 rounded-lg px-3 py-1.5 w-fit mb-2">
          <Icon src={CurrencyDollar} size="14" micro />
          <span>Treasury: <strong>{communityBalance}</strong> credits</span>
        </div>
      {/if}
    {/if}
    {#if data.community.community_view.blocked}
      <Note>You've blocked this community.</Note>
    {/if}
    {#if profile.current.user}
      {#if !data.community.discussion_languages.every( (l) => profile.current.user?.discussion_languages.includes(l), ) && profile.current.user.discussion_languages.length > 0}
        {@const missing = data.community.discussion_languages.filter(
          (i) => !profile.current.user?.discussion_languages.includes(i),
        )}
        <Note class="p-1! pl-3! flex-col md:flex-row">
          <div>Content in this community is in languages you do not have selected. Some or all posts may be missing.</div>
          <Button
            class="inline-block ml-auto"
            href="/profile/settings"
            color="tertiary"
            rounding="pill"
            size="md"
          >
            Profile
            {#snippet suffix()}
              <Icon src={ArrowRight} size="16" micro />
            {/snippet}
          </Button>
        </Note>
        <div class="flex flex-row gap-4 flex-wrap -mt-2">
          {#if site.data?.all_languages}
            {@const allLanguages = site.data.all_languages}
            {#each missing as language (language)}
              <a href="/profile/settings#languages" class="inline-block w-max">
                <Badge color="blue-subtle">
                  <Icon src={Plus} size="16" micro />
                  {allLanguages.find((i) => language == i.id)?.name}
                </Badge>
              </a>
            {/each}
          {/if}
        </div>
      {/if}
    {/if}
  {/snippet}
</PostListShell>

<script lang="ts">
  import { page } from '$app/state'
  import { client } from '$lib/api/client.svelte'
  import { profile } from '$lib/app/auth'

  import { LINKED_INSTANCE_URL } from '$lib/app/instance.svelte'
  import Location from '$lib/feature/filter/Location.svelte'
  import Sort from '$lib/feature/filter/Sort.svelte'
  import { Header, SearchBar, Tabs } from '$lib/ui/layout'
  import { Option } from 'mono-svelte'
  import { ServerStack } from 'svelte-hero-icons/dist'

  let { data, children } = $props()

  let search = $state(page.data.query || '')
  let form = $state<HTMLFormElement>()
</script>

{#if client().getTopics && client().getFeeds}
  <Tabs
    routes={[
      {
        href: '/explore/communities',
        name:'Communities',
      },
      {
        href: '/explore/feeds',
        name:'Feeds',
      },
      {
        href: '/explore/topics',
        name:'Topics',
      },
    ]}
  ></Tabs>
{/if}

<svelte:head>
  <title>Explore</title>
</svelte:head>

<Header pageHeader>
  Explore
  {#snippet extended()}
    {#if page.route.id == '/explore/communities'}
      <form
        method="get"
        action={page.url.pathname}
        class="contents"
        bind:this={form}
      >
        <SearchBar bind:query={search} />

        <div class="flex flex-row flex-wrap gap-4 items-center">
          <Location
            name="type"
            selected={data.type}
            onchange={() => form?.requestSubmit()}
          >
            {#if !LINKED_INSTANCE_URL}
              {@const instanceSet = new Set(
                profile.meta.profiles.map((i) => i.instance),
              )}
              {#if instanceSet.size > 1}
                <Option disabled data-label="true">—</Option>
                {#each instanceSet as instance}
                  <Option
                    icon={ServerStack}
                    value={encodeURIComponent(`instance-${instance}`)}
                  >
                    {instance}
                  </Option>
                {/each}
              {/if}
            {/if}
          </Location>
          <Sort
            name="sort"
            selected={data.sort}
            onchange={() => form?.requestSubmit()}
          />
        </div>
      </form>
    {/if}
  {/snippet}
</Header>

{@render children?.()}

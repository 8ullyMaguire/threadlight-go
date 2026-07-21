<script lang="ts">
  import { goto } from '$app/navigation'
  import { page } from '$app/state'
  import { client } from '$lib/api/client.svelte'

  import { searchParam } from '$lib/app/util.svelte'
  import CommunityLink from '$lib/feature/community/CommunityLink.svelte'
  import UserLink from '$lib/feature/user/UserLink.svelte'
  import Fixate from '$lib/ui/generic/Fixate.svelte'
  import Placeholder from '$lib/ui/info/Placeholder.svelte'
  import ProgressBar from '$lib/ui/info/ProgressBar.svelte'
  import { CommonList, Header, Pageination } from '$lib/ui/layout'
  import { Button, Material, Option, Select, Spinner, toast } from 'mono-svelte'
  import { tick } from 'svelte'
  import {
    Check,
    Funnel,
    Icon,
    ShieldCheck,
    XMark,
    Clock,
  } from 'svelte-hero-icons/dist'
  import RelativeDate from 'mono-svelte/util/RelativeDate.svelte'
  import Report from './Report.svelte'
  import type { ModLog } from '../modlog/+page'
  import { _toModLog } from '../modlog/+page'

  let { data = $bindable() } = $props()

  let batch = $state({
    progress: -1,
  })
  async function markAllAsResolved() {
    if (!data.items?.value) return
    batch.progress = 0

    await Promise.all(
      data.items?.value.map((report) => {
        report.map((r) => {
          switch (r.type) {
            case 'comment': {
              const promise = client().resolveCommentReport({
                report_id: r.id,
                resolved: true,
              })
              promise.then(
                () => (batch.progress += 1 / data.items?.value!.length),
              )
              return promise
            }
            case 'post': {
              const promise = client().resolvePostReport({
                report_id: r.id,
                resolved: true,
              })
              promise.then(
                () => (batch.progress += 1 / data.items?.value!.length),
              )
              return promise
            }
            case 'message': {
              const promise = client().resolvePrivateMessageReport({
                report_id: r.id,
                resolved: true,
              })
              promise.then(
                () => (batch.progress += 1 / data.items?.value!.length),
              )
              return promise
            }
          }
        })
      }),
    )

    toast({ content:'Batch action completed.', type: 'success' })
    await goto(page.url, { invalidateAll: true })

    batch.progress = -1
  }

  // Map moderation actions to ModLog items for display
  let actionLogItems = $derived(
    (data.moderationActions ?? []).map((action: any) => ({
      actionName: 'Unknown' as const,
      timestamp: new Date(action.created_at).getTime(),
      reason: action.reason || '',
      moderator: action.moderator || undefined,
      community: action.community || undefined,
      moderatee: action.target_user || undefined,
      content: action.target_post?.title || action.reason || '',
      id: action.id,
      link: action.target_post_id ? `/post/${action.target_post_id}` : undefined,
    })) as Array<ModLog & { id?: number; link?: string }>,
  )
</script>

<Header pageHeader>
  Moderation
  {#snippet extended()}
    <div class="flex flex-row gap-2 flex-wrap items-end">
      <Select
        bind:value={data.type.value}
        onchange={async () => {
          await tick()
          searchParam(page.url, 'type', data.type.value, 'page')
        }}
      >
        {#snippet customLabel()}
          <span class="flex items-center gap-1">
            <Icon src={Funnel} size="15" mini />
            Filter
          </span>
        {/snippet}
        <Option value="all">All</Option>
        <Option value="unread">Unread</Option>
      </Select>
      <Button
        loading={batch.progress >= 0}
        disabled={batch.progress >= 0 || data.items.value?.length == 0}
        onclick={markAllAsResolved}
        size="lg"
        class="ml-auto"
        color="primary"
      >
        <Icon src={Check} size="16" mini />
        Mark page resolved
      </Button>
    </div>
    {#if data.filters.community}
      <ul class="font-normal flex flex-col gap-2 mt-2">
        <li>
          <span class="text-sm text-slate-600 dark:text-zinc-400">
            Community
          </span>
          {#await client().getCommunity({ id: data.filters.community })}
            <Spinner width={24} />
          {:then community}
            <a
              class="inline"
              aria-label=Remove
              href="?community="
            >
              <Icon src={XMark} size="16" micro class="inline" />
            </a>
            <CommunityLink
              class="w-max inline"
              community={community.community_view.community}
            />
          {/await}
        </li>
      </ul>
    {/if}
  {/snippet}
</Header>

<div class="flex flex-col gap-6">
  {#if batch.progress > 0}
    <ProgressBar progress={batch.progress} />
  {/if}

  {#if data.items?.value && data.items?.value.length > 0}
    <CommonList size="lg">
      {#each data.items.value ?? [] as item}
        <svelte:element
          this={item.length == 1 ? 'li' : 'div'}
          class="z-0 relative"
        >
          <Material
            rounding={item.length == 1 ? 'none' : '2xl'}
            color={item.length == 1 ? 'none' : 'distinct'}
            padding={item.length == 1 ? 'none' : 'md'}
            class={['space-y-2 w-full z-10 relative']}
          >
            <Report {item} />
          </Material>
          {#if item.length > 1}
            <Material
              padding="none"
              rounding="none"
              color="uniform"
              class="-mt-1 rounded-b-2xl w-[97%] h-6 opacity-70 left-1/2 -translate-x-1/2 -z-10 relative"
            ></Material>
          {/if}
        </svelte:element>
      {/each}
    </CommonList>
    <Fixate placement="bottom">
      <Pageination
        page={data.page}
        href={(current) => `?page=${current}`}
        hasMore={data.items.value.length >= 20}
      />
    </Fixate>
  {:else}
    <div class="h-full grid place-items-center">
      <Placeholder
        icon={ShieldCheck}
        title="No new reports"
        description="When submissions are reported, you can act on them here."
      />
    </div>
  {/if}

  <!-- Recent Moderation Actions -->
  {#if actionLogItems.length > 0}
    <div class="flex flex-col gap-3">
      <h2 class="text-lg font-semibold flex items-center gap-2">
        <Icon src={Clock} size="20" mini />
        Recent Moderation Actions
      </h2>
      <Material color="uniform" rounding="2xl" class="overflow-hidden">
        <div style="width:100%; overflow-x: auto;" class="table-container">
          <table class="table overflow-x-auto table-fixed relative w-full min-w-xl">
            <colgroup>
              <col style="width: 15%;" />
              <col style="width: 20%;" />
              <col style="width: 30%;" />
              <col style="width: 35%;" />
            </colgroup>
            <thead class="text-left">
              <tr class="rounded-t-lg overflow-hidden">
                <th>Date</th>
                <th>Moderator</th>
                <th>Type</th>
                <th>Reason / Target</th>
              </tr>
            </thead>
            <tbody class="text-sm divide-y divide-slate-200 dark:divide-zinc-800">
              {#each actionLogItems as action (action.id ?? action.timestamp)}
                <tr>
                  <td>
                    <span class="text-xs">
                      <RelativeDate date={new Date(action.timestamp)} />
                    </span>
                  </td>
                  <td>
                    {#if action.moderator}
                      <UserLink
                        showInstance={false}
                        avatar
                        avatarSize={18}
                        user={action.moderator}
                      />
                    {:else}
                      <span class="text-slate-400 dark:text-zinc-500 text-xs">Auto</span>
                    {/if}
                  </td>
                  <td>
                    <span class="text-xs font-medium text-slate-600 dark:text-zinc-400">
                      {action.actionName === 'Unknown' ? 'Mod Action' : action.actionName}
                    </span>
                  </td>
                  <td>
                    <span class="text-xs truncate block max-w-[250px]">
                      {#if action.content}
                        {action.content}
                      {:else if action.reason}
                        {action.reason}
                      {:else}
                        <span class="text-slate-400 dark:text-zinc-500">No details</span>
                      {/if}
                    </span>
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      </Material>
    </div>
  {/if}
</div>

<style>
  @reference '../../app.css';

  .table-container {
    border: 1px solid var(--color-slate-200);
    background-color: white;
    border-radius: var(--radius-2xl);

    @variant dark {
      background-color: var(--color-zinc-950);
      border: 1px solid var(--color-zinc-800);
    }
  }

  .table {
    table-layout: fixed !important;

    thead {
      tr {
        th {
          padding: calc(var(--spacing) * 3);
          border-bottom: 1px solid var(--color-slate-200);
          font-weight: var(--font-weight-medium);
          font-size: var(--text-sm);
          background-color: var(--color-slate-25);

          @variant dark {
            background-color: var(--color-zinc-900);
            border-bottom: 1px solid var(--color-zinc-800);
          }
        }
      }
    }
  }

  .table tbody {
    font-size: var(--text-xs);
    :global(tr td) {
      padding: calc(var(--spacing) * 1) calc(var(--spacing) * 3);
    }
  }
</style>

<script lang="ts">
  import { page } from '$app/state'
  import { PiefedClient } from '$lib/api/piefed/adapter'
  import { profile } from '$lib/app/auth'

  import { searchParam } from '$lib/app/util.svelte'
  import { Option, Select } from 'mono-svelte'
  import { type SelectProps } from 'mono-svelte/forms/select/Select.svelte'
  import {
    ChartBar,
    GlobeAmericas,
    Icon,
    MapPin,
    Newspaper,
    ShieldCheck,
  } from 'svelte-hero-icons/dist'

  interface Props extends SelectProps<string> {
    selected: string
    navigate?: boolean
    changeDefault?: boolean
    showLabel?: boolean
  }

  let {
    selected = $bindable(),
    navigate = true,
    showLabel = true,
    children,
    ...rest
  }: Props = $props()
</script>

<Select
  {...rest}
  bind:value={selected}
  onchange={() => {
    if (navigate) searchParam(page.url, 'type', selected, 'page', 'cursor')
  }}
>
  {#snippet customLabel()}
    {#if showLabel}
      <span class="flex items-center gap-1">
        <Icon src={GlobeAmericas} size="16" micro />
        Location
      </span>
    {/if}
  {/snippet}
  <Option value="All" icon={GlobeAmericas}>
    All
  </Option>
  {#if profile.client instanceof PiefedClient}
    <Option value="Popular" icon={ChartBar}>
      Popular
    </Option>
  {/if}
  <Option value="Local" icon={MapPin}>Local</Option>
  <Option
    value="Subscribed"
    disabled={profile.current?.jwt == undefined}
    icon={Newspaper}
  >
    Subscriptions
  </Option>
  <Option value="ModeratorView" disabled={!profile.isMod()} icon={ShieldCheck}>
    Moderator
  </Option>
  {@render children?.()}
</Select>

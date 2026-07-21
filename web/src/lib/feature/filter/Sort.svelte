<script lang="ts">
  import { page } from '$app/state'
  import { PiefedClient } from '$lib/api/piefed/adapter'
  import { profile } from '$lib/app/auth'

  import { searchParam } from '$lib/app/util.svelte'
  import { Option, Select } from 'mono-svelte'
  import { type SelectProps } from 'mono-svelte/forms/select/Select.svelte'
  import {
    ArrowTrendingDown,
    ArrowTrendingUp,
    Calendar,
    CalendarDays,
    ChartBar,
    ChatBubbleLeftRight,
    ChatBubbleOvalLeft,
    Clock,
    Fire,
    Icon,
    PlusCircle,
    Scale,
    Star,
    Sun,
    Trophy,
  } from 'svelte-hero-icons/dist'

  interface Props extends SelectProps<string> {
    selected: string
    navigate?: boolean
    changeDefault?: boolean
    class?: string
  }

  let {
    selected = $bindable(),
    navigate = true,
    class: clazz = '',
    ...rest
  }: Props = $props()

  let sort: string = $state(selected?.startsWith('Top') ? 'TopAll' : selected)
  const setSelected = () => (selected = sort)
</script>

<div class="flex flex-row {clazz}">
  <Select
    {...rest}
    class={selected?.startsWith('Top') ? 'rounded-r-none' : ''}
    bind:value={sort}
    onchange={() => {
      setSelected()
      if (navigate) searchParam(page.url, 'sort', selected, 'page', 'cursor')
    }}
  >
    {#snippet customLabel()}
      <span class="flex items-center gap-1">
        <Icon src={ChartBar} size="13" micro />
        Sort
      </span>
    {/snippet}
    <Option value="Active" icon={ArrowTrendingUp}>
      Active
    </Option>
    <Option value="Hot" icon={Fire}>
      Hot
    </Option>
    <Option value="Scaled" icon={Scale}>
      Scaled
    </Option>
    <Option value="TopAll" icon={Trophy}>
      Top
    </Option>
    <Option value="New" icon={Star}>
      New
    </Option>
    {#if !(profile.client instanceof PiefedClient)}
      <Option value="Old" icon={Clock}>
        Old
      </Option>
      <Option value="Controversial" icon={ArrowTrendingDown}>
        Controversial
      </Option>
      <Option value="MostComments" icon={ChatBubbleOvalLeft}>
        Most Comments
      </Option>
      <Option value="NewComments" icon={ChatBubbleLeftRight}>
        New Replies
      </Option>
    {/if}
  </Select>
  {#if selected?.startsWith('Top')}
    <Select
      class="border-l-0 rounded-l-none"
      bind:value={selected}
      onchange={() => {
        sort = 'TopAll'
        if (navigate) searchParam(page.url, 'sort', selected, 'page', 'cursor')
      }}
    >
      {#snippet customLabel()}
        <span class="flex items-center gap-1">
          <Icon src={Clock} size="15" micro />
          Period
        </span>
      {/snippet}
      <Option value="TopAll" icon={PlusCircle}>
        All time
      </Option>
      <Option value="TopYear" icon={Calendar}>
        Past year
      </Option>
      <Option value="TopNineMonths" icon={Calendar}>
        9 months
      </Option>
      <Option value="TopSixMonths" icon={Calendar}>
        6 months
      </Option>
      <Option value="TopThreeMonths" icon={Calendar}>
        3 months
      </Option>
      <Option value="TopMonth" icon={CalendarDays}>
        Past month
      </Option>
      <Option value="TopWeek" icon={CalendarDays}>
        Past week
      </Option>
      <Option value="TopDay" icon={Sun}>
        Past day
      </Option>
      <Option value="TopTwelveHour" icon={Clock}>
        12 hours
      </Option>
      <Option value="TopSixHour" icon={Clock}>
        6 hours
      </Option>
      <Option value="TopHour" icon={Clock}>
        Past hour
      </Option>
    </Select>
  {/if}
</div>

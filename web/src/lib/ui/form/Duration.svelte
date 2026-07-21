<script lang="ts">

  import { Option, Select, TextInput } from 'mono-svelte'

  const multipliers = {
    minute: 60,
    hour: 60 * 60,
    day: 60 * 60 * 24,
    month: 60 * 60 * 24 * 30,
    year: 60 * 60 * 24 * 365,
    permanent: -1,
  }

  type Duration = keyof typeof multipliers

  function calculateValue(number: number, duration: Duration): number {
    return number * multipliers[duration]
  }

  interface Props {
    value: number
    max?: number
    min?: number
  }

  let { value = $bindable(), max, min = 0 }: Props = $props()

  let number = $state<number>(0)
  let duration = $state<Duration>(value == -1 ? 'permanent' : 'day')

  $effect(() => {
    value = calculateValue(number, duration)
  })
</script>

<div class="flex flex-row items-end gap-2">
  {#if duration != 'permanent'}
    <TextInput
      {min}
      {max}
      type="number"
      label="Number"
      class="w-24"
      bind:value={number}
    ></TextInput>
  {/if}
  <Select
    baseClass="flex-1"
    label=Duration
    bind:value={duration}
  >
    <Option value="minute">Minutes</Option>
    <Option value="hour">Hours</Option>
    <Option value="day">Days</Option>
    <Option value="month">Months</Option>
    <Option value="year">Years</Option>
    <Option value="permanent">Permanent</Option>
  </Select>
</div>

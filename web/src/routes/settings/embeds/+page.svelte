<script lang="ts">

  import { settings } from '$lib/app/settings.svelte'
  import { DOMAIN_REGEX_FORMS } from '$lib/app/util.svelte'
  import { CommonList } from '$lib/ui/layout'
  import { Option, Select, TextInput } from 'mono-svelte'
  import { CursorArrowRays, VideoCamera } from 'svelte-hero-icons/dist'
  import Setting from '../Setting.svelte'
  import ToggleSetting from '../ToggleSetting.svelte'
</script>

<CommonList>
  <ToggleSetting
    icon={CursorArrowRays}
    title=Click to view
    description=Only load embeds once you click on them.
    bind:checked={settings.embeds.clickToView}
  />
  <Setting icon={VideoCamera}>
    {#snippet title()}
      <span>YouTube</span>
    {/snippet}
    {#snippet description()}
      <span>
        The website to embed YouTube content in.
      </span>
    {/snippet}
    <Select bind:value={settings.embeds.youtube}>
      <Option value="youtube">YouTube</Option>
      <Option value="invidious">Invidious</Option>
      <Option value="piped">Piped</Option>
    </Select>
  </Setting>
  {#if settings.embeds.youtube == 'invidious'}
    <Setting>
      {#snippet title()}
        <span>Invidious instance</span>
      {/snippet}
      {#snippet description()}
        <span>
          The instance to embed YouTube content in.
        </span>
      {/snippet}
      <TextInput
        label=Invidious instance
        pattern={DOMAIN_REGEX_FORMS}
        bind:value={settings.embeds.invidious}
      />
    </Setting>
  {/if}
  {#if settings.embeds.youtube == 'piped'}
    <Setting>
      {#snippet title()}
        <span>Piped instance</span>
      {/snippet}
      {#snippet description()}
        <span>
          The instance to embed YouTube content in.
        </span>
      {/snippet}
      <TextInput
        label=Piped instance
        pattern={DOMAIN_REGEX_FORMS}
        bind:value={settings.embeds.piped}
      />
    </Setting>
  {/if}
</CommonList>

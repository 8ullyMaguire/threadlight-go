<script lang="ts">
  import { env } from '$env/dynamic/public'

  import { settings } from '$lib/app/settings.svelte'
  import Sort from '$lib/feature/filter/Sort.svelte'
  import ViewSelect from '$lib/feature/filter/ViewSelect.svelte'
  import Link from '$lib/ui/form/Link.svelte'
  import Switch from '$lib/ui/form/Switch.svelte'
  import { CommonList } from '$lib/ui/layout'
  import { Button, Option, Select } from 'mono-svelte'
  import {
    ArrowsPointingOut,
    ArrowsRightLeft,
    ArrowsUpDown,
    ArrowTopRightOnSquare,
    ArrowTrendingDown,
    Bars2,
    BarsArrowDown,
    Calendar,
    ChartBar,
    ChatBubbleOvalLeftEllipsis,
    Clock,
    CubeTransparent,
    DocumentText,
    EyeSlash,
    Fire,
    GlobeAmericas,
    Heart,
    Icon,
    Language,
    Photo,
    Plus,
    Star,
    TableCells,
    Tag,
    Trophy,
    ViewColumns,
  } from 'svelte-hero-icons/dist'
  import Setting from '../Setting.svelte'
  import ToggleSetting from '../ToggleSetting.svelte'
  import Filter from './Filter.svelte'

  let localeMap: Map<
    string,
    {
      name: string
    }
  > = new Map([
    ['en', { name: 'English' }],
    ['ar', { name: 'العربية' }],
    ['he', { name: 'עברית' }],
    ['bg', { name: 'български' }],
    ['de', { name: 'Deutsch' }],
    ['es', { name: 'Español' }],
    ['et', { name: 'eesti keel' }],
    ['fi', { name: 'suomi' }],
    ['fr', { name: 'Français' }],
    ['hu', { name: 'Magyar' }],
    ['ja', { name: '日本語' }],
    ['nl', { name: 'Nederlands' }],
    ['pl', { name: 'Polski' }],
    ['pt', { name: 'Português (PT)' }],
    ['pt-BR', { name: 'Português (BR)' }],
    ['tr', { name: 'Türkçe' }],
    ['ru', { name: 'Русский' }],
    ['zh-Hans', { name: '简体中文' }],
    ['zh-Hant', { name: '繁體中文' }],
  ])
</script>

<CommonList>
  {#if env.PUBLIC_XYLIGHT_MODE}
    <Setting icon={Heart}>
      {#snippet title()}
        <span class="dark:text-pink-400 text-pink-600">
          Donate
        </span>
      {/snippet}
      {#snippet description()}
        Photon will always be free and open-source, and is continously developed and designed to be helpful. If you like Photon and want to financially support its development, you can donate here.
      {/snippet}
      <Button
        color="none"
        class="bg-linear-to-r ml-6 dark:from-pink-400 dark:to-fuchsia-400 from-pink-600 to-red-600 text-white dark:text-black"
        href="https://buymeacoffee.com/xylight"
        target="_blank"
        rounding="xl"
        icon={Heart}
      >
        Donate
      </Button>
    </Setting>
  {/if}
  <Setting icon={Language}>
    {#snippet title()}
      Language
    {/snippet}
    {#snippet description()}
      <p>
        The language used for text in the UI.
        <Link href="/translators" highlight class="text-base font-semibold">
          Credits
        </Link>
      </p>
    {/snippet}
    <!--@ts-ignore-->
    <Select bind:value={settings.language}>
      <Option icon={Language} value={null}>
        Auto
      </Option>
      {#each localeMap.entries() as [key, value]}
        <Option
          data-label={key == 'placeholder'}
          disabled={key == 'placeholder'}
          value={key}
        >
          {value.name}
        </Option>
      {/each}
    </Select>
  </Setting>
  {#if $locale == 'he' || $locale == 'ar'}
    <ToggleSetting
      bind:checked={settings.useRtl}
      title=Use RTL layout
    ></ToggleSetting>
  {/if}
  <Setting icon={ViewColumns}>
    {#snippet title()}
      <span>Post style</span>
    {/snippet}
    <ViewSelect showLabel={false} />
    {#snippet description()}
      <p>
        {#if settings.view == 'cozy'}
          Show posts with large images, rich embeds, and longer post bodies.
        {:else if settings.view == 'compact'}
          Show posts in a list, without post bodies and with tighter spacing.
        {/if}
      </p>
    {/snippet}
  </Setting>
  <Setting optionClass="flex-2 max-w-full flex-wrap min-w-0 " icon={ChartBar}>
    {#snippet title()}
      <span>Default sort</span>
    {/snippet}
    {#snippet description()}
      <span>Changes the default sort type used in feeds.</span>
    {/snippet}
    <div
      class="flex flex-row flex-wrap
      flex-1 gap-2 w-full lg:w-max max-w-full lg:self-end"
    >
      <Select bind:value={settings.defaultSort.feed}>
        {#snippet customLabel()}
          <div class="flex items-center gap-1">
            <Icon src={GlobeAmericas} size="16" mini />
            Location
          </div>
        {/snippet}
        <Option value="All">All</Option>
        <Option value="Local">Local</Option>
        <Option value="Subscribed">
          Subscriptions
        </Option>
        <Option value="Moderator">
          Moderator
        </Option>
      </Select>
      <Sort bind:selected={settings.defaultSort.sort} navigate={false} />
      <Select bind:value={settings.defaultSort.comments}>
        {#snippet customLabel()}
          <div class="flex items-center gap-1">
            <Icon src={ChatBubbleOvalLeftEllipsis} size="14" mini />
            Comments
          </div>
        {/snippet}

        <Option icon={Fire} value="Hot">Hot</Option>
        <Option icon={Trophy} value="Top">
          Top
        </Option>
        <Option icon={Star} value="New">New</Option>
        <Option icon={Clock} value="Old">Old</Option>
        <Option icon={ArrowTrendingDown} value="Controversial">
          Controversial
        </Option>
      </Select>
    </div>
  </Setting>
  <ToggleSetting
    icon={BarsArrowDown}
    bind:checked={settings.infiniteScroll}
    title=Infinite scroll
    description=Loads new posts as you scroll and appends them to the feed.
  />
  <Setting icon={EyeSlash} adaptive={false} mainClass="rounded-b-none!">
    {#snippet title()}
      <span>Content filters</span>
    {/snippet}
    {#snippet description()}
      <span>Hide or minimize posts matching given strings. Supports regular expressions.</span>
    {/snippet}
  </Setting>
  {#each settings.filters as filter, index}
    <Filter onremove={() => settings.filters.splice(index, 1)} bind:filter />
  {/each}
  <li class="rounded-t-none! -mt-1! border-t-0!">
    <Button
      color="none"
      class="w-full p-2"
      icon={Plus}
      onclick={() =>
        settings.filters.push({
          match: 'New filter.',
          action: 'minimize',
          type: 'keyword',
        })}
    >
      Add
    </Button>
  </li>
  <Setting icon={Photo}>
    {#snippet title()}
      <span>Thumbnail alignment</span>
    {/snippet}
    {#snippet description()}
      <span>
        Changes where thumbnails for posts are displayed in compact mode.
      </span>
    {/snippet}
    <Switch
      options={[true, false]}
      optionNames={['Left','Right',
      ]}
      bind:selected={settings.leftAlign}
    />
  </Setting>
  <ToggleSetting
    icon={ArrowsRightLeft}
    bind:checked={settings.posts.reverseActions}
    title=Reverse action row
    description=Reverses the order of actions on posts and comments.
  />
  <ToggleSetting
    icon={TableCells}
    supportedPlatforms={{ desktop: true, tablet: false, mobile: false }}
    bind:checked={settings.newWidth}
    title=Limit layout width
    description=Improve readability by limiting the main content width.
  />
  <ToggleSetting
    icon={Calendar}
    bind:checked={settings.absoluteDates}
    title=Absolute timestamps
    description=Replaces relative timestamps (3 hours ago) with absolute ones (15/6/25).
  />
  <ToggleSetting
    icon={ArrowsUpDown}
    bind:checked={settings.voteRatioBar}
    title=Vote ratio bar
    description=When the controversiality of a post reaches greater than 15%, a bar will be displayed in the vote buttons to help visualize sentiment.
  />
  <ToggleSetting
    icon={CubeTransparent}
    supportedPlatforms={{ desktop: false, tablet: false, mobile: true }}
    bind:checked={settings.dock.autoHide}
    title=Auto hide dock
    description=Hides the dock on mobile when you scroll down, re-appears when you scroll up.
  />
  <ToggleSetting
    icon={ArrowTopRightOnSquare}
    bind:checked={settings.openLinksInNewTab}
    title=Open posts in new tab
    description=Opens posts in a separate tab instead of the current.
  />
  <Setting icon={DocumentText}>
    {#snippet title()}
      <span>Font</span>
    {/snippet}
    {#snippet description()}
      <span>The font used for text in the UI.</span>
    {/snippet}
    <Select bind:value={settings.font}>
      <Option value="inter">Inter</Option>
      <Option value="system">System UI</Option>
      <Option value="browser">Browser</Option>
      <Option value="serifs">Roboto Slab</Option>
    </Select>
  </Setting>
  <ToggleSetting
    icon={ArrowsPointingOut}
    bind:checked={settings.expandImages}
    title=Expandable thumbnails
    description=Clicking on a post's image brings you to an expanded view rather than the post page.
  />
  <ToggleSetting
    icon={Bars2}
    bind:checked={settings.posts.deduplicateEmbed}
    title=Hide duplicate embed content
    description=Hides the post title and body if they are identical to the embed.
  />
  <ToggleSetting
    icon={ArrowTopRightOnSquare}
    bind:checked={settings.posts.titleOpensUrl}
    title=Title opens URL
    description=Makes clicking the title open the post's attached URL, rather than the comments.
  />
  <ToggleSetting
    icon={Tag}
    title=Parse tags in title
    description=Take text in [brackets] and turn it into tags that you can click to search for.
    bind:checked={settings.parseTags}
  />
</CommonList>

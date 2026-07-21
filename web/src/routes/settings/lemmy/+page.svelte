<script lang="ts">

  import { settings } from '$lib/app/settings.svelte'
  import { CommonList } from '$lib/ui/layout'
  import { Switch } from 'mono-svelte'
  import {
    Bars3,
    BookOpen,
    CubeTransparent,
    ExclamationTriangle,
    Eye,
    EyeSlash,
    ServerStack,
    Share,
    User,
  } from 'svelte-hero-icons/dist'
  import Setting from '../Setting.svelte'
  import ToggleSetting from '../ToggleSetting.svelte'
</script>

<CommonList>
  <ToggleSetting
    icon={Eye}
    bind:checked={settings.posts.showHidden}
    title=Show hidden posts
    description=Shows the posts you've manually hidden.
  />
  <ToggleSetting
    icon={Bars3}
    bind:checked={settings.posts.compactFeatured}
    title=Compact featured posts
    description=Makes featured posts more compact.
  />
  <ToggleSetting
    icon={BookOpen}
    bind:checked={settings.markPostsAsRead}
    title=Mark posts as read automatically
    description=Marks posts as read when you visit them.
  />
  <ToggleSetting
    icon={CubeTransparent}
    bind:checked={settings.markReadPosts}
    title=Fade read posts
    description=Fade the title of posts you've already read
  />
  <ToggleSetting
    icon={Share}
    bind:checked={settings.crosspostOriginalLink}
    title=Mark crossposts
    description=Insert 'crossposted from' and a link when crossposting.
  />
  <Setting icon={EyeSlash}>
    {#snippet title()}
      <span>Hide submissions</span>
    {/snippet}
    {#snippet description()}
      <span>
        <p>Hide certain types of submissions.</p>
      </span>
    {/snippet}
    <div class="flex flex-col items-start gap-4 flex-wrap">
      <Switch bind:checked={settings.hidePosts.deleted}>
        Deleted by user
      </Switch>
      <Switch bind:checked={settings.hidePosts.removed}>
        Removed by moderator
      </Switch>
    </div>
  </Setting>
  <ToggleSetting
    icon={ExclamationTriangle}
    bind:checked={settings.nsfwBlur}
    title=NSFW blur
    description=Blurs images and remove post bodies of NSFW posts.
  />
  <Setting icon={ServerStack}>
    {#snippet title()}
      Show server identifiers
    {/snippet}
    {#snippet description()}
      Shows the servers where certain content originated from.
    {/snippet}
    <div class="flex flex-row flex-wrap items-center gap-4">
      <Switch bind:checked={settings.showInstances.user}>
        Users
      </Switch>
      <Switch bind:checked={settings.showInstances.comments}>
        Comments
      </Switch>
      <Switch bind:checked={settings.showInstances.community}>
        Communities
      </Switch>
    </div>
  </Setting>
  <ToggleSetting
    icon={User}
    bind:checked={settings.displayNames}
    title=Use display name
    description=Show users' display names instead of their actual usernames.
  />
</CommonList>

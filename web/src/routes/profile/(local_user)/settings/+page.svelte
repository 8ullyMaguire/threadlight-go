<script lang="ts">
  import { getClient, site } from '$lib/api/client.svelte'
  import type { SaveUserSettings } from '$lib/api/types'
  import { profile } from '$lib/app/auth'

  import MarkdownEditor from '$lib/app/markdown/MarkdownEditor.svelte'
  import ImageInputUpload from '$lib/ui/form/ImageInputUpload.svelte'
  import { Header } from '$lib/ui/layout'
  import {
    Badge,
    Button,
    Label,
    Material,
    Menu,
    MenuButton,
    Switch,
    TextInput,
    toast,
  } from 'mono-svelte'
  import { Icon, Plus } from 'svelte-hero-icons/dist'
  import type { PageData } from './$types'

  interface Props {
    inline?: boolean
    data: PageData
    children?: import('svelte').Snippet
  }

  let { inline = false, data, children }: Props = $props()

  let formData: Omit<SaveUserSettings, 'auth'> | undefined = $state({
    ...data.my_user?.local_user_view?.local_user,
    ...data.my_user?.local_user_view?.person,
    discussion_languages: data.my_user?.discussion_languages,
  })

  async function save() {
    if (!formData || !profile.current?.jwt) return

    loading = true

    try {
      await getClient().saveUserSettings({
        ...formData,
      })

      toast({
        content:'Saved your user settings. If you changed your email, a verification link was sent.',
        type: 'success',
      })
    } catch (err) {
      toast({
        content: err as string,
        type: 'error',
      })
    }
    loading = false
  }

  let loading = $state(false)
</script>

<form
  class="flex flex-col gap-4 h-full"
  onsubmit={(e) => {
    e.preventDefault()
    save()
  }}
>
  {#if !inline}
    <Header pageHeader>Edit</Header>
  {/if}
  {@render children?.()}
  {#if data.my_user?.local_user_view?.local_user && formData}
    <TextInput label=Email bind:value={formData.email} />
    <TextInput
      label=Display name
      bind:value={formData.display_name}
      placeholder="Optional"
    />
    <MarkdownEditor
      images={false}
      bind:value={() => formData.bio ?? '', (v) => (formData.bio = v)}
      label=Bio
      previewButton
    />
    <div class="flex gap-2 items-center *:flex-1">
      <ImageInputUpload
        label=Avatar
        bind:imageUrl={formData.avatar}
      />
      <ImageInputUpload
        label=Banner
        bind:imageUrl={formData.banner}
      />
    </div>
    <TextInput
      label=Chat on Matrix
      bind:value={formData.matrix_user_id}
      placeholder="@user:example.com"
    />
    <Switch bind:checked={formData.show_nsfw}>
      Show NSFW content
    </Switch>
    <Switch bind:checked={formData.bot_account}>
      Bot account
    </Switch>
    <Switch bind:checked={formData.show_bot_accounts}>
      Show bots
    </Switch>
    <Switch bind:checked={formData.show_read_posts}>
      Show read posts
    </Switch>
    <div class="space-y-1">
      <Label id="languages">
        Languages
      </Label>
      <p>If any are set, you will only see content explicitly tagged as that language. Most content, untagged, falls under 'Undetermined'.</p>
      <Material rounding="xl" color="uniform" class="dark:bg-zinc-950">
        {#if site.data && formData.discussion_languages}
          <div class="flex gap-2 flex-wrap flex-row">
            <Menu class="gap-px">
              {#snippet target(attachment)}
                <button {@attach attachment} type="button">
                  <Badge color="blue-subtle">
                    <Icon src={Plus} micro size="14" />
                    Add
                  </Badge>
                </button>
              {/snippet}
              {#each site.data.all_languages.filter((l) => !formData.discussion_languages?.includes(l.id)) as language (language.id)}
                <MenuButton
                  class="min-h-[16px] py-0"
                  onclick={() => {
                    formData.discussion_languages?.push(language.id)
                  }}
                >
                  {language.name}
                </MenuButton>
              {/each}
            </Menu>
            {#each formData.discussion_languages as languageId, index (languageId)}
              {@const language = site.data.all_languages.find(
                (l) => l.id == languageId,
              )}
              <button
                type="button"
                class="hover:brightness-150 transition-all"
                onclick={() => {
                  formData.discussion_languages?.splice(index, 1)
                }}
              >
                <Badge class="cursor-pointer">{language?.name}</Badge>
              </button>
            {/each}
          </div>
        {/if}
      </Material>
    </div>

    <Button
      submit
      size="lg"
      color="primary"
      class="mt-auto"
      {loading}
      disabled={loading}
    >
      Save
    </Button>
  {/if}
</form>

<script lang="ts">
  import { client, site } from '$lib/api/client.svelte'
  import type { EditSite } from '$lib/api/types'
  import { profile } from '$lib/app/auth'
  import { errorMessage } from '$lib/app/error'

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
    Option,
    Select,
    Switch,
    TextInput,
    toast,
  } from 'mono-svelte'
  import { Icon, Plus } from 'svelte-hero-icons/dist'

  let { data: pageData } = $props()
  let data = $state(pageData)

  const formData: Omit<EditSite, 'auth'> | undefined = $state(
    data.site
      ? {
          ...data.site.site_view.local_site,
          ...data.site.site_view.site,
          discussion_languages: data.site.discussion_languages,
        }
      : undefined,
  )

  async function save() {
    if (!profile.current?.jwt) return

    saving = true

    try {
      await client().editSite({
        ...formData,
      })
      toast({
        content:'Updated your site.',
        type: 'success',
      })
    } catch (err) {
      toast({
        content: errorMessage(err as string),
        type: 'error',
      })
    }

    saving = false
  }

  let saving = $state(false)
</script>

<svelte:head>
  <title>Administration</title>
</svelte:head>

<form
  class="flex flex-col gap-4"
  onsubmit={(e) => {
    e.preventDefault()
    save()
  }}
>
  <Header pageHeader>Configuration</Header>
  {#if formData}
    <TextInput
      bind:value={() => formData.name ?? '', (v) => (formData.description = v)}
      label=Name
    />
    <TextInput
      bind:value={
        () => formData.description ?? '', (v) => (formData.description = v)
      }
      label=Description
    />
    <MarkdownEditor
      previewButton
      bind:value={() => formData.sidebar ?? '', (v) => (formData.sidebar = v)}
      label=Sidebar
    />
    <MarkdownEditor
      previewButton
      bind:value={
        () => formData.legal_information ?? '',
        (v) => (formData.legal_information = v)
      }
      label=Legal
    />
    <ImageInputUpload
      label=Icon
      bind:imageUrl={formData.icon}
    />
    <ImageInputUpload
      label=Banner
      bind:imageUrl={formData.banner}
    />
    <Switch bind:checked={formData.enable_downvotes}>
      Enable downvotes
    </Switch>
    <Switch bind:checked={formData.enable_nsfw}>
      Enable NSFW
    </Switch>
    <Select
      label=Registration mode
      bind:value={formData.registration_mode}
      class="w-max"
    >
      <Option value="Open">
        Open
      </Option>
      <Option value="RequireApplication">
        Require application
      </Option>
      <Option value="Closed">
        Closed
      </Option>
    </Select>
    {#if formData.registration_mode == 'RequireApplication'}
      <MarkdownEditor
        previewButton
        label=Application question
        bind:value={formData.application_question}
      />
    {/if}
    <Switch bind:checked={formData.community_creation_admin_only}>
      Only admins can create communities
    </Switch>
    <Switch bind:checked={formData.require_email_verification}>
      Require email verification
    </Switch>
    <Switch bind:checked={formData.application_email_admins}>
      Email admins on receiving new applications
    </Switch>
    <Switch bind:checked={formData.reports_email_admins}>
      Email admins on receiving new reports
    </Switch>
    <Select
      label=Listing Type
      bind:value={formData.default_post_listing_type}
      class="w-max"
    >
      <Option value="All">All</Option>
      <Option value="Local">
        Local
      </Option>
    </Select>
    <Switch bind:checked={formData.private_instance}>
      Private instance
    </Switch>
    <Switch bind:checked={formData.hide_modlog_mod_names}>
      Hide modlog mod names
    </Switch>
    <TextInput
      bind:value={formData.slur_filter_regex}
      label=Slur filter regex
      placeholder="(word1|word2)"
    />
    <Switch bind:checked={formData.federation_enabled}>
      Federation enabled
    </Switch>
    <Switch bind:checked={formData.federation_debug}>
      Federation debug mode
    </Switch>
    <Switch bind:checked={formData.captcha_enabled}>
      CAPTCHA enabled
    </Switch>

    <div class="space-y-1">
      <Label>Languages</Label>
      <p>If any are set, you will only see content explicitly tagged as that language. Most content, untagged, falls under 'Undetermined'.</p>
      <Material rounding="xl" color="uniform" class="dark:bg-zinc-950">
        {#if site.data}
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
                  class="min-h-4 py-0"
                  onclick={() => {
                    formData.discussion_languages?.push(language.id)
                  }}
                >
                  {language.name}
                </MenuButton>
              {/each}
            </Menu>
            {#each formData.discussion_languages ?? [] as languageId, index (languageId)}
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
  {/if}
  <Button color="primary" size="lg" loading={saving} disabled={saving} submit>
    Save
  </Button>
</form>

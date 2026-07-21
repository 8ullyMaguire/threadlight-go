<script lang="ts">

  import { defaultSettings, settings } from '$lib/app/settings.svelte'
  import { Header } from '$lib/ui/layout'
  import TabbedLayoutShell from '$lib/ui/layout/pages/TabbedLayoutShell.svelte'
  import { action, Button, Modal, modal, TextArea, toast } from 'mono-svelte'
  import {
      ArrowDownTray,
      ArrowPath,
      ArrowUpTray,
      AtSymbol,
      Cog,
      EllipsisHorizontal,
      Heart,
      ListBullet,
      RectangleStack,
      ShieldCheck,
  } from 'svelte-hero-icons/dist'

  let { children } = $props()
  let importing = $state(false)
  let importText = $state('')
</script>

{#if importing}
  <Modal
    bind:open={importing}
    onaction={() => {
      try {
        if (importText == '') {
          throw new Error('import failed')
        }
        const parsed = JSON.parse(importText)
        const merged = { ...defaultSettings, ...parsed }

        Object.assign(settings, merged)

        toast({ content:'Successfully imported settings', type: 'success' })
        importing = false
      } catch (err) {
        toast({ content: err as string, type: 'error' })
      }
    }}
    title=Import
    action=Import
  >
    <TextArea bind:value={importText} style="font-family: monospace;" />
  </Modal>
{/if}

<svelte:head>
  <title>Settings</title>
</svelte:head>

<!--TODO use route ID instead of pathname-->
<TabbedLayoutShell
  baseRoute="/settings"
  routes={[
    {
      href: '/settings/app',
      name:'App',
      icon: Cog,
    },
    {
      href: '/settings/lemmy',
      name:'Lemmy',
      icon: AtSymbol,
    },
    {
      href: '/settings/embeds',
      name:'Embeds',
      icon: RectangleStack,
    },
    {
      href: '/settings/lists',
      name:'Lists',
      icon: ListBullet,
    },
    {
      href: '/settings/affinity',
      name:'Affinity',
      icon: Heart,
    },
    {
      href: '/settings/moderation',
      name:'Moderation',
      icon: ShieldCheck,
    },
    {
      href: '/settings/other',
      name:'Other',
      icon: EllipsisHorizontal,
    },
  ]}
>
  {#snippet header()}
    <Header pageHeader class="text-3xl font-bold flex justify-between">
      Settings
      {#snippet extended()}
        <div class="flex items-center tracking-normal gap-2">
          <Button
            onclick={() => {
              importText = ''
              importing = true
            }}
            icon={ArrowDownTray}
            size="lg"
          >
            Import
          </Button>
          <Button
            onclick={() => {
              const json = JSON.stringify(settings)
              navigator?.clipboard?.writeText?.(json)
              toast({ content:'Copied to clipboard.' })
            }}
            icon={ArrowUpTray}
            size="lg"
          >
            Export
          </Button>
          <Button
            onclick={() => {
              modal({
                title:'Reset',
                body:'Do you want to reset your settings to the default?',
                actions: [
                  action({
                    action: () => Object.assign(settings, defaultSettings),
                    close: true,
                    type: 'danger',
                    content:'Reset',
                  }),
                  action({
                    content:'Cancel',
                  }),
                ],
              })
            }}
            icon={ArrowPath}
            size="lg"
          >
            Reset
          </Button>
        </div>
      {/snippet}
    </Header>
  {/snippet}

  {@render children?.()}
</TabbedLayoutShell>

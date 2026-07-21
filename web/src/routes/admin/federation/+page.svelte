<script lang="ts">
  import { client, getClient } from '$lib/api/client.svelte'
  import type { Instance } from '$lib/api/types'
  import { profile } from '$lib/app/auth'
  import { errorMessage } from '$lib/app/error'

  import Placeholder from '$lib/ui/info/Placeholder.svelte'
  import { Header } from '$lib/ui/layout'
  import { publishedToDate } from '$lib/ui/util/date'
  import {
    Button,
    FileInput,
    Material,
    Popover,
    TextInput,
    toast,
  } from 'mono-svelte'
  import RelativeDate from 'mono-svelte/util/RelativeDate.svelte'
  import {
    Check,
    ExclamationTriangle,
    Icon,
    Plus,
    XMark,
  } from 'svelte-hero-icons/dist'
  import { flip } from 'svelte/animate'
  import { expoOut } from 'svelte/easing'
  import { preventDefault } from 'svelte/legacy'

  let { data = $bindable() } = $props()

  let blockInstance = $state({
      instance: '',
      loading: false,
    }),
    allowInstance = $state({
      instance: '',
      loading: false,
    }),
    saving = $state(false)

  async function moderateInstance(instance: string, blocked: boolean) {
    if (instance == '') return
    try {
      if (
        !profile.current?.jwt ||
        !data.instances.value?.blocked ||
        !data.instances.value?.allowed
      )
        return

      if (blocked) blockInstance.loading = true
      else allowInstance.loading = true

      const blockedInstances = data.instances.value.blocked.map((i) => i.domain)

      const allowedInstances = data.instances.value.allowed.map((i) => i.domain)

      if (blocked) blockedInstances.push(instance)
      if (!blocked) allowedInstances.push(instance)

      await getClient().editSite({
        blocked_instances: blockedInstances,
        allowed_instances: allowedInstances,
      })

      const instances = await getClient().getFederatedInstances()

      // @ts-expect-error goofy typing thing that i shouldnt have done
      data.instances.value = instances.federated_instances

      blockInstance.instance = ''
    } catch (err) {
      toast({
        content: errorMessage(err as string),
        type: 'error',
      })
    }

    blockInstance.loading = false
    allowInstance.loading = false
  }

  async function save() {
    try {
      if (!profile.current?.jwt || !data.instances.value?.blocked) return

      saving = true

      const blockedInstances = data.instances.value.blocked.map((i) => i.domain)
      const allowedInstances = data.instances.value.allowed?.map(
        (i) => i.domain,
      )

      await client().editSite({
        allowed_instances: allowedInstances,
        blocked_instances: blockedInstances,
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

  let csv = $state<FileList | null>()

  async function parseCsv(files: FileList) {
    if (!(files.length > 0)) return

    const reader = new FileReader()

    reader.onload = (e) => {
      if (!data.instances.value?.blocked) throw new Error('Missing instance')
      const content = e.target?.result
      if (!content) toast({ content:'Failed to read CSV', type: 'warning' })

      try {
        const instances: Instance[] = []
        const str = content?.toString()
        if (!str) throw new Error('No content')

        const lines = str.split(/\r?\n/).slice(1)

        for (const line of lines) {
          if (line == '') continue
          const domain = line.split(',')[0]

          const item: Instance = {
            domain: domain,
            id: Math.floor(Math.random() * 1000000),
            published: new Date().toISOString(),
          }

          instances.push(item)
        }

        data.instances.value.blocked = instances
      } catch {
        toast({ content:'Failed to read CSV', type: 'error' })
      }
    }

    reader.onerror = () =>
      toast({ content:'Failed to read CSV', type: 'error' })

    reader.readAsText(files[0])
  }

  $effect(() => {
    if (csv) parseCsv(csv)
  })
</script>

<svelte:head>
  <title>Federation</title>
</svelte:head>

<Header pageHeader class="font-bold text-2xl flex items-center justify-between">
  Federation
  <Button color="primary" onclick={save} loading={saving} disabled={saving}>
    Save
  </Button>
</Header>
{#if data.site && data.instances.value?.blocked}
  <FileInput preview={false} bind:files={csv}>
    {#snippet button()}
      <Button class="w-max" icon={Plus}>
        Upload CSV
      </Button>
    {/snippet}
    {#snippet choose()}{/snippet}
  </FileInput>
  <div class="flex flex-col md:flex-row gap-4">
    <div class="flex-1 w-full max-h-168 h-full flex flex-col gap-2">
      <h2 class="font-bold text-lg">Blocked servers</h2>
      <form
        onsubmit={preventDefault(() =>
          moderateInstance(blockInstance.instance, true),
        )}
        class="flex flex-row gap-2"
      >
        <TextInput
          placeholder=Block a server...
          class="flex-1"
          bind:value={blockInstance.instance}
        />
        <Button
          submit
          loading={blockInstance.loading}
          disabled={blockInstance.loading}
        >
          Block
        </Button>
      </form>
      <Material class="h-full overflow-auto" color="uniform" rounding="2xl">
        <ul
          class="*:py-3 dark:divide-zinc-800! {allowInstance.instance != ''
            ? 'opacity-50'
            : ''}"
        >
          {#if data.instances.value.blocked.length > 0}
            {#each data.instances.value.blocked.toSorted( (b, a) => b.domain.localeCompare(a.domain), ) as instance (instance.id)}
              <div
                animate:flip={{ duration: 300, easing: expoOut }}
                class="flex justify-between items-center first:pt-0 last:pb-0"
              >
                <div class="flex flex-col">
                  <span class="font-medium">{instance.domain}</span>
                  <span
                    class="text-xs text-slate-600 dark:text-zinc-400 capitalize"
                  >
                    {instance.software ?? 'Unknown'} • <RelativeDate
                      date={publishedToDate(instance.published)}
                    />
                  </span>
                </div>
                <Button
                  size="square-md"
                  onclick={() => {
                    if (!data.instances.value?.blocked) return

                    data.instances.value.blocked =
                      data.instances.value?.blocked.filter(
                        (i) => i.id != instance.id,
                      )
                  }}
                  icon={XMark}
                ></Button>
              </div>
            {/each}
          {:else}
            <Placeholder
              icon={Check}
              title=No blocked servers
              description=Blocking an server will hide new submissions from that server.
            />
          {/if}
        </ul>
      </Material>
    </div>
    <div class="md:flex-1 w-full max-h-168 flex flex-col gap-2">
      <h2 class="font-bold text-lg flex items-center space-x-1">
        <span>Whitelisted servers</span>
        {#if allowInstance.instance || !(data.instances.value.allowed?.length == 0)}
          <Popover openOnHover placement="bottom-end">
            {#snippet target(attachment)}
              <Icon
                {@attach attachment}
                src={ExclamationTriangle}
                solid
                class="text-yellow-500"
                size="20"
              />
            {/snippet}
            <p class="font-normal">
              Whitelisting an server will block every other server.
            </p>
          </Popover>
        {/if}
      </h2>
      <form
        onsubmit={preventDefault(() =>
          moderateInstance(allowInstance.instance, false),
        )}
        class="flex flex-row gap-2"
      >
        <TextInput
          placeholder=Whitelist a server
          class="flex-1"
          bind:value={allowInstance.instance}
        />
        <Button
          submit
          loading={allowInstance.loading}
          disabled={allowInstance.loading}
        >
          Whitelist
        </Button>
      </form>
      <Material class="h-full overflow-auto" color="uniform" rounding="2xl">
        <ul class="*:py-3 dark:divide-zinc-800!">
          {#if data.instances.value.allowed && (data?.instances?.value?.allowed?.length ?? 0) > 0}
            {#each data.instances.value.allowed.toSorted( (b, a) => b.domain.localeCompare(a.domain), ) as instance (instance.id)}
              <div
                animate:flip={{ duration: 300, easing: expoOut }}
                class="flex justify-between items-center first:pt-0 last:pb-0"
              >
                <div class="flex flex-col">
                  <span class="font-medium">{instance.domain}</span>
                  <span
                    class="text-xs text-slate-600 dark:text-zinc-400 capitalize"
                  >
                    {instance.software ?? 'Unknown'} • <RelativeDate
                      date={publishedToDate(instance.published)}
                    />
                  </span>
                </div>
                <Button
                  size="square-md"
                  onclick={() => {
                    if (!data.instances.value?.allowed) return

                    data.instances.value.allowed =
                      data.instances.value?.allowed.filter(
                        (i) => i.id != instance.id,
                      )
                  }}
                  icon={XMark}
                ></Button>
              </div>
            {/each}
          {:else}
            <div class="my-auto h-max">
              <Placeholder
                icon={Check}
                title=No whitelisted servers
                description={"Whitelisting an server will block every other server."}
              />
            </div>
          {/if}
        </ul>
      </Material>
    </div>
  </div>
{/if}

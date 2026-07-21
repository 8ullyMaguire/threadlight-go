<script lang="ts">
  import { client } from '$lib/api/client.svelte'
  import type { Community, Person } from '$lib/api/types'
  import { profile } from '$lib/app/auth'
  import { errorMessage } from '$lib/app/error'

  import MarkdownEditor from '$lib/app/markdown/MarkdownEditor.svelte'
  import Duration from '$lib/ui/form/Duration.svelte'
  import Avatar from '$lib/ui/generic/Avatar.svelte'
  import { Button, Label, Modal, Switch, toast } from 'mono-svelte'
  import CommunityLink from '../community/CommunityLink.svelte'

  interface Props {
    open?: boolean
    user: Person | undefined
    community: Community | undefined
    banned: boolean
  }

  let {
    open = $bindable(false),
    user: item = $bindable(),
    community,
    banned,
  }: Props = $props()

  let reason = $state('')
  let deleteData = $state(false)
  let expires = $state(-1)
  let loading = $state(false)

  // hack due to svelte's reactive declarations
  const resetReason = () => {
    reason = ''
    deleteData = false
    expires = -1
  }

  $effect(() => {
    if (item) resetReason()
  })

  async function submit() {
    if (!item || !profile.current?.user || !profile.current?.jwt) return

    loading = true

    try {
      let date: number | undefined

      if (expires > 0) {
        date = Math.floor(Date.now() / 1000) + expires
      }

      if (community) {
        await client().banFromCommunity({
          ban: !banned,
          community_id: community.id,
          person_id: item.id,
          reason: reason || undefined,
          remove_data: deleteData,
          expires: date,
        })
      } else {
        await client().banPerson({
          ban: !banned,
          person_id: item.id,
          reason: reason || undefined,
          remove_data: deleteData,
          expires: date,
        })
      }

      open = false

      toast({
        content: banned ?'Unbanned that user.' :'Banned that user.',
        type: 'success',
      })

      item.banned = !banned
    } catch (err) {
      toast({
        content: errorMessage(err),
        type: 'error',
      })
    }

    loading = false
  }
</script>

<Modal
  bind:open
  title={banned ?'Unbanning User' :'Banning User'}
>
  {#if item}
    <form
      class="flex flex-col gap-4"
      onsubmit={(e) => {
        e.preventDefault()
        submit()
      }}
    >
      <div class="flex items-center gap-1">
        <Avatar url={item.avatar} alt={item.name} width={24} />
        <span class="font-bold">{item.name}</span>
      </div>
      {#if community}
        <CommunityLink {community} avatar />
      {/if}
      <MarkdownEditor
        required
        bind:value={reason}
        label=Reason
      />
      {#if !banned}
        <Switch bind:checked={deleteData}>
          Delete data
          {#snippet description()}
            This will delete ALL of this user's submissions here.
          {/snippet}
        </Switch>
        <Label text=Expires (UTC)>
          <Duration bind:value={expires}></Duration>
        </Label>
      {/if}
      <Button submit color="primary" {loading} disabled={loading} size="lg">
        Submit
      </Button>
    </form>
  {/if}
</Modal>

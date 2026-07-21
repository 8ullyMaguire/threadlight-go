<script lang="ts">
  import type { Person } from '$lib/api/types'

  import MarkdownEditor from '$lib/app/markdown/MarkdownEditor.svelte'
  import { Button, Modal } from 'mono-svelte'
  import UserLink from '../user/UserLink.svelte'

  interface Props {
    open?: boolean
    user: Person
    denying: boolean
    denyReason?: string
  }

  let {
    open = $bindable(false),
    user,
    denying = $bindable(),
    denyReason = $bindable(''),
  }: Props = $props()

  let loading = false

  function deny() {
    denying = true
    open = false
  }
</script>

<Modal bind:open title="Deny application">
  <form
    onsubmit={(e) => {
      e.preventDefault()
      deny()
    }}
    class="flex flex-col gap-4"
  >
    <UserLink avatar {user} />
    <MarkdownEditor
      bind:value={denyReason}
      label=Reason
      rows={4}
    />
    <Button color="primary" size="lg" submit {loading} disabled={loading}>
      Deny
    </Button>
  </form>
</Modal>

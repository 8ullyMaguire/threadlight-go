<script lang="ts">
  import { client } from '$lib/api/client.svelte'
  import { errorMessage } from '$lib/app/error'

  import { Button, TextInput, toast } from 'mono-svelte'

  interface Props {
    flair?: string
    community: number
    onsubmit?: (flair: string) => void
  }

  let { flair: passedFlair = '', community, onsubmit }: Props = $props()
  let flair = $state(passedFlair)
  let loading = $state(false)

  async function submit() {
    try {
      loading = true
      if (!client().setFlair) throw new Error('unsupported')

      await client().setFlair!({ community_id: community, flair_text: flair })

      onsubmit?.(flair)
    } catch (err) {
      toast({
        content: errorMessage(err as string),
        type: 'error',
      })
    }
    loading = false
  }
</script>

<form
  onsubmit={(e) => {
    e.preventDefault()
    submit()
  }}
  class="contents"
>
  <TextInput bind:value={flair} label=Content required />
  <Button {loading} submit color="primary" size="lg">
    Submit
  </Button>
</form>

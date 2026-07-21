<script lang="ts">
  import { goto } from '$app/navigation'
  import { client } from '$lib/api/client.svelte'
  import { profile } from '$lib/app/auth'

  import { Button, removeToast, Switch, TextInput, toast } from 'mono-svelte'

  let deletion = $state({
    modal: false,
    password: '',
    deleteContent: false,
  })

  async function deleteAccount(level: number) {
    switch (level) {
      case 0: {
        toast({
          content:'Are you sure you want to delete your account?',
          action: () => deleteAccount(1),
        })
        return
      }
      case 1: {
        toast({
          content:'Are you really sure?',
          action: () => deleteAccount(2),
        })
        return
      }
      case 2: {
        toast({
          content:'Final warning. Are you really sure?',
          action: () => deleteAccount(3),
        })
        return
      }
      case 3: {
        deletion.modal = true
        deletion.password = ''
        return
      }
    }

    if (!(deletion.password || null)) {
      toast({
        content:'You must provide your password.',
        type: 'warning',
      })
    }

    const id = toast({
      content:'Deleting your account...',
      loading: true,
    })

    try {
      await client().deleteAccount({
        password: deletion.password,

        delete_content: deletion.deleteContent,
      })

      profile.remove(profile.current.id)

      profile.meta.profile = -1
      toast({
        content:'Your account was deleted.',
      })
      goto('/')
    } catch (err) {
      toast({
        content: err as string,
        type: 'error',
      })
    } finally {
      removeToast(id)
    }
  }
</script>

<form
  onsubmit={(e) => {
    e.preventDefault()
    deleteAccount(4)
  }}
  class="w-full flex flex-col gap-4 max-w-xl"
>
  <TextInput
    label=Password
    type="password"
    bind:value={deletion.password}
  />
  <Switch bind:checked={deletion.deleteContent}>
    Delete content
    {#snippet description()}
      <span>
        This will delete ALL of your submissions.
      </span>
    {/snippet}
  </Switch>
  <Button submit color="danger" size="lg">
    Submit
  </Button>
</form>

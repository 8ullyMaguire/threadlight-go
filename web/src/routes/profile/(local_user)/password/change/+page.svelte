<script lang="ts">
  import { client } from '$lib/api/client.svelte'
  import { profile } from '$lib/app/auth'
  import { errorMessage } from '$lib/app/error'

  import { Button, TextInput, toast } from 'mono-svelte'

  let oldPassword = $state(''),
    newPassword = $state(''),
    newPasswordVerify = $state('')
  let loading = $state(false)

  async function changePassword() {
    try {
      loading = true
      const res = await client().changePassword({
        new_password: newPassword,
        new_password_verify: newPasswordVerify,
        old_password: oldPassword,
      })
      if (res?.jwt) {
        const { instance, client } = profile.current
        profile.remove(profile.current.id)
        await profile.add(res.jwt, instance, client)

        toast({ content:'Your login was refreshed.', type: 'success' })
      } else {
        throw new Error('Invalid credentials')
      }

      toast({ content:'Your password was reset.' })
    } catch (e) {
      toast({
        content: errorMessage(e as string),
        type: 'error',
      })
    }

    loading = false
  }
</script>

<form
  onsubmit={(e) => {
    e.preventDefault()
    changePassword()
  }}
  class="flex flex-col gap-4 w-full max-w-xl"
>
  <TextInput
    bind:value={oldPassword}
    label="Current password"
    type="password"
    minlength={10}
    required
  />
  <TextInput
    bind:value={newPassword}
    label="New password"
    type="password"
    minlength={10}
    required
  />
  <TextInput
    bind:value={newPasswordVerify}
    label="Verify new password"
    type="password"
    minlength={10}
    required
  />
  <Button size="lg" color="primary" submit {loading} disabled={loading}>
    Submit
  </Button>
</form>

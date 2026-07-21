<script lang="ts">
  import { getClient, validateInstance } from '$lib/api/client.svelte'
  import { profile } from '$lib/app/auth'
  import { errorMessage } from '$lib/app/error'

  import {
    DEFAULT_INSTANCE_URL,
    LINKED_INSTANCE_URL,
  } from '$lib/app/instance.svelte'
  import { Header } from '$lib/ui/layout'
  import { Button, TextInput, toast } from 'mono-svelte'
  import { preventDefault } from 'svelte/legacy'

  let instance = $state(LINKED_INSTANCE_URL || profile.current.instance || '')
  let email = $state('')
  let loading = $state(false)

  async function submit() {
    loading = true
    try {
      if (!(await validateInstance(instance)))
        throw new Error('Failed to contact that server URL with those settings. Is it down?')

      await getClient(instance).passwordReset({
        email: email,
      })

      toast({
        content:'A password reset link was sent to your email.',
        type: 'success',
      })
    } catch (err) {
      toast({
        content: errorMessage(err as string),
        type: 'error',
      })
    }
    loading = false
  }
</script>

<div class="my-auto max-w-xl mx-auto flex flex-col gap-2">
  <Header>Reset Password</Header>
  <p>
    Enter your account's email, and a password reset link will be sent. If your account does not have an email, contact your server admins.
  </p>
  <form class="mt-2 flex flex-col gap-4" onsubmit={preventDefault(submit)}>
    {#if !LINKED_INSTANCE_URL}
      <TextInput
        bind:value={instance}
        label=Server domain
        placeholder={DEFAULT_INSTANCE_URL}
        required
      />
    {/if}
    <TextInput
      bind:value={email}
      label=Email
      type="email"
      required
      placeholder="example@example.com"
    />
    <Button color="primary" size="lg" {loading} disabled={loading} submit>
      Submit
    </Button>
  </form>
</div>

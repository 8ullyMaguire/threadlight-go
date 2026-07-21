<script lang="ts">
  import { goto } from '$app/navigation'
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

  interface Props {
    data: {
      token: string
    }
  }

  let { data }: Props = $props()

  let instance = $state(LINKED_INSTANCE_URL || profile.current.instance || '')
  let password = $state(''),
    password_verify = $state('')
  let loading = $state(false)

  async function submit() {
    loading = true
    try {
      if (!(await validateInstance(instance)))
        throw new Error('Failed to contact that server URL with those settings. Is it down?')

      await getClient(instance).passwordChangeAfterReset({
        password: password,
        password_verify: password_verify,
        token: data.token,
      })

      toast({
        content:'Your password was reset.',
        type: 'success',
      })

      goto('/login')
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
  <Header>Change password</Header>
  <p>
    Choose your new password.
  </p>
  <form class="mt-2 flex flex-col gap-4" onsubmit={preventDefault(submit)}>
    {#if !LINKED_INSTANCE_URL}
      <TextInput
        bind:value={instance}
        label=Password
        placeholder={DEFAULT_INSTANCE_URL}
        required
      >
        <span class="font-normal text-xs">
          Which server is your account hosted on?
        </span>
      </TextInput>
    {/if}
    <TextInput
      bind:value={password}
      label="New password"
      type="password"
      required
    />
    <TextInput
      bind:value={password_verify}
      label="Verify new password"
      type="password"
      required
    />
    <Button color="primary" size="lg" {loading} disabled={loading} submit>
      Submit
    </Button>
  </form>
</div>

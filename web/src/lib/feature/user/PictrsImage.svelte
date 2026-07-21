<script lang="ts">
  import { client } from '$lib/api/client.svelte'
  import type { LocalImage, Person } from '$lib/api/types'
  import { profile } from '$lib/app/auth'

  import { instanceToURL } from '$lib/app/util.svelte'
  import { showImage } from '$lib/ui/generic/ExpandableImage.svelte'
  import { publishedToDate } from '$lib/ui/util/date'
  import { action, Button, modal, toast } from 'mono-svelte'
  import RelativeDate from 'mono-svelte/util/RelativeDate.svelte'
  import { ArrowDownTray, Trash } from 'svelte-hero-icons/dist'
  import UserLink from './UserLink.svelte'

  let loading = $state(false)

  async function deleteImage(image: LocalImage) {
    if (!profile.current?.jwt) return

    try {
      loading = true
      const res = await client().deleteImage({
        token: image.pictrs_delete_token,
        filename: image.pictrs_alias,
      })

      ondelete?.(res)
    } catch (e) {
      toast({ content: e as string, type: 'error' })
    }
    loading = false
  }

  interface Props {
    image: LocalImage
    user?: Person
    ondelete?: (res: boolean) => void
  }

  let { image, user, ondelete }: Props = $props()
</script>

<div class="flex flex-col gap-1">
  {#snippet img()}
    <button
      onclick={() =>
        showImage(
          `${instanceToURL(profile.current.instance)}/pictrs/image/${image.pictrs_alias}`,
        )}
      class="cursor-pointer"
    >
      <img
        src="{instanceToURL(
          profile.current.instance,
        )}/pictrs/image/{image.pictrs_alias}"
        width="500"
        height="500"
        class="aspect-square w-full h-full object-cover rounded-xl"
        alt="pictrs"
      />
    </button>
  {/snippet}

  {@render img()}

  {#if user}
    <UserLink {user} />
  {/if}
  <div class="flex items-center gap-2">
    <RelativeDate
      class="text-sm text-slate-700 dark:text-zinc-300"
      date={publishedToDate(image.published)}
    />
    <Button
      title=Download
      href="{instanceToURL(
        profile.current.instance,
      )}/pictrs/image/{image.pictrs_alias}"
      size="square-md"
      class="ml-auto"
      icon={ArrowDownTray}
    />
    <Button
      title=Delete
      onclick={() => {
        modal({
          title:'Confirm',
          body: '',
          snippet: img,
          actions: [
            action({
              close: true,
              content:'Cancel',
            }),
            action({
              type: 'danger',
              action: () => deleteImage(image),
              close: true,
              content:'Delete',
            }),
          ],
        })
      }}
      size="square-md"
      {loading}
      disabled={loading}
      icon={Trash}
    />
  </div>
</div>

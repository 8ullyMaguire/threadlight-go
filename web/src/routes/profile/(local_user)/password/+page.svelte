<script lang="ts">

  import Placeholder from '$lib/ui/info/Placeholder.svelte'
  import { Key } from 'svelte-hero-icons/dist'
</script>

<div class="h-full grid place-items-center">
  <Placeholder
    icon={Key}
    title=Manage how you log in
    description=You can enable 2FA for greater security, change your password, view logged in devices, or delete your account here.
  />
</div>

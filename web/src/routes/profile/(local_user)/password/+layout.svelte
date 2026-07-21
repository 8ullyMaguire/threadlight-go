<script>
  import { page } from '$app/state'

  import { Header, Tabs } from '$lib/ui/layout'

  let { children } = $props()

  const routes = [
    {
      name:'2FA',
      href: '/profile/password/2fa',
    },
    {
      name:'Change password',
      href: '/profile/password/change',
    },
    {
      name:'Logins',
      href: '/profile/password/logins',
    },
    {
      name:'Delete account',
      href: '/profile/password/delete',
    },
  ]
</script>

<Header pageHeader>
  {routes.find((r) => page.url.pathname == r.href)?.name ??'Credentials'}
  {#snippet extended()}
    <Tabs style="subpage" margin={false} {routes}></Tabs>
  {/snippet}
</Header>

<div
  class={[
    page.url.pathname == '/profile/password' && ' p-8',
    'flex flex-col justify-center items-center h-full',
  ]}
>
  {@render children?.()}
</div>

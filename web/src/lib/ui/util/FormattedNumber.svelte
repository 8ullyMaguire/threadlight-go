<script lang="ts">

  interface Props {
    number: number
    locale?: string
    options?: Intl.NumberFormatOptions
    class?: string
  }

  let {
    number,
    locale: localeToUse = 'en',
    options = {
      notation: 'compact',
    },
    class: clazz = '',
  }: Props = $props()
</script>

<span class={clazz ?? ''}>
  {Intl.NumberFormat(localeToUse, options).format(number)}
</span>

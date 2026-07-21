import { env } from '$env/dynamic/public'

export const ssr = env.PUBLIC_SSR_ENABLED?.toLowerCase() == 'true'

export async function load() {
  return
}

import { client } from '$lib/api/client.svelte'
import type { AffinityEntry } from '$lib/api/threadlight/types'

export interface PageData {
  affinities: AffinityEntry[]
}

export async function load(): Promise<PageData> {
  const affinities = await client().getAffinities()
  return { affinities }
}
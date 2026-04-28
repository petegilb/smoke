export type GameListItem = {
  app_id: number
  name: string
  header_image: string
  type: string
  genres: string[]
  developers: string[]
  publishers: string[]
  release_date: string
  short_description: string
  current_followers: number
  wishlist_rank: number | null
  delta_24h: number | null
  delta_7d: number | null
  sparkline: number[] // backend serializes int64; safe within JS number range for follower counts
}

export type Meta = {
  last_scraped_at: string
  next_scrape_at: string
}

export type SortMode = 'followers' | 'trending' | 'pct'

export type ListParams = {
  sort: SortMode
  type?: string
  indie?: boolean
  limit?: number
}

export async function fetchGameList(params: ListParams): Promise<GameListItem[]> {
  const q = new URLSearchParams({ sort: params.sort })
  if (params.type) q.set('type', params.type)
  if (params.indie) q.set('indie', 'true')
  if (params.limit) q.set('limit', String(params.limit))

  const res = await fetch(`/api/games/list?${q.toString()}`)
  if (!res.ok) throw new Error(`fetchGameList: ${res.status}`)
  return res.json()
}

export async function fetchMeta(): Promise<Meta> {
  const res = await fetch('/api/meta')
  if (!res.ok) throw new Error(`fetchMeta: ${res.status}`)
  return res.json()
}

import { useEffect, useState } from 'react'
import { fetchGameList, type GameListItem } from '../lib/api'

const numberFmt = new Intl.NumberFormat('en-US')

// Deterministic 32-bit hash of a string. Same input → same output, so every
// visitor sees the same "Game of the Day" and it only flips at UTC midnight.
function hashString(s: string): number {
  let h = 2166136261
  for (let i = 0; i < s.length; i++) {
    h ^= s.charCodeAt(i)
    h = Math.imul(h, 16777619)
  }
  return h >>> 0
}

function GameOfDay() {
  const [game, setGame] = useState<GameListItem | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetchGameList({ sort: 'trending', limit: 50 })
      .then(items => {
        if (items.length === 0) {
          setGame(null)
          return
        }
        const today = new Date().toISOString().slice(0, 10) // UTC YYYY-MM-DD
        const idx = hashString(today) % items.length
        setGame(items[idx])
      })
      .catch(() => setGame(null))
      .finally(() => setLoading(false))
  }, [])

  if (loading) {
    return (
      <div className="hero bg-base-200 rounded-box my-4">
        <div className="hero-content py-8">
          <span className="loading loading-spinner loading-md" />
        </div>
      </div>
    )
  }

  if (!game) return null

  const storeURL = `https://store.steampowered.com/app/${game.app_id}/`

  return (
    <a
      href={storeURL}
      target="_blank"
      rel="noopener noreferrer"
      className="hero rounded-box my-4 overflow-hidden block hover:opacity-95 transition-opacity"
      style={{
        backgroundImage: `linear-gradient(rgba(0,0,0,0.5), rgba(0,0,0,0.7)), url(${game.header_image})`,
        backgroundSize: 'cover',
        backgroundPosition: 'center',
      }}
    >
      <div className="hero-content text-neutral-content py-12">
        <div className="max-w-md text-center">
          <div className="badge badge-warning mb-3">Game of the Day</div>
          <h1 className="text-3xl font-bold mb-2">{game.name}</h1>
          {game.wishlist_rank != null && (
            <div className="text-sm opacity-90 mb-1">Wishlist rank #{game.wishlist_rank}</div>
          )}
          {game.delta_7d !== null ? (
            <p className="text-sm opacity-90">
              +{numberFmt.format(game.delta_7d)} followers in the last 7 days
            </p>
          ) : (
            <p className="text-sm opacity-90">Trending right now</p>
          )}
        </div>
      </div>
    </a>
  )
}

export default GameOfDay

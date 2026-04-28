import { useEffect, useState } from 'react'
import { fetchGameList, type GameListItem } from '../lib/api'

const numberFmt = new Intl.NumberFormat('en-US')

function GameOfDay() {
  const [game, setGame] = useState<GameListItem | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetchGameList({ sort: 'trending', limit: 1 })
      .then(items => setGame(items[0] ?? null))
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

  return (
    <div
      className="hero rounded-box my-4 overflow-hidden"
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
          {game.delta_7d !== null ? (
            <p className="text-sm opacity-90">
              +{numberFmt.format(game.delta_7d)} followers in the last 7 days
            </p>
          ) : (
            <p className="text-sm opacity-90">Trending right now</p>
          )}
        </div>
      </div>
    </div>
  )
}

export default GameOfDay

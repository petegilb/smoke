import { useEffect, useState } from 'react'
import { useSearchParams } from 'react-router'
import { fetchGameList, type GameListItem, type SortMode } from '../lib/api'
import GameOfDay from '../components/GameOfDay'
import RefreshCountdown from '../components/RefreshCountdown'
import Filters from '../components/Filters'
import GameList from '../components/GameList'

function Home() {
  const [params, setParams] = useSearchParams()
  const sortParam = params.get('sort')
  const sort: SortMode =
    sortParam === 'trending' ? 'trending'
    : sortParam === 'pct' ? 'pct'
    : sortParam === 'gain_24h' ? 'gain_24h'
    : 'followers'
  const type = params.get('type') ?? ''
  const indie = params.get('indie') === 'true'

  const [state, setState] = useState<{
    status: 'loading' | 'ok' | 'error'
    games: GameListItem[]
    error?: string
  }>({ status: 'loading', games: [] })

  useEffect(() => {
    let cancelled = false
    fetchGameList({ sort, type: type || undefined, indie })
      .then(games => {
        if (!cancelled) setState({ status: 'ok', games })
      })
      .catch(e => {
        if (!cancelled) setState({ status: 'error', games: [], error: String(e) })
      })
    return () => {
      cancelled = true
    }
  }, [sort, type, indie])

  const updateParam = (key: string, value: string | null) => {
    const next = new URLSearchParams(params)
    if (value === null || value === '' || value === 'false') next.delete(key)
    else next.set(key, value)
    setParams(next, { replace: true })
  }

  return (
    <div className="container mx-auto px-4 max-w-6xl">
      <GameOfDay />

      <div className="flex flex-wrap items-center justify-between gap-4 mb-4">
        <div role="tablist" className="tabs tabs-boxed">
          <button
            role="tab"
            className={`tab ${sort === 'followers' ? 'tab-active' : ''}`}
            onClick={() => updateParam('sort', null)}
          >
            Top Wishlisted
          </button>
          <button
            role="tab"
            className={`tab ${sort === 'gain_24h' ? 'tab-active' : ''}`}
            onClick={() => updateParam('sort', 'gain_24h')}
            title="Biggest follower gains in the last 24 hours"
          >
            Last 24h
          </button>
          <button
            role="tab"
            className={`tab ${sort === 'trending' ? 'tab-active' : ''}`}
            onClick={() => updateParam('sort', 'trending')}
            title="Biggest follower gains in the last 7 days"
          >
            Trending
          </button>
          <button
            role="tab"
            className={`tab ${sort === 'pct' ? 'tab-active' : ''}`}
            onClick={() => updateParam('sort', 'pct')}
            title="Fastest percentage growth in the last 7 days"
          >
            % Movers
          </button>
        </div>
        <RefreshCountdown />
      </div>

      <div className="mb-4">
        <Filters
          type={type}
          indie={indie}
          onTypeChange={t => updateParam('type', t)}
          onIndieChange={v => updateParam('indie', v ? 'true' : null)}
        />
      </div>

      {state.status === 'loading' && (
        <div className="flex justify-center py-12">
          <span className="loading loading-spinner loading-lg" />
        </div>
      )}
      {state.status === 'error' && (
        <div role="alert" className="alert alert-error">
          <span>{state.error}</span>
        </div>
      )}
      {state.status === 'ok' && state.games.length === 0 && (
        <div className="text-center py-12 text-base-content/60">
          No games match these filters.
        </div>
      )}
      {state.status === 'ok' && state.games.length > 0 && <GameList games={state.games} />}
    </div>
  )
}

export default Home

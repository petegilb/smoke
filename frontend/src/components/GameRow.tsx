import type { GameListItem } from '../lib/api'
import Sparkline from './Sparkline'

type Props = { game: GameListItem }

const numberFmt = new Intl.NumberFormat('en-US')

function formatDelta(n: number | null): string {
  if (n === null) return '—'
  if (n === 0) return '0'
  return (n > 0 ? '+' : '') + numberFmt.format(n as number)
}

function deltaClass(n: number | null) {
  if (n === null || n === 0) return 'text-base-content/50'
  return n > 0 ? 'text-success' : 'text-error'
}

function GameRow({ game }: Props) {
  const baseline =
    game.delta_7d === null ? null : game.current_followers - game.delta_7d
  const pct7d =
    game.delta_7d !== null && baseline !== null && baseline > 0
      ? (game.delta_7d / baseline) * 100
      : null
  const pctLabel =
    pct7d === null
      ? '—'
      : (pct7d > 0 ? '+' : '') + pct7d.toFixed(Math.abs(pct7d) >= 100 ? 0 : 1) + '%'

  const storeURL = `https://store.steampowered.com/app/${game.app_id}/`
  const release = game.release_date?.trim() || 'TBA'
  const dev = game.developers?.[0] || '—'
  const pub = game.publishers?.[0] || '—'

  return (
    <div className="flex items-start gap-4 px-4 py-3 border-b border-base-300 hover:bg-base-200/50">
      <div
        className="w-10 flex-shrink-0 text-right font-mono text-sm text-base-content/60 pt-1"
        title="Today's wishlist rank"
      >
        {game.wishlist_rank != null ? `#${game.wishlist_rank}` : '—'}
      </div>
      <a
        href={storeURL}
        target="_blank"
        rel="noopener noreferrer"
        className="flex-shrink-0"
        title="Open on Steam"
      >
        <img
          src={game.header_image}
          alt=""
          loading="lazy"
          className="w-32 h-16 object-cover rounded-md bg-base-300 hover:opacity-80 transition-opacity"
        />
      </a>
      <div className="flex-1 min-w-0">
        <div className="font-medium truncate">{game.name}</div>
        <div className="flex flex-wrap gap-1 my-1">
          <span className="badge badge-xs badge-ghost">{game.type}</span>
          {game.genres.slice(0, 3).map(g => (
            <span key={g} className="badge badge-xs badge-outline">{g}</span>
          ))}
        </div>
        <div className="text-xs text-base-content/70 truncate">
          <span className="font-medium">{dev}</span>
          {dev !== pub && <> · pub. <span className="font-medium">{pub}</span></>}
          {' · '}<span>{release}</span>
        </div>
        {game.short_description && (
          <p
            className="text-xs text-base-content/60 line-clamp-1 mt-1"
            title={game.short_description}
          >
            {game.short_description}
          </p>
        )}
      </div>
      <div className="text-right w-28 flex-shrink-0">
        <div className="font-mono text-sm">{numberFmt.format(game.current_followers)}</div>
        <div className="text-xs text-base-content/60">followers</div>
      </div>
      <div className="text-right w-24 flex-shrink-0 font-mono text-sm">
        <div className={deltaClass(game.delta_24h)}>{formatDelta(game.delta_24h)}</div>
        <div className="text-xs text-base-content/60">24h</div>
        <div className={`mt-0.5 ${deltaClass(pct7d)}`}>{pctLabel}</div>
        <div className="text-xs text-base-content/60">7d</div>
      </div>
      <Sparkline points={game.sparkline} className="flex-shrink-0 mt-1" />
    </div>
  )
}

export default GameRow

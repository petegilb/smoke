import { useRef } from 'react'
import { useVirtualizer } from '@tanstack/react-virtual'
import type { GameListItem } from '../lib/api'
import GameRow from './GameRow'

type Props = { games: GameListItem[] }

const ROW_HEIGHT = 116

function GameList({ games }: Props) {
  const parentRef = useRef<HTMLDivElement>(null)

  const virtualizer = useVirtualizer({
    count: games.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => ROW_HEIGHT,
    overscan: 8,
  })

  return (
    <div
      ref={parentRef}
      className="rounded-box border border-base-300 bg-base-100 overflow-auto"
      style={{ height: '70vh' }}
    >
      <div style={{ height: virtualizer.getTotalSize(), position: 'relative' }}>
        {virtualizer.getVirtualItems().map(v => {
          const game = games[v.index]
          return (
            <div
              key={game.app_id}
              style={{
                position: 'absolute',
                top: 0,
                left: 0,
                right: 0,
                height: ROW_HEIGHT,
                transform: `translateY(${v.start}px)`,
              }}
            >
              <GameRow game={game} />
            </div>
          )
        })}
      </div>
    </div>
  )
}

export default GameList

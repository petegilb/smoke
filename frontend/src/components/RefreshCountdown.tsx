import { useEffect, useState } from 'react'
import { fetchMeta } from '../lib/api'

function format(ms: number): { h: number; m: number; s: number } {
  const total = Math.max(0, Math.floor(ms / 1000))
  return {
    h: Math.floor(total / 3600),
    m: Math.floor((total % 3600) / 60),
    s: total % 60,
  }
}

function RefreshCountdown() {
  const [nextAt, setNextAt] = useState<Date | null>(null)
  const [now, setNow] = useState(() => Date.now())

  useEffect(() => {
    fetchMeta()
      .then(m => setNextAt(new Date(m.next_scrape_at)))
      .catch(() => setNextAt(null))
  }, [])

  useEffect(() => {
    const id = setInterval(() => setNow(Date.now()), 1000)
    return () => clearInterval(id)
  }, [])

  if (!nextAt) return null

  const { h, m, s } = format(nextAt.getTime() - now)

  return (
    <div className="text-sm text-base-content/70 flex items-center gap-2">
      <span>Next refresh in</span>
      <span className="font-mono tabular-nums">
        {String(h).padStart(2, '0')}:{String(m).padStart(2, '0')}:{String(s).padStart(2, '0')}
      </span>
    </div>
  )
}

export default RefreshCountdown

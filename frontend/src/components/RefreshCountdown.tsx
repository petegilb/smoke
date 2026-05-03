import { useEffect, useState } from 'react'
import { fetchMeta, type Meta } from '../lib/api'

function format(ms: number): { h: number; m: number; s: number } {
  const total = Math.max(0, Math.floor(ms / 1000))
  return {
    h: Math.floor(total / 3600),
    m: Math.floor((total % 3600) / 60),
    s: total % 60,
  }
}

function RefreshCountdown() {
  const [meta, setMeta] = useState<Meta | null>(null)
  const [now, setNow] = useState(() => Date.now())

  useEffect(() => {
    let cancelled = false
    fetchMeta()
      .then(m => { if (!cancelled) setMeta(m) })
      .catch(() => { if (!cancelled) setMeta(null) })
    return () => { cancelled = true }
  }, [])

  useEffect(() => {
    const id = setInterval(() => setNow(Date.now()), 1000)
    return () => clearInterval(id)
  }, [])

  if (!meta) return null

  if (meta.scrape_in_progress) {
    return (
      <div className="text-sm text-base-content/70 flex items-center gap-2">
        <span className="loading loading-ring loading-xs" />
        <span>DB Refresh in Progress...</span>
      </div>
    )
  }

  const { h, m, s } = format(new Date(meta.next_scrape_at).getTime() - now)

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

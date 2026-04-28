type Props = {
  points: number[]
  width?: number
  height?: number
  className?: string
}

const MIN_POINTS_FOR_LINE = 7

function Sparkline({ points, width = 96, height = 28, className = '' }: Props) {
  if (points.length === 0) {
    return (
      <div
        style={{ width, height }}
        className={`flex items-center justify-center text-xs text-base-content/40 ${className}`}
        title="No history yet"
      >
        —
      </div>
    )
  }

  const min = Math.min(...points)
  const max = Math.max(...points)
  const range = max - min || 1
  const stepX = points.length > 1 ? width / (points.length - 1) : 0
  const trendingUp = points[points.length - 1] >= points[0]
  const colorClass = trendingUp ? 'fill-success stroke-success' : 'fill-error stroke-error'

  const positions = points.map((v, i) => ({
    x: points.length === 1 ? width / 2 : i * stepX,
    y: height - ((v - min) / range) * height,
  }))

  if (points.length < MIN_POINTS_FOR_LINE) {
    return (
      <svg width={width} height={height} className={className} aria-hidden="true">
        {positions.map((p, i) => (
          <circle key={i} cx={p.x} cy={p.y} r="1.75" className={colorClass} />
        ))}
      </svg>
    )
  }

  const path = positions
    .map((p, i) => `${i === 0 ? 'M' : 'L'} ${p.x.toFixed(1)} ${p.y.toFixed(1)}`)
    .join(' ')

  return (
    <svg width={width} height={height} className={className} aria-hidden="true">
      <path d={path} fill="none" className={colorClass} strokeWidth="1.5" />
    </svg>
  )
}

export default Sparkline

type Props = {
  type: string
  indie: boolean
  onTypeChange: (t: string) => void
  onIndieChange: (v: boolean) => void
}

const TYPE_OPTIONS = [
  { value: '', label: 'All types' },
  { value: 'game', label: 'Games' },
  { value: 'dlc', label: 'DLC' },
  { value: 'demo', label: 'Demos' },
  { value: 'music', label: 'Soundtracks' },
  { value: 'software', label: 'Software' },
]

function Filters({ type, indie, onTypeChange, onIndieChange }: Props) {
  return (
    <div className="flex flex-wrap items-center gap-4">
      <select
        className="select select-sm select-bordered"
        value={type}
        onChange={e => onTypeChange(e.target.value)}
      >
        {TYPE_OPTIONS.map(o => (
          <option key={o.value} value={o.value}>{o.label}</option>
        ))}
      </select>
      <label className="label cursor-pointer gap-2">
        <span className="label-text">Indie only</span>
        <input
          type="checkbox"
          className="toggle toggle-sm toggle-primary"
          checked={indie}
          onChange={e => onIndieChange(e.target.checked)}
        />
      </label>
    </div>
  )
}

export default Filters

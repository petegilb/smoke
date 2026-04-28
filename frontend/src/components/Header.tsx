import { Link } from 'react-router'
import incenseIcon from '../assets/incense_icon.png'

function Header() {
  return (
    <div className="navbar bg-base-200">
      <div className="flex-1">
        <Link to="/" className="btn btn-ghost text-xl gap-2">
          <img src={incenseIcon} alt="" className="w-7 h-7" />
          smoke
        </Link>
      </div>
      <div className="flex-none text-right text-xs text-base-content/70 max-w-xs hidden sm:block pr-2">
        Tracking Steam follower counts as a wishlist proxy.
      </div>
    </div>
  )
}

export default Header

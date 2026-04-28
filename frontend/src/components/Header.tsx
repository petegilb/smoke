import { Link } from 'react-router'

function Header() {
  return (
    <div className="navbar bg-base-200">
      <div className="flex-1">
        <Link to="/" className="btn btn-ghost text-xl">smoke</Link>
      </div>
      <div className="flex-none">
        <ul className="menu menu-horizontal px-1">
          <li><Link to="/rankings">Rankings</Link></li>
          <li><Link to="/games">Games</Link></li>
        </ul>
      </div>
    </div>
  )
}

export default Header
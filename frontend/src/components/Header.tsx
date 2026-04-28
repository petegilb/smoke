import { Link } from 'react-router'

function Header() {
  return (
    <header>
      <Link to="/">smoke</Link>
      <nav>
        <Link to="/rankings">Rankings</Link>
        <Link to="/games">Games</Link>
      </nav>
    </header>
  )
}

export default Header
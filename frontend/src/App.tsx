import { BrowserRouter, Routes, Route } from 'react-router'
import Home from './pages/Home'
import Rankings from './pages/Rankings'
import Header from './components/Header'

function App() {

  return (
    <BrowserRouter>
      <Header></Header>

      {/* route definitions */}
      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/rankings" element={<Rankings />} />
      </Routes>
    </BrowserRouter>
  )
}

export default App

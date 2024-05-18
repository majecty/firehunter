import './app.css'
import Router from 'preact-router'
import { Home } from './pages/Home'
import { Movie } from './pages/Movie'

export function App() {
  return (
    <>
    <Router>
      <Home path="/"/>
      <Movie path="/movie/:id"/>
    </Router>
    </>
  )
}

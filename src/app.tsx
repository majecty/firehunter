import './app.css'
import Router from 'preact-router'
import { Home } from './pages/Home'
import { Movie } from './pages/Movie'
import { AframeSample } from './pages/AframeSample'

export function App() {
  return (
    <>
    <Router>
      <Home path="/"/>
      <Movie path="/movie/:id"/>
      <AframeSample path="/aframe"/>
    </Router>
    </>
  )
}

import './app.css'
import Router from 'preact-router'
import { Home } from './pages/firsttry/Home'
import { Movie } from './pages/firsttry/Movie'
import { AframeSample } from './pages/firsttry/AframeSample'
import { Home2 } from './pages/secondtry/SecondHome'
import { Movie2 } from './pages/secondtry/Movie2'

export function App() {
  return (
    <>
    <Router>
      <Home path="/"/>
      <Movie path="/movie/:id"/>
      <AframeSample path="/aframe"/>

      <Home2 path="/second"/>
      <Movie2 path="/second/movie/:id"/>
    </Router>
    </>
  )
}

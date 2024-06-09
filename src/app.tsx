import './app.css'
import Router from 'preact-router'
import { Home } from './pages/firsttry/Home'
import { Movie } from './pages/firsttry/Movie'
import { AframeSample } from './pages/firsttry/AframeSample'
import { Home2 } from './pages/secondtry/SecondHome'
import { Movie2 } from './pages/secondtry/Movie2'
import { VideoLoadTest } from './pages/secondtry/ViodeoLoadTest'
import { RootHome } from './pages/RootHome'

export function App() {
  return (
    <>
    <Router>
      <RootHome path="/"/>

      <Home path="/first"/>
      <Movie path="/first/movie/:id"/>
      <AframeSample path="/first/aframe"/>

      <Home2 path="/second"/>
      <Movie2 path="/second/movie/:id"/>
      <VideoLoadTest path="/second/videoloadtest"/>
    </Router>
    </>
  )
}

import './app.css'
import Router from 'preact-router'

function Home({ ...props }) {
  return <div>
    <h1>Home</h1>
    <pre>{JSON.stringify(props, null, 2)}</pre>
  </div>
}

function Movie({ ...props }) {
  return <div>
    <h1>Movie</h1>
    <pre>{JSON.stringify(props, null, 2)}</pre>
  </div>
}

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

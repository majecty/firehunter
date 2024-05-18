
export function Movie({ ...props }) {
  return <div>
    <h1>Movie</h1>
    <pre>{JSON.stringify(props, null, 2)}</pre>
  </div>
}

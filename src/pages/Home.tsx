
export function Home({ ...props }) {
  return <div>
    <h1>Home</h1>
    <pre>{JSON.stringify(props, null, 2)}</pre>
  </div>
}

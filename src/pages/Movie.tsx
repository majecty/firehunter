import { route } from "preact-router";

export function Movie({ ...props }) {

  return <div>
    <h1>Movie {props.id}</h1>
    <button onClick={() => route('/')}>Home으로 돌아가기</button>
    <pre>{JSON.stringify(props, null, 2)}</pre>
  </div>
}

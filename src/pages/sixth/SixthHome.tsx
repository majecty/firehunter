import { route } from "preact-router"

export function SixthHome({ ...props }) {
  console.log("SixthHome", props)

  return <div>
    <h1>여섯번째 시도</h1>
    <p>
      공유기 안에서 영상을 가져와서 재생합니다.
    </p>
    <button onClick={() => route("/sixth/movie/1")}>첫번째 영상</button>
    <button onClick={() => route("/sixth/movie/2")}>두번째 영상</button>
    <button onClick={() => route("/sixth/movie/3")}>세번째 영상</button>
    <button onClick={() => route("/sixth/movie/4")}>네번째 영상</button>
    <button onClick={() => route("/sixth/movie/5")}>다섯번째 영상</button>
  </div>
}

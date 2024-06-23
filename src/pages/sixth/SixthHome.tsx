import { route } from "preact-router"

export function SixthHome({ ...props }) {
  console.log("SixthHome", props)
  const videoBaseUrl = props.videoBaseUrl;
  // if (!videoBaseUrl) {
  //   return <div>
  //     <h1>여섯번째 시도</h1>
  //     <p>노트북에서 보이는 QR코드로 접속해주세요.</p>
  //     <p>{videoBaseUrl}</p>
  //   </div>
  // }

  // let decodedUrl;
  // try {
  //  decodedUrl = atob(videoBaseUrl.replace(/_/g, '/').replace(/-/g, '+'))
  // } catch (err) {
  //   if (err instanceof Error) {
  //     decodedUrl = "올바르지 않은 주소입니다. " + err.message;
  //   } else {
  //     decodedUrl = "올바르지 않은 주소입니다. " + err;
  //   }
  // }

  return <div>
    <h1>여섯번째 시도</h1>
    <p>
      공유기 안에서 영상을 가져와서 재생합니다.
    </p>
    {/* <p>
      공유기 주소: {decodedUrl}
    </p> */}
    <button onClick={() => route("/sixth/movie/1?videoBaseUrl=" + videoBaseUrl)}>첫번째 영상</button>
    <button onClick={() => route("/sixth/movie/2?videoBaseUrl=" + videoBaseUrl)}>두번째 영상</button>
    <button onClick={() => route("/sixth/movie/3?videoBaseUrl=" + videoBaseUrl)}>세번째 영상</button>
    <button onClick={() => route("/sixth/movie/4?videoBaseUrl=" + videoBaseUrl)}>네번째 영상</button>
    <button onClick={() => route("/sixth/movie/5?videoBaseUrl=" + videoBaseUrl)}>다섯번째 영상</button>
  </div>
}

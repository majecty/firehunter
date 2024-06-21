
export function FifthHome({ ...props }) {
  console.log("FifthHome", props);

  return <div>
    <h1>다섯번째 시도</h1>
    <p>
      원본 영상과 같은 품질의 영상을 2초로 잘라서 360도로 무한재생합니다.
      iphone SE2에서 TV로 airplay했을 때 영상의 품질을 확인하기 위함입니다.
    </p>
    <button>750MB 원본을 2초로 자른 영상 테스트</button>
    <button>700MB 원본을 2초로 자른 영상 테스트</button>
    <button>400MB 원본을 2초로 자른 영상 테스트</button>
    <button>300MB 원본을 2초로 자른 영상 테스트</button>
    <button>200MB 원본을 2초로 자른 영상 테스트</button>
  </div>
}

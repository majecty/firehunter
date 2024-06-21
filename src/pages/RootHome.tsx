import { route } from "preact-router";

export function RootHome({ ...props }) {
  console.log('RootHome', props);
  return <div>
    <p>
      360 영상을 테스트하기 위한 페이지
    </p>
    <button onClick={() => route('/first')}>
      첫번째 시도: 영상을 실시간 스트리밍하면서 싱크를 맞춤
    </button>
    <br />
    <button onClick={() => route('/second')}>
      두번째 시도: 영상을 미리 로딩하고 싱크를 맞춤
    </button>
    <br />
    <button onClick={() => route('/third')}>
      세번째 시도: webrtc 스트리밍을 사용
    </button>
    <button onClick={() => route('/fourth')}>
      네번째 시도: hls 스트리밍을 사용
    </button>

    <button onClick={() => route('/fifth')}>
      다섯번째 시도: 크키별 360 영상을 짧게 잘라서 TV에서 재생
    </button>
    </div>;
}

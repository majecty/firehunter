import { route } from "preact-router";

export function ThirdHome({ ...props }) {
  console.log("ThirdHome", props);

  return <div>
    <h2>Third Home</h2>
    <button onClick={() => {
      route('/third/webrtcmanualsessionexchange')
    }}>수동 webrtc 테스트</button>
  </div>
}

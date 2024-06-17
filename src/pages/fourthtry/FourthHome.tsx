import { route } from "preact-router";

export function FourthHome({ ...props }) {
  console.log("FourthHome", props);

  return <div>
    <h2>Fourth Home</h2>
    <button onClick={() => {
      route('/fourth/hlstest')
    }}>hls 테스트</button>
  </div>
}

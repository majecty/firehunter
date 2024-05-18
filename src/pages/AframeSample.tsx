import { route } from "preact-router";
import "aframe";
// @ts-ignore
import { Entity, Scene } from "aframe-react";

const sampleVideoUrl = "https://firehunter.s3.ap-northeast-2.amazonaws.com/0518sample.mp4";

declare module "preact" {
  namespace JSX {
    interface IntrinsicElements {
      "a-videosphere": any;
      "a-assets": any;
      "a-text": any;
      "a-camera": any;
      "a-entity": any;
    }
  }
}

export function AframeSample({ ...props }) {
  console.log(props);


  return <div>
    <h1>AframeSample</h1>
    <button onClick={() => route('/')}>Home으로 돌아가기</button>
    <Scene>
      <a-assets>
        <video id="sample-video" autoplay loop={true}
          src={sampleVideoUrl} crossorigin="anonymous"
        />
      </a-assets>
      <a-videosphere src="#sample-video"></a-videosphere>
    </Scene>

    <p style={{ position: "absolute", top: "70%", left: "50%", transform: "translate(-50%, -50%)", zIndex: "9999", color: "white", backgroundColor: "black" }}>hi</p>
  </div>
}

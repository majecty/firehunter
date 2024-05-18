import { route } from "preact-router";
import { useEffect, useRef, useState } from "preact/hooks";
// @ts-ignore
import { Entity, Scene } from "aframe-react";

const sampleVideoUrl = "https://firehunter.s3.ap-northeast-2.amazonaws.com/0518sample.mp4";

export function Movie({ ...props }) {
  const [target, setTarget] = useState(0);
  const [current, setCurrent] = useState(0);
  const videoRef = useRef<HTMLVideoElement>(null);

  useEffect(() => {
    const interval = setInterval(() => {
      setTarget(new Date().getSeconds());
      if (videoRef.current === null) {
        console.log("videoRef is null");
        return;
      }
      setCurrent(Math.floor(videoRef.current.currentTime));
    }, 100);

    return () => {
      clearInterval(interval);
    };
  }, []);

  useEffect(() => {
    if (videoRef.current === null) {
      console.log("videoRef is null");
      return;
    }
    const currentMillis = videoRef.current.currentTime * 1000;
    const now = new Date();
    const currentTimeMillis =  now.getSeconds() * 1000 + now.getMilliseconds();

    let diff = Math.abs(currentMillis - currentTimeMillis);
    if (diff > 30 * 1000) {
      diff = 60 * 1000 - diff;
    }
    if (diff > 500) {
      videoRef.current.currentTime = now.getSeconds() + now.getMilliseconds() / 1000;
      console.log("current ", currentMillis, "now ", now.getSeconds() * 1000 + now.getMilliseconds(), "diff", diff);
    }
  }, [target]);


  return <div>
    <h1>Movie {props.id}</h1>
    <button onClick={() => route('/')}>Home으로 돌아가기</button>
    <p>
      영상 싱크: {target}
    </p>
    <pre>{JSON.stringify(props, null, 2)}</pre>

    <Scene>
      <a-assets>
        <video id="sample-video" autoplay loop={true}
          src={sampleVideoUrl} crossorigin="anonymous"
          ref={videoRef}
        />
      </a-assets>
      <a-videosphere src="#sample-video"></a-videosphere>
    </Scene>

    <p style={{ position: "absolute", top: "70%", left: "50%", transform: "translate(-50%, -50%)", zIndex: "9999", color: "white", backgroundColor: "black" }}>{props.id}번째 타블렛 싱크 목표: {target} 현재: {current}</p>
  </div>
}

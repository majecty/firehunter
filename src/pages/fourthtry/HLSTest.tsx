import { route } from "preact-router";
import { useEffect, useRef, useState } from "preact/hooks";

// @ts-ignore
import { Entity, Scene } from "aframe-react";

const sampleVideoUrl = "https://firehunter.s3.ap-northeast-2.amazonaws.com/0518hls/0518sample.m3u8";

export function HLSTest({ ...props }) {
  // console.log("HLSTest", props);

  const [target, setTarget] = useState(0);
  const [current, setCurrent] = useState(0);
  const videoRef = useRef<HTMLVideoElement>(null);
  const [enableSync, setEnableSync] = useState(true);
  const [fov, setFov] = useState(80);

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
    if (!enableSync) {
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
    <h2>HLSTest</h2>
    <button onClick={() => route('/fourth')}>Home으로 돌아가기</button>
    <p>
      영상 싱크: {target}
      FOV: {fov}
    </p>
    <pre>{JSON.stringify(props, null, 2)}</pre>

    <video id="sample-video" autoplay loop={true}
      // https://gist.github.com/lukebussey/4d27678c72580aeb660c19a6fb73e9ee
      src={sampleVideoUrl} crossorigin="anonymous"
      ref={videoRef}
      type="application/x-mpegURL"
    style={{ width: "50%", height: "50%", position: "absolute", top: "50%", left: "50%", transform: "translate(-50%, -50%)", zIndex: "9999" }}
    />

    <p style={{ position: "absolute", top: "70%", left: "50%", transform: "translate(-50%, -50%)", zIndex: "9999", color: "white", backgroundColor: "black" }}>{props.id}번째 타블렛 싱크 목표: {target} 현재: {current}</p>
    <button style={{ position: "absolute", top: "90%", left: "50%", transform: "translate(-50%, -50%)", zIndex: "9999"}} onClick={() => setEnableSync(!enableSync)}>싱크 {enableSync ? "끄기" : "켜기"}</button>
    <p style={{ position: "absolute", top: "95%", left: "50%", transform: "translate(-50%, -50%)", zIndex: "9999", color: "white", backgroundColor: "black" }}>FOV</p>
    <input type="range" min="30" max="120" value={fov} onChange={(e) => setFov(parseInt((e.target! as any).value))} style={{ position: "absolute", top: "95%", left: "50%", transform: "translate(-50%, -50%)", zIndex: "9999"}} />
    {videoRef.current !== null && videoRef.current.paused && <button style={{ position: "absolute", top: "95%", left: "50%", transform: "translate(-50%, -50%)", zIndex: "9999"}} onClick={() => videoRef.current!.play()}>재생</button>}
    {/* {videoRef.current !== null && !videoRef.current.paused && <button style={{ position: "absolute", top: "95%", left: "50%", transform: "translate(-50%, -50%)", zIndex: "9999"}} onClick={() => videoRef.current!.pause()}>멈춤</button>} */}
  </div>
}

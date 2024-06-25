import { route } from "preact-router";
import { useEffect, useRef, useState } from "preact/hooks";
// @ts-ignore
import { Entity, Scene } from "aframe-react";
import Hls from "hls.js";

const sampleVideoUrl = "https://192-168-17-2.i.juhyung.dev:8443/videos/0518hls/0518sample.m3u8";

export function SeventhMovieHLS({ ...props }) {
  // console.log("Movie", props);
  const [target, setTarget] = useState(0);
  const [current, setCurrent] = useState(0);
  const videoRef = useRef<HTMLVideoElement>(null);
  const [enableSync, setEnableSync] = useState(true);
  const [fov, setFov] = useState(80);

  // if (Hls.isSupported()) {
  if (!Hls.isSupported()) {
    return <div>
      <h1>이 브라우저는 HLS를 지원하지 않습니다.</h1>
    </div>
  }

  useEffect(() => {
    if (videoRef.current === null) {
      console.log("videoRef is null");
      return;
    }
    if (videoRef.current.canPlayType('application/vnd.apple.mpegurl')) {
      videoRef.current.src = sampleVideoUrl;
    } else {
      const hls = new Hls();
      hls.loadSource(sampleVideoUrl);
      hls.attachMedia(videoRef.current);
    }
  }, [videoRef]);

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
    if (diff > 3000) {
      videoRef.current.currentTime = now.getSeconds() + now.getMilliseconds() / 1000;
      console.log("current ", currentMillis, "now ", now.getSeconds() * 1000 + now.getMilliseconds(), "diff", diff);
    }
  }, [target]);


  return <div>
    <h1>Movie {props.id}</h1>
    <button onClick={() => route('/')}>Home으로 돌아가기</button>
    <p>
      영상 싱크: {target}
      FOV: {fov}
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
      <a-camera fov={fov.toString()}>  </a-camera>
    </Scene>

    <p style={{ position: "absolute", top: "70%", left: "50%", transform: "translate(-50%, -50%)", zIndex: "9999", color: "white", backgroundColor: "black" }}>{props.id}번째 타블렛 싱크 목표: {target} 현재: {current}</p>
    <button style={{ position: "absolute", top: "90%", left: "50%", transform: "translate(-50%, -50%)", zIndex: "9999"}} onClick={() => setEnableSync(!enableSync)}>싱크 {enableSync ? "끄기" : "켜기"}</button>
    <p style={{ position: "absolute", top: "95%", left: "50%", transform: "translate(-50%, -50%)", zIndex: "9999", color: "white", backgroundColor: "black" }}>FOV</p>
    <input type="range" min="30" max="120" value={fov} onChange={(e) => setFov(parseInt((e.target! as any).value))} style={{ position: "absolute", top: "95%", left: "50%", transform: "translate(-50%, -50%)", zIndex: "9999"}} />
    {videoRef.current !== null && videoRef.current.paused && <button style={{ position: "absolute", top: "95%", left: "50%", transform: "translate(-50%, -50%)", zIndex: "9999"}} onClick={() => videoRef.current!.play()}>재생</button>}
  </div>
}

import { route } from "preact-router";
import { useEffect, useRef, useState } from "preact/hooks";
// @ts-ignore
import { Entity, Scene } from "aframe-react";

const sampleVideoUrl = "https://firehunter.s3.ap-northeast-2.amazonaws.com/0518sample.mp4";
const smallSampleVideoUrl = "https://firehunter.s3.ap-northeast-2.amazonaws.com/0609sample.mp4";

export function Movie2({ ...props }) {
  const [target, setTarget] = useState(0);
  const [current, setCurrent] = useState(0);
  const videoRef = useRef<HTMLVideoElement>(null);
  const [fov, setFov] = useState(80);
  const [progress, setProgress] = useState(0);

  function handleXMLHTTPRequestEvent(event: Event) {
    console.log("handleXMLHTTPRequestEvent", event);
    if (event.type !== "progress") {
      console.log(event.type, (event as any).loaded, (event as any).total);
    }
    if ((event as any).total !== 0) {
      setProgress((event as any).loaded / (event as any).total);
    }

    if (event.type === "loadend") {
      setProgress(1);
      console.log("loadend");
      const newBlobURI = URL.createObjectURL((event.target as XMLHttpRequest).response);
      if (videoRef.current === null) {
        console.log("videoRef is null", videoRef);
        return;
      }
      videoRef.current.setAttribute("src", newBlobURI);
    }
  }

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
    if (progress !== 1) {
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
  }, [target, videoRef, progress]);


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
          crossorigin="anonymous"
          ref={videoRef}
        />
      </a-assets>
      <a-videosphere src="#sample-video"></a-videosphere>
      <a-camera fov={fov.toString()}>  </a-camera>
    </Scene>

    <p style={{ position: "absolute", top: "70%", left: "50%", transform: "translate(-50%, -50%)", zIndex: "9999", color: "white", backgroundColor: "black" }}>{props.id}번째 타블렛 싱크 목표: {target} 현재: {current}</p>
    <p style={{ position: "absolute", top: "95%", left: "50%", transform: "translate(-50%, -50%)", zIndex: "9999", color: "white", backgroundColor: "black" }}>FOV</p>
    <input type="range" min="30" max="120" value={fov} onChange={(e) => setFov(parseInt((e.target! as any).value))} style={{ position: "absolute", top: "95%", left: "50%", transform: "translate(-50%, -50%)", zIndex: "9999"}} />
    {progress !== 1 && <progress value={progress} max="1" style={{ position: "absolute", top: "60%", left: "50%", transform: "translate(-50%, -50%)", zIndex: "9999"}}></progress>}
    {progress === 0 && <button style={{ position: "absolute", top: "50%", left: "50%", transform: "translate(-50%, -50%)", zIndex: "9999"}} onClick={() => loadVideo(handleXMLHTTPRequestEvent, sampleVideoUrl)}>영상 로드</button>}
    {progress === 0 && <button style={{ position: "absolute", top: "40%", left: "50%", transform: "translate(-50%, -50%)", zIndex: "9999"}} onClick={() => loadVideo(handleXMLHTTPRequestEvent, smallSampleVideoUrl)}>작은 영상 로드</button>}
    {progress === 1 && videoRef.current !== null && videoRef.current.paused && <button style={{ position: "absolute", top: "95%", left: "50%", transform: "translate(-50%, -50%)", zIndex: "9999"}} onClick={() => videoRef.current!.play()}>재생</button>}
  </div>
}

function loadVideo(handleXMLHTTPRequestEvent: (event: Event) => void, videoUrl: string){
  const xhr = new XMLHttpRequest();
  xhr.addEventListener("loadstart", handleXMLHTTPRequestEvent);
  xhr.addEventListener("load", handleXMLHTTPRequestEvent);
  xhr.addEventListener("loadend", handleXMLHTTPRequestEvent);
  xhr.addEventListener("progress", handleXMLHTTPRequestEvent);
  xhr.addEventListener("error", handleXMLHTTPRequestEvent);
  xhr.addEventListener("abort", handleXMLHTTPRequestEvent);
  xhr.open("GET", videoUrl);
  xhr.responseType = "blob";
  xhr.send();
}

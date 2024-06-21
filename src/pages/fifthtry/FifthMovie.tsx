import { route } from "preact-router";
// @ts-ignore
import { Entity, Scene } from "aframe-react";
import { useRef, useState } from "preact/hooks";

type MovieSize = "750MB" | "700MB" | "400MB" | "300MB" | "200MB";
const movieSize: MovieSize[] = ["750MB", "700MB", "400MB", "300MB", "200MB"];

export function FifthMovie({ ...props }) {
  console.log("FifthMovie", props)

  if (!movieSize.includes(props.size)) {
    return (
      <div>
        <h1>영상</h1>
        <p>잘못된 사이즈입니다.</p>
        <p>{props.size}</p>
        <p>{JSON.stringify(props, null, 2)}</p>
      </div>
    )
  }

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

  return <div>
    <h1> 영상 </h1>
    <p> {props.size} </p>
    <button onClick={() => route('/fifth')}>Home으로 돌아가기</button>
    <p>
      FOV: {fov}
    </p>

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

    <p style={{ position: "absolute", top: "95%", left: "50%", transform: "translate(-50%, -50%)", zIndex: "9999", color: "white", backgroundColor: "black" }}>FOV</p>
    <input type="range" min="30" max="120" value={fov} onChange={(e) => setFov(parseInt((e.target! as any).value))} style={{ position: "absolute", top: "95%", left: "50%", transform: "translate(-50%, -50%)", zIndex: "9999"}} />
    {progress !== 1 && <progress value={progress} max="1" style={{ position: "absolute", top: "60%", left: "50%", transform: "translate(-50%, -50%)", zIndex: "9999"}}></progress>}
    {progress === 0 && <button style={{ position: "absolute", top: "50%", left: "50%", transform: "translate(-50%, -50%)", zIndex: "9999"}} onClick={() => loadVideo(handleXMLHTTPRequestEvent, getMovieUrl(props.size))}>영상 로드</button>}
    {progress === 1 && videoRef.current !== null && videoRef.current.paused && <button style={{ position: "absolute", top: "85%", left: "50%", transform: "translate(-50%, -50%)", zIndex: "9999"}} onClick={() => videoRef.current!.play()}>재생</button>}
  </div>
}

function getMovieUrl(size: MovieSize) {
  switch (size) {
    case "750MB":
      return "https://firehunter.s3.ap-northeast-2.amazonaws.com/twoseconds/mv750_2s.mp4";
    case "700MB":
      return "https://firehunter.s3.ap-northeast-2.amazonaws.com/twoseconds/mv700_2s.mp4";
    case "400MB":
      return "https://firehunter.s3.ap-northeast-2.amazonaws.com/twoseconds/mv400_2s.mp4";
    case "300MB":
      return "https://firehunter.s3.ap-northeast-2.amazonaws.com/twoseconds/mv300_2s.mp4";
    case "200MB":
      return "https://firehunter.s3.ap-northeast-2.amazonaws.com/twoseconds/mv200_2s.mp4";
  }
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

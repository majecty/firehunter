import { useState } from "preact/hooks";

const sampleVideoUrl = "https://firehunter.s3.ap-northeast-2.amazonaws.com/0518sample.mp4";

export function VideoLoadTest({ ...props }) {
  console.log(props);
  const [progress, setProgress] = useState(0);

  function handleXMLHTTPRequestEvent(event: Event) {
    console.log("handleXMLHTTPRequestEvent", event);
    console.log(event.type, (event as any).loaded, (event as any).total);
    if ((event as any).total === 0) {
      return;
    }
    setProgress((event as any).loaded / (event as any).total);

    if (event.type === "loadend") {
      console.log("loadend");
      const newBlobURI = URL.createObjectURL((event.target as XMLHttpRequest).response);
      document.getElementById("video-test-video")!.setAttribute("src", newBlobURI);
    }
  }


  return <div>
    hi
    <button onClick={() => loadVideo(handleXMLHTTPRequestEvent)}> Load Video </button>
    <br />
    <progress value={progress} max="1"></progress>
    <br />
    <button onClick={() => {
      const video = document.getElementById("video-test-video") as HTMLVideoElement;
      if (video.paused) {
        video.play();
      } else {
        video.pause();
      }
    }}> Play/Pause </button>
    <br />
    <video id="video-test-video"></video>
  </div>
}

function loadVideo(handleXMLHTTPRequestEvent: (event: Event) => void){
  const xhr = new XMLHttpRequest();
  xhr.addEventListener("loadstart", handleXMLHTTPRequestEvent);
  xhr.addEventListener("load", handleXMLHTTPRequestEvent);
  xhr.addEventListener("loadend", handleXMLHTTPRequestEvent);
  xhr.addEventListener("progress", handleXMLHTTPRequestEvent);
  xhr.addEventListener("error", handleXMLHTTPRequestEvent);
  xhr.addEventListener("abort", handleXMLHTTPRequestEvent);
  xhr.open("GET", sampleVideoUrl);
  xhr.responseType = "blob";
  xhr.send();
}



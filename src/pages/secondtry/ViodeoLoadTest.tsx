import { useState } from "preact/hooks";

const sampleVideoUrl = "https://firehunter.s3.ap-northeast-2.amazonaws.com/0518sample.mp4";

export function VideoLoadTest({ ...props }) {
  const [progress, setProgress] = useState(0);

  function handleXMLHTTPRequestEvent(event: Event) {
    console.log("handleXMLHTTPRequestEvent", event);
    console.log(event.type, (event as any).loaded, (event as any).total);
    if ((event as any).total === 0) {
      return;
    }
    setProgress((event as any).loaded / (event as any).total);
  }


  return <div>
    hi
    <button onClick={() => loadVideo(handleXMLHTTPRequestEvent)}> Load Video </button>
    <br />
    <progress value={progress} max="1"></progress>
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
  xhr.send();
}



import { route } from "preact-router";
import { useEffect, useRef, useState } from "preact/hooks";
// @ts-ignore
import { Entity, Scene } from "aframe-react";

const pc = new RTCPeerConnection({
  iceServers: [
    {
      urls: "turn:turn.i.juhyung.dev:3478",
      username: "juhyung",
      credential: "juhyung",
    }
  //   {
  //   urls: 'stun:stun.l.google.com:19302'
  // }, {
  //   urls: "stun:stun2.1.google.com:19302",
  // }
]
});

export function SixthMovie({ ...props }) {
  console.log("SixthMovie", props);

  if (!["1", "2", "3", "4", "5"].includes(props.id)) {
    return <div>
      <h1>여섯번째 시도</h1>
      <p>잘못된 주소입니다. 주소에서 movie 뒤의 숫자는 1,2,3,4,5 중 하나여야 합니다.</p>
      <p>현재 movie 뒤의 숫자: {props.id}</p>
      </div>
  }

  const videoRef = useRef<HTMLVideoElement>(null);
  const [logs, setLogs] = useState<string[]>([]);
  const [localSessionDescription, setLocalSessionDescription] = useState<string>('');
  const [remoteSessionDescription, setRemoteSessionDescription] = useState<string>('');
  const [fov, setFov] = useState(80);

  useEffect(() => {
    console.log("ThirdHome useEffect");
    pc.ontrack = (event) => {
      console.log('ontrack', event);
      const el = document.createElement(event.track.kind) as any
      el.srcObject = event.streams[0];
      el.autoplay = true;
      el.controls = true;

      if (videoRef.current == null) {
        console.error('videoRootRef.current is null while ontrack');
        return;
      }
      if (event.streams.length !== 1) {
        console.error('event.streams.length !== 1 while ontrack ' + event.streams.length);
        console.error(event.streams);
        return;
      }
      videoRef.current.srcObject = event.streams[0];
    };

    pc.oniceconnectionstatechange = (event) => {
      console.log('oniceconnectionstatechange', ".");
      // do nothing
      // console.log('oniceconnectionstatechange', event);
      // setLogs((prev) => [...prev, `oniceconnectionstatechange: ${pc.iceConnectionState}`]);
    };

    pc.onicecandidate = (event) => {
      console.log('onicecandidate', ".");
      if (event.candidate == null) {
        console.log("localDescription", (JSON.stringify(pc.localDescription)?.substring(0, 50) ?? 'localDescription is null') + (JSON.stringify(pc.localDescription)?.length > 50 ? '...' : ''));
        setLocalSessionDescription(
          btoa(JSON.stringify(pc.localDescription))
        )

        fetch(`https://localhost:8123/client/offer`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            offer: JSON.stringify(pc.localDescription)
          })
        }).then((response) => {
          return response.json()
        }).then((data) => {
          if (typeof data.answer != 'string') {
            console.error('data.answer is not a string');
            console.error(data);
            return;
          }

          if (data.answer === '') {
            console.error('data.answer is empty');
            console.error(data);
            return;
          }

          console.log("setRemoteSessionDescription");
          setRemoteSessionDescription(data.answer);

          pc.setRemoteDescription(JSON.parse(atob(data.answer)));
        });
      } else {
        setLocalSessionDescription('.' + Math.random().toPrecision(2))
        console.log('onicecandidate', event.candidate);
      }
    };

    pc.addTransceiver('video', {
      direction: 'sendrecv'
    });

    pc.createOffer().then(d => pc.setLocalDescription((d))).catch(e => console.error(e));

  }, [])


  return <div>
    <h1>여섯번째 시도 {props.id}</h1>
    <p>
      공유기 안에서 영상을 가져와서 재생합니다.
    </p>
    <button onClick={() => route('/')}>Home으로 돌아가기</button>
    <p>
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

    <p style={{ position: "absolute", top: "95%", left: "50%", transform: "translate(-50%, -50%)", zIndex: "9999", color: "white", backgroundColor: "black" }}>FOV</p>
    <input type="range" min="30" max="120" value={fov} onChange={(e) => setFov(parseInt((e.target! as any).value))} style={{ position: "absolute", top: "95%", left: "50%", transform: "translate(-50%, -50%)", zIndex: "9999"}} />
  </div>
}

function getMovieFileUrl(id: "1" | "2" | "3" | "4" | "5", videoBaseUrl: string) {
  return `${videoBaseUrl}/videos/video-${id}.mp4`;
}

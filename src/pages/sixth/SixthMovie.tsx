import { route } from "preact-router";
import { useEffect, useRef, useState } from "preact/hooks";
// @ts-ignore
import { Entity, Scene } from "aframe-react";

const pc = new RTCPeerConnection({
  iceServers: [
    {
      urls: "stun:stun.i.juhyung.dev:3478",
      // urls: "stun:localhost:3478",
    }
  //   {
  //   urls: 'stun:stun.l.google.com:19302'
  // }, {
  //   urls: "stun:stun2.1.google.com:19302",
  // }
]
});

const signalServerUrl = 'wss://signal-firehunter.i.juhyung.dev/client/ws';
// const signalServerUrl = 'ws://localhost:8124/client/ws';


export function SixthMovie({ ...props }) {
  console.log("SixthMovie", props);
  const [wsOpen, setWsOpen] = useState(false);
  const [ws, setWs] = useState<WebSocket | null>(null);

  useEffect(() => {
    if (wsOpen === true) {
      return;
    }
    if (ws != null) {
      return
    }
    console.log("initialize websocket")
    const ws_ = new WebSocket(signalServerUrl);
    setWs(ws_);
    ws_.onopen = () => {
      console.log('ws.onopen');
      setWsOpen(true);
    }

    ws_.onclose = () => {
      console.log('ws.onclose');
      setWsOpen(false);
    }
  }, [wsOpen, ws]);

  if (wsOpen === false) {
    return <div>
      <h1>여섯번째 시도</h1>
      <p>서버와 연결 중입니다.</p>
    </div>
  }


  if (!["1", "2", "3", "4", "5"].includes(props.id)) {
    return <div>
      <h1>여섯번째 시도</h1>
      <p>잘못된 주소입니다. 주소에서 movie 뒤의 숫자는 1,2,3,4,5 중 하나여야 합니다.</p>
      <p>현재 movie 뒤의 숫자: {props.id}</p>
      </div>
  }

  const videoRef = useRef<HTMLVideoElement>(null);
  const [fov, setFov] = useState(80);

  useEffect(() => {
    if (ws == null) {
      return;
    }
    console.log("ThirdHome useEffect");

    ws.onmessage = e => {
      let msg = JSON.parse(e.data)
      if (msg == null) {
        console.error('websocket msg is null');
        return;
      }

      if (msg.type === 'offer') {
        console.error("offer type should not be received");
        return;
      }

      if (msg.type === 'answer') {
        console.log("setRemoteSessionDescription")
        pc.setRemoteDescription(msg.data);
        return;
      }

      if (msg.type === "candidate") {
        console.log("addIceCandidate")
        pc.addIceCandidate(msg.data);
        return;
      }
    };

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

    pc.onicecandidate = (event) => {
      if (event.candidate != null && event.candidate.candidate != null && event.candidate.candidate.length > 0) {
        console.log("send candidate")
        ws.send(JSON.stringify({
          type: 'candidate',
          data: event.candidate
        }));
      }
    }

    pc.oniceconnectionstatechange = (event) => {
      console.log('oniceconnectionstatechange', ".");
      console.log(event);
      // do nothing
      // console.log('oniceconnectionstatechange', event);
      // setLogs((prev) => [...prev, `oniceconnectionstatechange: ${pc.iceConnectionState}`]);
    };

    pc.addTransceiver('video', {
      direction: 'sendrecv'
    });

    pc.createOffer()
    .then(d => {
      pc.setLocalDescription((d))
      console.log("send offer")
      ws.send(JSON.stringify({
        type: 'offer',
        data: d
      }))
    })
    .catch(e => console.error(e));

  }, [ws])


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

        {/* <video id="sample-video" autoplay loop={true}
           crossorigin="anonymous"
          ref={videoRef}
        /> */}

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
    <button style={{position: "absolute", top: "90%", left: "50%", transform: "translate(-50%, -50%)", zIndex: "9999"}} onClick={() => {
      videoRef.current?.play();
    }}>Play</button>
  </div>
}

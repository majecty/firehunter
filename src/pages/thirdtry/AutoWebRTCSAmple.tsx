import { route } from "preact-router";
import { useEffect, useRef, useState } from "preact/hooks";

const pc = new RTCPeerConnection({
  iceServers: [
  //   {
  //   urls: 'stun:stun.l.google.com:19302'
  // }, {
  //   urls: "stun:stun2.1.google.com:19302",
  // }
]
});

export function WebRTCAutoSessionExchangeSample({ ...props }) {
  console.log("ThirdHome", props);
  const [logs, setLogs] = useState<string[]>([]);
  const [localSessionDescription, setLocalSessionDescription] = useState<string>('');
  const [remoteSessionDescription, setRemoteSessionDescription] = useState<string>('');
  const videoRef = useRef<HTMLVideoElement>(null);

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
      console.log('oniceconnectionstatechange', event);
      setLogs((prev) => [...prev, `oniceconnectionstatechange: ${pc.iceConnectionState}`]);
    };

    pc.onicecandidate = (event) => {
      console.log('onicecandidate', event);
      if (event.candidate == null) {
        console.log("localDescription", (JSON.stringify(pc.localDescription)?.substring(0, 50) ?? 'localDescription is null') + (JSON.stringify(pc.localDescription)?.length > 50 ? '...' : ''));
        setLocalSessionDescription(
          btoa(JSON.stringify(pc.localDescription))
        )

        fetch("http://localhost:8124/sessiondescription", {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            sessionDescription: btoa(JSON.stringify(pc.localDescription))
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
    <h2>Auto webrtc connect</h2>
    <button onClick={() => {
      route('/third')
    }}>Go back</button>
    <p>{localSessionDescription}</p>
    <button onClick={() => {
      navigator.clipboard.writeText(localSessionDescription)
        .then(() => {
          console.log('Session description copied to clipboard');
        })
        .catch((error) => {
          console.error('Failed to copy session description to clipboard:', error);
        });
    }}>Copy browser Session Description</button>
    <p>Remote Session Description</p>
    <p>{remoteSessionDescription}</p>
    <textarea value={remoteSessionDescription} onChange={(e) => {
      if (e.currentTarget == null) {
        console.error('e.target is null in textarea onChange')
        console.error(e);
        return;
      }
      if (typeof e.currentTarget.value !== 'string') {
        console.error('e.target.value is not a string in textarea onChange')
        console.error(e.target)
        return;
      }
      setRemoteSessionDescription(e.currentTarget.value);
    }}></textarea>
    <button onClick={() => {
      if (remoteSessionDescription === '') {
        console.error('remoteSessionDescription is empty');
        return;
      }

      try {
        pc.setRemoteDescription(JSON.parse(atob(remoteSessionDescription)));
      } catch (e) {
        console.error('Failed to set remote description:', e);
      }
    }}>Start Session</button>
    <video ref={videoRef} autoplay muted></video>
    <p>{logs.join('\n')}</p>
    <button onClick={() => {
      if (videoRef.current == null) {
        console.log('videoRootRef.current is null')
      } else {
        console.log('videoRootRef.current is not null')
        console.log(videoRef.current);
      }
    }}>check videoref</button>
  </div>
}

import { createRef } from "preact";
import { useEffect, useState } from "preact/hooks";

const pc = new RTCPeerConnection({
  iceServers: [{
    urls: 'stun:stun.l.google.com:19302'
  }, {
    urls: "stun:stun2.1.google.com:19302",
  }]
});

export function ThirdHome({ ...props }) {
  console.log("ThirdHome", props);
  const [logs, setLogs] = useState<string[]>([]);
  const [localSessionDescription, setLocalSessionDescription] = useState<string>('');
  const [remoteSessionDescription, setRemoteSessionDescription] = useState<string>('');
  const videoRootRef = createRef();

  useEffect(() => {
    console.log("ThirdHome useEffect");
    pc.ontrack = (event) => {
      console.log('ontrack', event);
      const el = document.createElement(event.track.kind) as any
      el.srcObject = event.streams[0];
      el.autoplay = true;
      el.controls = true;

      if (videoRootRef.current == null) {
        console.error('videoRootRef.current is null while ontrack');
        return;
      }
      videoRootRef.current.appendChild(el);

      // const remoteVideosElement = document.getElementById('remoteVideos');
      // if (remoteVideosElement == null) {
      //   console.error('remoteVideosElement is null');
      //   return;
      // }
      // remoteVideosElement.appendChild(el);
    };

    pc.oniceconnectionstatechange = (event) => {
      // console.log('oniceconnectionstatechange', event);
      // setLogs((prev) => [...prev, `oniceconnectionstatechange: ${pc.iceConnectionState}`]);
    };

    pc.onicecandidate = (event) => {
      console.log('onicecandidate', event);
      if (event.candidate == null) {
        console.log(JSON.stringify(pc.localDescription))
        setLocalSessionDescription(
          btoa(JSON.stringify(pc.localDescription))
        )
      }
    };

    pc.addTransceiver('video', {
      direction: 'sendrecv'
    });

    pc.createOffer().then(d => pc.setLocalDescription((d))).catch(e => console.error(e));

  }, [])

  return <div>
    <h2>Third Home</h2>
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
    <div ref={videoRootRef}></div>
    <p>{logs.join('\n')}</p>
  </div>
}

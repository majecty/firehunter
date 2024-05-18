import { route } from "preact-router";
import { useEffect, useState } from "preact/hooks";
import "aframe";
// @ts-ignore
import { Entity, Scene } from "aframe-react";

export function Movie({ ...props }) {
  const [count, setCount] = useState(0);

  useEffect(() => {
    const interval = setInterval(() => {
      setCount(new Date().getSeconds());
    }, 100);

    return () => {
      clearInterval(interval);
    };
  }, []);


  return <div>
    <h1>Movie {props.id}</h1>
    <button onClick={() => route('/')}>Home으로 돌아가기</button>
    <p>
      영상 싱크: {count}
    </p>
    <pre>{JSON.stringify(props, null, 2)}</pre>
    <Scene >
      <Entity geometry={{primitive: 'box'}} material={{color: 'red'}} position={{x: 0, y: 0, z: -5}}/>
      <Entity geometry={{primitive: 'sphere'}} material={{color: 'blue'}} position={{x: 2, y: 0, z: -5}}/>
      <Entity geometry={{primitive: 'cylinder'}} material={{color: 'green'}} position={{x: -2, y: 0, z: -5}}/>
      <Entity light={{type: 'point'}}/>
      <Entity gltf-model={{src: 'virtualcity.gltf'}}/>
      <Entity text={{value: 'Hello, WebVR!'}}/>
    </Scene>
  </div>
}

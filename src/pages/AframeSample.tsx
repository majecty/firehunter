import { route } from "preact-router";
import "aframe";
// @ts-ignore
import { Entity, Scene } from "aframe-react";

export function AframeSample({...props}) {
  return <div>
    <h1>AframeSample</h1>
    <button onClick={() => route('/')}>Home으로 돌아가기</button>
    <Scene>
      <Entity geometry={{primitive: 'box'}} material={{color: 'red'}} position={{x: 0, y: 0, z: -5}}/>
      <Entity geometry={{primitive: 'sphere'}} material={{color: 'blue'}} position={{x: 2, y: 0, z: -5}}/>
      <Entity geometry={{primitive: 'cylinder'}} material={{color: 'green'}} position={{x: -2, y: 0, z: -5}}/>
      <Entity light={{type: 'point'}}/>
      <Entity gltf-model={{src: 'virtualcity.gltf'}}/>
      <Entity text={{value: 'Hello, WebVR!'}}/>
    </Scene>
  </div>
}

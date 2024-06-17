import './app.css'
import Router from 'preact-router'
import { Home } from './pages/firsttry/Home'
import { Movie } from './pages/firsttry/Movie'
import { AframeSample } from './pages/firsttry/AframeSample'
import { Home2 } from './pages/secondtry/SecondHome'
import { Movie2 } from './pages/secondtry/Movie2'
import { VideoLoadTest } from './pages/secondtry/ViodeoLoadTest'
import { RootHome } from './pages/RootHome'
import { ThirdHome } from './pages/thirdtry/ThirdHome'
import { WebRTCManualSessionExchangeSample } from './pages/thirdtry/ManualWebRTCSample'
import { WebRTCAutoSessionExchangeSample } from './pages/thirdtry/AutoWebRTCSAmple'
import { FourthHome } from './pages/fourthtry/FourthHome'
import { HLSTest } from './pages/fourthtry/HLSTest'
import { HLSTest360 } from './pages/fourthtry/HLSTest360'
import { HLSTestWithLocal } from './pages/fourthtry/HLSTestWithLocal'

export function App() {
  return (
    <>
    <Router>
      <RootHome path="/"/>

      <Home path="/first"/>
      <Movie path="/first/movie/:id"/>
      <AframeSample path="/first/aframe"/>

      <Home2 path="/second"/>
      <Movie2 path="/second/movie/:id"/>
      <VideoLoadTest path="/second/videoloadtest"/>

      <ThirdHome path="/third"/>
      <WebRTCManualSessionExchangeSample path="/third/webrtcmanualsessionexchange"/>
      <WebRTCAutoSessionExchangeSample path="/third/webrtcautomaticsessionexchange"/>

      <FourthHome path="/fourth"/>
      <HLSTest path="/fourth/hlstest"/>
      <HLSTest360 path="/fourth/hlstest360"/>
      <HLSTestWithLocal path="/fourth/hlstestwithlocal"/>
    </Router>
    </>
  )
}

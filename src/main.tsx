import { render } from 'preact'
import { App } from './app.tsx'
import './index.css'

const version = '1.0.3'
console.log(`App version: ${version}`)

render(<App />, document.getElementById('app')!)

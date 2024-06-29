import { useState } from "preact/hooks";

export function BrowserTest({ ...props }) {
  console.log('BrowserTest', props)
  const [logs, setLogs] = useState<string[]>([]);

  return (
    <div>
      <h1>BrowserTest</h1>
      <p>
        브라우저 동작을 테스트하는 페이지
      </p>
      <button onClick={() => {
        const orientation = window.screen.orientation as any;
        if (orientation.lock == null) {
          setLogs([...logs, 'orientation.lock is not supported']);
          return;
        }
        orientation.lock("landscape-primary");
        setLogs([...logs, 'orientation.lock("landscape-primary")']);
      }}>Lock Screen</button>
      <button onClick={() => {
        window.screen.orientation.unlock();
        setLogs([...logs, 'orientation.unlock()']);
      }}>
       Unlock Screen</button>
       <p>
          {logs.map((log, index) => (
            <div key={index}>{log}</div>
          ))}
       </p>
    </div>
  )
}

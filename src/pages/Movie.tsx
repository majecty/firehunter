import { route } from "preact-router";
import { useEffect, useState } from "preact/hooks";

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
  </div>
}

import { route } from 'preact-router';
import { useEffect, useState } from 'preact/hooks';

export function Home({ ...props }) {
  const [count, setCount] = useState(0);

  const handleClick = (movieId: number) => {
    route(`/first/movie/${movieId}`);
  };

  useEffect(() => {
    const interval = setInterval(() => {
      setCount(new Date().getSeconds());
    }, 100);

    return () => {
      clearInterval(interval);
    };
  }, []);


  return <div>
    <h1>Home</h1>
    <p>
      단순하게 윕에서 영상을 실시간 스트리밍하면서 싱크를 맞춤
    </p>
    <br />
    <button onClick={() => handleClick(1)}> 첫번째 패드 </button>
    <button onClick={() => handleClick(2)}> 두번째 패드 </button>
    <button onClick={() => handleClick(3)}> 세번째 패드 </button>
    <br />
    <button onClick={() => handleClick(4)}> 네번째 패드 </button>
    <button onClick={() => handleClick(5)}> 다섯번째 패드 </button>
    <button onClick={() => handleClick(6)}> 여섯번째 패드 </button>
    <br />
    <p>
      영상 싱크: {count}
    </p>
    <button onClick={() => route('/aframe')}>aframe 샘플로 이동</button>
    <pre>{JSON.stringify(props, null, 2)}</pre>
  </div>
}

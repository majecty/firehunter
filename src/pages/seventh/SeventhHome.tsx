import { route } from 'preact-router';
import { useEffect, useState } from 'preact/hooks';

export function SeventhHome({ ...props }) {
  console.log('SeventhHome', props);
  const [count, setCount] = useState(0);

  const handleClick = (movieId: number) => {
    route(`/seventh/movie/${movieId}`);
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
    <h1>일곱번째 시도</h1>
    <p>
      공유기 안에서 https를 사용해서 영상 재생
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
  </div>
}

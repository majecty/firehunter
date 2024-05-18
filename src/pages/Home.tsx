import { route } from 'preact-router';

export function Home({ ...props }) {
  const handleClick = (movieId: number) => {
    route(`/movie/${movieId}`);
  };
  return <div>
    <h1>Home</h1>
    <button onClick={() => handleClick(1)}> 첫번째 패드 </button>
    <button onClick={() => handleClick(2)}> 두번째 패드 </button>
    <button onClick={() => handleClick(3)}> 세번째 패드 </button>
    <br />
    <button onClick={() => handleClick(4)}> 네번째 패드 </button>
    <button onClick={() => handleClick(5)}> 다섯번째 패드 </button>
    <button onClick={() => handleClick(6)}> 여섯번째 패드 </button>
    <pre>{JSON.stringify(props, null, 2)}</pre>
  </div>
}

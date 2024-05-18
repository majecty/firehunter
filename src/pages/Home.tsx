import { route } from 'preact-router';

export function Home({ ...props }) {
  const handleClick = () => {
    route('/movie/1');
  };
  return <div>
    <h1>Home</h1>
    <button onClick={handleClick}> 첫번째 패드 </button>
    <pre>{JSON.stringify(props, null, 2)}</pre>
  </div>
}

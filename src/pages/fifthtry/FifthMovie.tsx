type MovieSize = "750MB" | "700MB" | "400MB" | "300MB" | "200MB";
const movieSize: MovieSize[] = ["750MB", "700MB", "400MB", "300MB", "200MB"];

export function FifthMovie({ ...props }) {
  console.log("FifthMovie", props)

  if (!movieSize.includes(props.size)) {
    return (
      <div>
        <h1>영상</h1>
        <p>잘못된 사이즈입니다.</p>
        <p>{props.size}</p>
        <p>{JSON.stringify(props, null, 2)}</p>
      </div>
    )
  }

  return (<div>
    <h1> 영상 </h1>
    <p> {props.size} </p>
  </div>);
}

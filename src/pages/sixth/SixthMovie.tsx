export function SixthMovie({ ...props }) {
  console.log("SixthMovie", props);

  if (!["1", "2", "3", "4", "5"].includes(props.id)) {
    return <div>
      <h1>여섯번째 시도</h1>
      <p>잘못된 주소입니다. 주소에서 movie 뒤의 숫자는 1,2,3,4,5 중 하나여야 합니다.</p>
      <p>현재 movie 뒤의 숫자: {props.id}</p>
      </div>
  }

  return <div>
    <h1>여섯번째 시도 {props.id}</h1>
    <p>
      공유기 안에서 영상을 가져와서 재생합니다.
    </p>
  </div>
}

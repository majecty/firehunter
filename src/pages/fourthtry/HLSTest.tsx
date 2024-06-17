export function HLSTest({ ...props }) {
  console.log("HLSTest", props);

  return <div>
    <h2>HLSTest</h2>
    <video controls>
      <source src="http://localhost:8000/0518sample.m3u8" type="application/x-mpegURL"/>
    </video>
  </div>
}

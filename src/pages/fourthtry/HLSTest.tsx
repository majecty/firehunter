export function HLSTest({ ...props }) {
  console.log("HLSTest", props);

  return <div>
    <h2>HLSTest</h2>
    // https://gist.github.com/lukebussey/4d27678c72580aeb660c19a6fb73e9ee
    <video controls>
      <source src="https://firehunter.s3.ap-northeast-2.amazonaws.com/0518hls/0518sample.m3u8" type="application/x-mpegURL"/>
    </video>
  </div>
}

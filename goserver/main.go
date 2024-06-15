package main

// https://github.com/pion/webrtc/blob/master/examples/play-from-disk/main.go

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media"
	"github.com/pion/webrtc/v4/pkg/media/ivfreader"
)

const (
	largeVideoFileName = "./resource/output.ivf"
)

func main() {
	println("Hello, World!")

	_, err := os.Stat(largeVideoFileName)
	haveLargeVideoFile := !os.IsNotExist(err)
	if haveLargeVideoFile {
		println("Have large video file")
	} else {
		println("Don't have large video file " + largeVideoFileName)
	}

	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	})
	if err != nil {
		panic(err)
	}

	defer func() {
		if cErr := peerConnection.Close(); cErr != nil {
			fmt.Printf("Failed to close PeerConnection: %v\n", cErr)
		}
	}()

	iceConnectedCtx, iceConnectedCtxCancel := context.WithCancel(context.Background())

	file, openErr := os.Open(largeVideoFileName)
	if openErr != nil {
		panic(openErr)
	}

	_, header, openErr := ivfreader.NewWith(file)
	if openErr != nil {
		panic(openErr)
	}

	var trackCodec string
	// fourcc: https://en.wikipedia.org/wiki/FourCC
	switch header.FourCC {
	case "AV01":
		trackCodec = webrtc.MimeTypeAV1
	case "VP90":
		trackCodec = webrtc.MimeTypeVP9
	case "VP80":
		trackCodec = webrtc.MimeTypeVP8
	default:
		panic(fmt.Sprintf("Unable to handle FourCC %s", header.FourCC))
	}

	videoTrack, videoTrackErr := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{
		MimeType: trackCodec,
	}, "video", "pion")
	if videoTrackErr != nil {
		panic(videoTrackErr)
	}

	rtpSender, videoTrackErr := peerConnection.AddTrack(videoTrack)
	if videoTrackErr != nil {
		panic(videoTrackErr)
	}

	// Read incoming RTCP packets
	// Before these packets are returned they are processed by interceptors. For things
	// like NACK this needs to be called.
	go func() {
		rtcpBuf := make([]byte, 1500)
		for {
			if _, _, rtcpErr := rtpSender.Read(rtcpBuf); rtcpErr != nil {
				return
			}
		}
	}()

	go func() {
		// Open a IVF file and start reading using our IVFReader
		file, ivfErr := os.Open(largeVideoFileName)
		if ivfErr != nil {
			panic(ivfErr)
		}

		ivf, header, ivfErr := ivfreader.NewWith(file)
		if ivfErr != nil {
			panic(ivfErr)
		}

		<-iceConnectedCtx.Done()

		ticker := time.NewTicker(
			time.Millisecond *
				time.Duration(
					(float32(header.TimebaseNumerator) /
						float32(header.TimebaseDenominator) *
						1000)))
		for ; true; <-ticker.C {
			frame, _, ivfErr := ivf.ParseNextFrame()
			if errors.Is(ivfErr, io.EOF) {
				fmt.Printf("All video frames parsed and sent\n")
				os.Exit(0)
			}

			if ivfErr != nil {
				panic(ivfErr)
			}

			// Sample에 Timestamp 안 적어도 되나?
			// Sample에 duration을 임의로 저렇게 넣어도 되나?
			if ivfErr = videoTrack.WriteSample(
				media.Sample{Data: frame, Duration: time.Second}); ivfErr != nil {
				panic(ivfErr)
			}
		}
	}()

	peerConnection.OnICEConnectionStateChange(func(conectionState webrtc.ICEConnectionState) {
		fmt.Printf("Connection State has changed %s \n", conectionState.String())
		if conectionState == webrtc.ICEConnectionStateConnected {
			iceConnectedCtxCancel()
		}
	})

	peerConnection.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		fmt.Printf("Connection State has changed %s \n", s.String())

		if s == webrtc.PeerConnectionStateFailed {
			// 여기서 다른 처리를 더 할 수 있나?
			// TODO: connection이 끊겼다가 다시 열결되는 경우 어떻게 해야하는지 찾을 것
			fmt.Println("PeerConnection is failed")
			os.Exit(0)
		}

		if s == webrtc.PeerConnectionStateClosed {
			fmt.Println("PeerConnection is closed")
			os.Exit(0)
		}
	})

	offer := webrtc.SessionDescription{}
	decode(readUntilNewLine(), &offer)

	// TODO: remote description 의미 찾기
	if err = peerConnection.SetRemoteDescription(offer); err != nil {
		panic(err)
	}

	// TODO: answer의 의미 찾기
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		panic(err)
	}

	// TODO: gatheringComplete의 의미 찾기
	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

	if err = peerConnection.SetLocalDescription(answer); err != nil {
		panic(err)
	}

	<-gatherComplete

	// answer in base64
	fmt.Println(encode(peerConnection.LocalDescription()))

	// Block forever
	select {}
}

func decode(in string, obj *webrtc.SessionDescription) {
	b, err := base64.StdEncoding.DecodeString(in)
	if err != nil {
		panic(err)
	}

	if err = json.Unmarshal(b, obj); err != nil {
		panic(err)
	}
}

func readUntilNewLine() (in string) {
	var err error
	r := bufio.NewReader(os.Stdin)
	for {
		in, err = r.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			panic(err)
		}

		if in = strings.TrimSpace(in); len(in) > 0 {
			break
		}
	}

	fmt.Println("")
	return
}

func encode(obj *webrtc.SessionDescription) string {
	b, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}

	return base64.StdEncoding.EncodeToString(b)
}

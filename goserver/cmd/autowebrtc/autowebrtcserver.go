package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media"
	"github.com/pion/webrtc/v4/pkg/media/ivfreader"
)

const (
	largeVideoFileName = "./resource/output.ivf"
)

var (
	waitingReadForSessionDescription atomic.Bool
	sessionDescriptionReceived       = false
	sessionDecriptionChannel         = make(chan string)

	answerChannel = make(chan string)
)

func main() {
	go webrtcMain()
	// giuMain()

	// httpserver
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message": "hello world"}`))
	})
	http.HandleFunc("/sessiondescription", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.Header().Set("Content-Type", "application/json")

		if !waitingReadForSessionDescription.Load() && !sessionDescriptionReceived {
			fmt.Println("Not ready for session description")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"message": "Not ready for session description"}`))
			return
		}

		if !waitingReadForSessionDescription.Load() && sessionDescriptionReceived {
			fmt.Println("Already received session description")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"message": "Already received session description"}`))
			return
		}

		if r.Method != "POST" {
			fmt.Println("Method not allowed " + r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		// body에 sessiondescription이 있음
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			fmt.Printf("Failed to read body: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		jsonBody := make(map[string]string)
		if err := json.Unmarshal(body, &jsonBody); err != nil {
			fmt.Printf("Failed to unmarshal json: %v\n", err)
			fmt.Println(string(body))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		sessionDescription := jsonBody["sessionDescription"]
		if sessionDescription == "" {
			fmt.Println("SessionDescription is empty")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		sessionDecriptionChannel <- sessionDescription
		sessionDescriptionReceived = true

		answer := <-answerChannel

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "SessionDescription received", "answer": "` + answer + `"}`))
	})

	http.ListenAndServe(":8124", nil)
}

func webrtcMain() {
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

		fmt.Println("Wait for ICE Connection Connected")
		<-iceConnectedCtx.Done()
		fmt.Println("ICE Connection Connected, start sending video track")

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
		fmt.Printf("ICE Connection State has changed %s \n", conectionState.String())
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
	// fmt.Println(encode(peerConnection.LocalDescription()))
	fmt.Println("Answer is ready")
	answerChannel <- encode(peerConnection.LocalDescription())

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
	fmt.Println("Waiting for input")
	waitingReadForSessionDescription.Store(true)
	in = <-sessionDecriptionChannel
	fmt.Println("Received")
	return
}

func encode(obj *webrtc.SessionDescription) string {
	b, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}

	return base64.StdEncoding.EncodeToString(b)
}

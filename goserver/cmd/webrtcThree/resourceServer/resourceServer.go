package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media"
	"github.com/pion/webrtc/v4/pkg/media/h264reader"
)

var (
	h264FrameDuration = time.Millisecond * 33
)

func main() {
	fmt.Println("resourceServer start")
	ctx := context.Background()
	if err := webrtcMain(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func webrtcMain(ctx context.Context) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	fmt.Println("webrtcMain start")

	if err := checkVideoFile(); err != nil {
		return fmt.Errorf("failed to check video file: %w", err)
	}

	offer := ""
	registerWebRTCEvents(ctx, offer)

	<-ctx.Done()
	fmt.Println("webrtcMain end")
	return nil
}

var videoFileName = "videoFileName.mp4"

func checkVideoFile() error {
	if _, err := os.Stat(videoFileName); err != nil {
		return fmt.Errorf("failed to check video file: %w", err)
	}

	return nil
}

func registerWebRTCEvents(ctx context.Context, offer string) (answer webrtc.SessionDescription, err error) {
	fmt.Println("registerWebRTCEvents")

	peerConnection, err := createPeerConnection()
	if err != nil {
		return answer, fmt.Errorf("failed to create PeerConnection: %w", err)
	}
	defer func() {
		if err := peerConnection.Close(); err != nil {
			fmt.Printf("Failed to close PeerConnection: %v\n", err)
		}
	}()

	iceConnectedCtx, iceConnectedCtxCancel := context.WithCancel(ctx)

	videoTrack, videoTrackErr := createVideoTrack(peerConnection)
	if videoTrackErr != nil {
		return answer, fmt.Errorf("failed to create video track: %w", videoTrackErr)
	}

	go streamingVideo(iceConnectedCtx, videoTrack)

	registerConnectionStartedEvent(iceConnectedCtxCancel, peerConnection)
	registerConnectionFailedEvent(peerConnection)

	if err := acceptOffer(peerConnection, offer); err != nil {
		return fmt.Errorf("failed to accept offer: %w", err)
	}
	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

	answer, err = createAnswer(peerConnection)
	if err != nil {
		return answer, fmt.Errorf("failed to create answer: %w", err)
	}

	<-gatherComplete

	return answer, nil
}

func createPeerConnection() (*webrtc.PeerConnection, error) {
	return webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			webrtc.ICEServer{
				URLs:       []string{"turn:turn.i.juhyung.dev:3478"},
				Username:   "juhyung",
				Credential: "juhyung",
			},
		},
	})
}

func createVideoTrack(peerConnection *webrtc.PeerConnection) (*webrtc.TrackLocalStaticSample, error) {
	videoTrack, videoTrackErr := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264}, "video", "pion")
	if videoTrackErr != nil {
		return nil, fmt.Errorf("failed to create video track: %w", videoTrackErr)
	}

	rtpSender, videoTrackErr := peerConnection.AddTrack(videoTrack)
	if videoTrackErr != nil {
		return nil, fmt.Errorf("failed to add video track: %w", videoTrackErr)
	}

	// 이유는 모르겠지만 rtcp 를 받고 버려야함.
	go func() {
		rtcpBuf := make([]byte, 1500)
		for {
			if _, _, rtcpErr := rtpSender.Read(rtcpBuf); rtcpErr != nil {
				fmt.Printf("Failed to read RTCP packet: %v", rtcpErr)
				return
			}
		}
	}()

	return videoTrack, nil
}

func streamingVideo(iceConnectedCtx context.Context, videoTrack *webrtc.TrackLocalStaticSample) {
	file, h264Err := os.Open(videoFileName)
	if h264Err != nil {
		fmt.Printf("Failed to open video file: %v", h264Err)
		return
	}

	h264, h264Err := h264reader.NewReader(file)
	if h264Err != nil {
		fmt.Printf("Failed to create H264 reader: %v", h264Err)
		return
	}

	// connection이 되길 기다림
	<-iceConnectedCtx.Done()

	ticker := time.NewTicker(h264FrameDuration)
	for ; true; <-ticker.C {
		nal, h264Err := h264.NextNAL()
		if errors.Is(h264Err, io.EOF) {
			fmt.Println("End of video file")
			return
		}
		if h264Err != nil {
			fmt.Printf("Failed to read NAL: %v", h264Err)
			return
		}

		if h264Err = videoTrack.WriteSample(media.Sample{
			Data:     nal.Data,
			Duration: h264FrameDuration,
		}); h264Err != nil {
			fmt.Printf("Failed to write sample: %v", h264Err)
			return
		}
	}
}

func registerConnectionStartedEvent(iceConnectedCtxCancel context.CancelFunc, peerConnection *webrtc.PeerConnection) {
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("ICE Connection State has changed: %s\n", connectionState.String())
		if connectionState == webrtc.ICEConnectionStateConnected {
			iceConnectedCtxCancel()
		}
	})
}

func registerConnectionFailedEvent(peerConnection *webrtc.PeerConnection) {
	peerConnection.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		fmt.Printf("Peer Connection State has changed: %s\n", s.String())

		if s == webrtc.PeerConnectionStateFailed {
			// TODO peer connection cleanup
			fmt.Println("Peer Connection State has failed exiting")
			return
		}
	})
}

func acceptOffer(peerConnection *webrtc.PeerConnection, sessionDescription string) error {
	offer := webrtc.SessionDescription{}
	if err := json.NewDecoder(strings.NewReader(sessionDescription)).Decode(&offer); err != nil {
		return fmt.Errorf("failed to unmarshal offer: %w", err)
	}
	if err := peerConnection.SetRemoteDescription(offer); err != nil {
		return fmt.Errorf("failed to set remote description: %w", err)
	}
	return nil
}

func createAnswer(peerConnection *webrtc.PeerConnection) (answer webrtc.SessionDescription, err error) {
	answer, err = peerConnection.CreateAnswer(nil)
	if err != nil {
		return answer, fmt.Errorf("failed to create answer: %w", err)
	}
	if err := peerConnection.SetLocalDescription(answer); err != nil {
		return answer, fmt.Errorf("failed to set local description: %w", err)
	}
	return answer, nil
}

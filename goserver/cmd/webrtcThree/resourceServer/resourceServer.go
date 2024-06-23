package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media"
	"github.com/pion/webrtc/v4/pkg/media/h264reader"
)

var (
	h264FrameDuration = time.Millisecond * 33
	signallingServer  = "localhost:8124"
)

func main() {
	fmt.Println("resourceServer start")
	ctx := context.Background()
	if err := webrtcMain(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

type WebSocketMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type OfferData struct {
	Offer     webrtc.SessionDescription `json:"offer"`
	RequestID int32                     `json:"requestId"`
}

type AnswerData struct {
	Answer    webrtc.SessionDescription `json:"answer"`
	RequestID int32                     `json:"requestId"`
}

func (wsm *WebSocketMessage) isHeartBeat() bool {
	if wsm.Type == "heartbeat" {
		return true
	}
	return false
}

func (wsm *WebSocketMessage) parseOffer() (offer OfferData, err error) {
	if wsm.Type != "offer" {
		return offer, errors.New("invalid type " + wsm.Type)
	}
	if err := json.Unmarshal(wsm.Data, &offer); err != nil {
		return offer, fmt.Errorf("failed to unmarshal offer: %w", err)
	}
	return offer, nil
}

func createAnswerMessage(answer webrtc.SessionDescription, requestID int32) ([]byte, error) {
	data, err := json.Marshal(AnswerData{
		Answer:    answer,
		RequestID: requestID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal answer data: %w", err)
	}

	return json.Marshal(
		WebSocketMessage{
			Type: "answer",
			Data: json.RawMessage(data),
		})
}

func webrtcMain(ctx context.Context) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	fmt.Println("webrtcMain start")

	if err := checkVideoFile(); err != nil {
		return fmt.Errorf("failed to check video file: %w", err)
	}

	c, err := connectToWebsocket()
	if err != nil {
		return fmt.Errorf("failed to connect to websocket: %w", err)
	}
	defer c.Close()

	go func() {
		for {
			webSocketMessage, err := readMessage(c)
			if err != nil {
				fmt.Printf("failed to read message: %v\n", err)
				continue
			}

			if webSocketMessage.isHeartBeat() {
				continue
			}

			offer, err := webSocketMessage.parseOffer()
			if err != nil {
				fmt.Printf("failed to parse offer: %v\n", err)
				continue
			}

			answer, err := registerWebRTCEvents(ctx, offer.Offer)
			if err != nil {
				fmt.Printf("failed to register WebRTC events: %v\n", err)
				continue
			}
			answerMessage, err := createAnswerMessage(answer, offer.RequestID)
			if err != nil {
				fmt.Printf("failed to create answer message: %v\n", err)
				continue
			}
			if err := c.WriteMessage(websocket.TextMessage, answerMessage); err != nil {
				fmt.Printf("failed to write message: %v\n", err)
				continue
			}
		}
	}()

	<-ctx.Done()
	fmt.Println("webrtcMain end")
	return nil
}

var videoFileName = "resource/0518sample.mp4"

func checkVideoFile() error {
	if _, err := os.Stat(videoFileName); err != nil {
		return fmt.Errorf("failed to check video file: %w", err)
	}

	return nil
}

func registerWebRTCEvents(ctx context.Context, offer webrtc.SessionDescription) (answer webrtc.SessionDescription, err error) {
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
		return answer, fmt.Errorf("failed to accept offer: %w", err)
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

func acceptOffer(peerConnection *webrtc.PeerConnection, offer webrtc.SessionDescription) error {
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

func connectToWebsocket() (*websocket.Conn, error) {
	u := url.URL{Scheme: "ws", Host: signallingServer, Path: "/ws"}
	fmt.Printf("connecting to %s\n", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to websocket: %w", err)
	}
	return c, nil
}

func readMessage(c *websocket.Conn) (*WebSocketMessage, error) {
	_, message, err := c.ReadMessage()
	if err != nil {
		return nil, fmt.Errorf("failed to read message: %w", err)
	}
	log.Printf("recv: %s", message)
	var webSocketMessage WebSocketMessage
	if err := json.Unmarshal(message, &webSocketMessage); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w\n", err)
	}

	return &webSocketMessage, nil
}

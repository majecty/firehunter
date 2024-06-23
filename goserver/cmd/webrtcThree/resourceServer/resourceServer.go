package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media"
	"github.com/pion/webrtc/v4/pkg/media/h264reader"
)

var (
	h264FrameDuration = time.Millisecond * 33
	// signallingServer  = "localhost:8124"
	// signalScheme = "ws"
	signalScheme     = "wss"
	signallingServer = "signal-firehunter.i.juhyung.dev"
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
	Type     string          `json:"type"`
	Data     json.RawMessage `json:"data"`
	ClientID int32           `json:"clientId"`
}

type WebSocketData interface{}

type Offer = webrtc.SessionDescription
type Answer = webrtc.SessionDescription
type Candidate = webrtc.ICECandidateInit

func parseWebSocketMessage(message []byte) (*WebSocketMessage, WebSocketData, error) {
	var wsMessage WebSocketMessage
	if err := json.Unmarshal(message, &wsMessage); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	if wsMessage.Type == "offer" {
		var offer Offer
		if err := json.Unmarshal(wsMessage.Data, &offer); err != nil {
			return nil, nil, fmt.Errorf("failed to unmarshal offer: %w", err)
		}
		return &wsMessage, &offer, nil
	}

	if wsMessage.Type == "candidate" {
		var candidate Candidate
		if err := json.Unmarshal(wsMessage.Data, &candidate); err != nil {
			return nil, nil, fmt.Errorf("failed to unmarshal candidate: %w", err)
		}
		return &wsMessage, &candidate, nil
	}

	return nil, nil, fmt.Errorf("unknown message type: %s", wsMessage.Type)
}

func createAnswerMessage(answer Answer, clientID int32) ([]byte, error) {
	data, err := json.Marshal(answer)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal answer: %w", err)
	}

	message, err := json.Marshal(WebSocketMessage{
		Type:     "answer",
		Data:     data,
		ClientID: clientID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	return message, nil
}

func createCandidateMessage(candidate *webrtc.ICECandidate, clientID int32) ([]byte, error) {
	data, err := json.Marshal(candidate.ToJSON())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal candidate: %w", err)
	}

	message, err := json.Marshal(WebSocketMessage{
		Type:     "candidate",
		Data:     data,
		ClientID: clientID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	return message, nil
}

type Peer struct {
	ID int32
	pc *webrtc.PeerConnection
	c  *websocket.Conn
}

type Peers struct {
	peers map[int32]*Peer
	mu    sync.RWMutex
}

var (
	peers = Peers{peers: make(map[int32]*Peer)}
)

func getPeerConnection(clientID int32) (*webrtc.PeerConnection, error) {
	peers.mu.RLock()
	defer peers.mu.RUnlock()

	peer, ok := peers.peers[clientID]
	if !ok {
		return nil, fmt.Errorf("peer not found")
	}

	return peer.pc, nil
}

func addPeer(clientID int32, pc *webrtc.PeerConnection, c *websocket.Conn) {
	peers.mu.Lock()
	defer peers.mu.Unlock()

	peers.peers[clientID] = &Peer{ID: clientID, pc: pc, c: c}
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
			webSocketMessage, websocketData, err := readMessage(c)
			if err != nil {
				fmt.Printf("failed to read message: %v\n", err)
				continue
			}
			fmt.Printf("readMessage: %v\n", webSocketMessage.Type)

			switch data := websocketData.(type) {
			case *Candidate:
				pc, err := getPeerConnection(webSocketMessage.ClientID)
				if err != nil {
					fmt.Printf("failed to get peer connection: %v\n", err)
					continue
				}
				pc.AddICECandidate(*data)

			case *Offer:
				// TODO: need to cldanup peerConenction
				peerConnection, err := registerWebRTCEvents(ctx)
				if err != nil {
					fmt.Printf("failed to register WebRTC events: %v\n", err)
					continue
				}
				addPeer(webSocketMessage.ClientID, peerConnection, c)
				peerConnection.OnICECandidate(func(candidate *webrtc.ICECandidate) {
					if candidate == nil {
						return
					}

					candidateMessage, err := createCandidateMessage(candidate, webSocketMessage.ClientID)
					if err != nil {
						fmt.Printf("failed to create candidate message: %v\n", err)
						return
					}
					fmt.Println("send candidate message")
					if err := c.WriteMessage(websocket.TextMessage, candidateMessage); err != nil {
						fmt.Printf("failed to write message: %v\n", err)
						return
					}
				})

				if err := acceptOffer(peerConnection, *data); err != nil {
					// return nil, answer, fmt.Errorf("failed to accept offer: %w", err)
					fmt.Printf("failed to accept offer: %v\n", err)
					continue
				}

				answer, err := createAnswer(peerConnection)
				if err != nil {
					fmt.Printf("failed to create answer: %v\n", err)
					continue
					// return nil, answer, fmt.Errorf("failed to create answer: %w", err)
				}

				answerMessage, err := createAnswerMessage(answer, webSocketMessage.ClientID)
				if err != nil {
					fmt.Printf("failed to create answer message: %v\n", err)
					continue
				}
				fmt.Println("send answer message")
				if err := c.WriteMessage(websocket.TextMessage, answerMessage); err != nil {
					fmt.Printf("failed to write message: %v\n", err)
					continue
				}

			}
		}
	}()

	<-ctx.Done()
	fmt.Println("webrtcMain end")
	return nil
}

var videoFileName = "resource/0518sample_annexb.h264"

func checkVideoFile() error {
	if _, err := os.Stat(videoFileName); err != nil {
		return fmt.Errorf("failed to check video file: %w", err)
	}

	return nil
}

func registerWebRTCEvents(ctx context.Context) (peerConnection *webrtc.PeerConnection, err error) {
	fmt.Println("registerWebRTCEvents")
	peerConnection, err = createPeerConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to create PeerConnection: %w", err)
	}

	iceConnectedCtx, iceConnectedCtxCancel := context.WithCancel(ctx)

	videoTrack, videoTrackErr := createVideoTrack(peerConnection)
	if videoTrackErr != nil {
		return nil, fmt.Errorf("failed to create video track: %w", videoTrackErr)
	}

	go streamingVideo(iceConnectedCtx, videoTrack)

	registerConnectionStartedEvent(iceConnectedCtxCancel, peerConnection)
	registerConnectionFailedEvent(peerConnection)

	return peerConnection, nil
}

func createPeerConnection() (*webrtc.PeerConnection, error) {
	return webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			webrtc.ICEServer{
				URLs: []string{"stun:stun.i.juhyung.dev:3478"},
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
				// fmt.Printf("Failed to read RTCP packet: %v\n", rtcpErr)
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

	fmt.Println("streamingVideo wait for connection")
	// connection이 되길 기다림
	<-iceConnectedCtx.Done()
	fmt.Println("streamingVideo start")

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
	// fmt.Printf("acceptOffer: %v\n", offer)
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
	u := url.URL{Scheme: signalScheme, Host: signallingServer, Path: "/ws"}
	fmt.Printf("connecting to %s\n", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to websocket: %w", err)
	}
	return c, nil
}

func readMessage(c *websocket.Conn) (*WebSocketMessage, WebSocketData, error) {
	_, message, err := c.ReadMessage()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read message: %w", err)
	}

	wsMessage, data, err := parseWebSocketMessage(message)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	return wsMessage, data, nil
}

// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

// Package main implements a simple TURN server
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/pion/turn/v3"
	"github.com/pion/webrtc/v4"
	"github.com/rs/cors"
)

type SessionDescriptionRequest struct {
	Offer webrtc.SessionDescription `json:"offer"`
}

type SessionDescriptionResponse struct {
	Answer webrtc.SessionDescription `json:"answer"`
}

var (
	toWebSocket chan toWebSocketRequest
)

type toWebSocketRequest struct {
	offer        webrtc.SessionDescription
	responseChan chan toWebSocketResponse
}

type toWebSocketResponse struct {
	answer webrtc.SessionDescription
	error  error
}

type WebSocketSDRequest struct {
	Offer     string `json:"sessionDescription"`
	RequestId int32  `json:"requestId"`
}
type WebSocketMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type OfferData struct {
	Offer     webrtc.SessionDescription `json:"offer"`
	RequestID int32                     `json:"requestId"`
}

func createOfferMessage(offer webrtc.SessionDescription, requestId int32) ([]byte, error) {
	offerData := OfferData{
		Offer:     offer,
		RequestID: requestId,
	}

	data, err := json.Marshal(offerData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal offer: %w", err)
	}

	message, err := json.Marshal(WebSocketMessage{
		Type: "offer",
		Data: data,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	return message, nil
}

type AnswerData struct {
	Answer    webrtc.SessionDescription `json:"answer"`
	RequestID int32                     `json:"requestId"`
}

func parseAnswer(message []byte) (AnswerData, error) {
	var wsMessage WebSocketMessage
	if err := json.Unmarshal(message, &wsMessage); err != nil {
		return AnswerData{}, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	if wsMessage.Type != "answer" {
		return AnswerData{}, fmt.Errorf("invalid type %s", wsMessage.Type)
	}

	var answer AnswerData
	if err := json.Unmarshal(wsMessage.Data, &answer); err != nil {
		return AnswerData{}, fmt.Errorf("failed to unmarshal answer: %w", err)
	}

	return answer, nil
}

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	fmt.Println("turn server start goroutine")
	go runStunServer(ctx)
	fmt.Println("register WebSocket")
	registerWebSocketHandler(http.DefaultServeMux)
	fmt.Println("register http server")
	registerHTTPHandler(http.DefaultServeMux)
	fmt.Println("add cors")
	handler := cors.AllowAll().Handler(http.DefaultServeMux)

	if err := runHTTPServer(ctx, handler); err != nil {
		return fmt.Errorf("failed to run HTTP server: %w", err)
	}

	return nil
}

func runHTTPServer(ctx context.Context, handler http.Handler) error {
	server := &http.Server{
		Addr:    ":8124",
		Handler: handler,
	}
	go func() {
		fmt.Println("start http server at :8124")
		if err := server.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				fmt.Printf("HTTP server closed %v", err)
			} else {
				fmt.Printf("Failed to start server: %v", err)
			}
		}
	}()

	<-ctx.Done()
	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}

	return nil
}

type WaitingResponse struct {
	waitings map[int32]chan toWebSocketResponse
	mu       sync.Mutex
}

var waitngResponse = WaitingResponse{
	waitings: make(map[int32]chan toWebSocketResponse),
}

var addWaiter = func(requestId int32, ch chan toWebSocketResponse) {
	waitngResponse.mu.Lock()
	defer waitngResponse.mu.Unlock()
	waitngResponse.waitings[requestId] = ch
}

var consumeWaiter = func(requestId int32) (chan toWebSocketResponse, error) {
	waitngResponse.mu.Lock()
	defer waitngResponse.mu.Unlock()
	ch, ok := waitngResponse.waitings[requestId]
	if !ok {
		return nil, fmt.Errorf("no waiter found for requestId: %d", requestId)
	}
	return ch, nil
}

var websocketUpgrader = websocket.Upgrader{}

func registerWebSocketHandler(serverMux *http.ServeMux) {
	serverMux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("websocket connected")
		c, err := websocketUpgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println("Error upgrading websocket: ", err)
			return
		}
		defer c.Close()

		go func() {
			for {
				fromClient := <-toWebSocket
				requestId := rand.Int31()

				offerMessage, err := createOfferMessage(fromClient.offer, requestId)
				if err != nil {
					fmt.Println("Error creating offer message: ", err)
					continue
				}

				if err := c.WriteMessage(websocket.TextMessage, offerMessage); err != nil {
					fmt.Println("Error emitting clientSessionDescription: ", err)
				}

				addWaiter(requestId, fromClient.responseChan)
			}
		}()

		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				fmt.Println("Error reading message: ", err)
				break
			}

			fmt.Printf("recv: %s\n", message)
			answer, err := parseAnswer(message)
			if err != nil {
				fmt.Println("Error parsing answer: ", err)
				continue
			}

			responseSocket, err := consumeWaiter(answer.RequestID)
			if err != nil {
				fmt.Println("Error consuming waiter: ", err)
				return
			}

			responseSocket <- toWebSocketResponse{
				answer: answer.Answer,
				error:  nil,
			}
		}
	})
}

func registerHTTPHandler(serverMux *http.ServeMux) {
	serverMux.HandleFunc("/client/offer", func(w http.ResponseWriter, r *http.Request) {
		sessionDescription, err := decode[SessionDescriptionRequest](r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fmt.Println("sessionDescription: ", sessionDescription.Offer)
		response := make(chan toWebSocketResponse)
		toWebSocket <- toWebSocketRequest{
			offer:        sessionDescription.Offer,
			responseChan: response,
		}
		responseData := <-response
		if responseData.error != nil {
			http.Error(w, responseData.error.Error(), http.StatusInternalServerError)
			return
		}

		if err := encode(w, r, http.StatusOK, SessionDescriptionResponse{
			Answer: responseData.answer,
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}

func runStunServer(ctx context.Context) {
	port := 3478

	udpListener, err := net.ListenPacket("udp4", "0.0.0.0:"+strconv.Itoa(port))
	if err != nil {
		fmt.Printf("failed to create TURN server listener: %v\n", err)
		return
	}

	s, err := turn.NewServer(turn.ServerConfig{
		PacketConnConfigs: []turn.PacketConnConfig{
			{
				PacketConn: udpListener,
			},
		},
	})
	if err != nil {
		fmt.Printf("failed to create TURN server: %v\n", err)
		return
	}

	fmt.Println("Listening on", udpListener.LocalAddr())

	<-ctx.Done()

	if err = s.Close(); err != nil {
		fmt.Printf("failed to close TURN server: %v\n", err)
		return
	}
}

// https://grafana.com/blog/2024/02/09/how-i-write-http-services-in-go-after-13-years/
func encode[T any](w http.ResponseWriter, r *http.Request, status int, v T) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		return fmt.Errorf("failed to encode response: %w", err)
	}
	return nil
}

// https://grafana.com/blog/2024/02/09/how-i-write-http-services-in-go-after-13-years/
func decode[T any](r *http.Request) (T, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, fmt.Errorf("failed to decode request: %w", err)
	}
	return v, nil
}

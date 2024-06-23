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
	"github.com/rs/cors"
)

var (
	toResourceServer chan ResourceServerRequest = make(chan ResourceServerRequest)
)

type ResourceServerRequest struct {
	Request      ResourceServerWebSocketMessage
	ResponseChan chan ResourceServerResponse
}

type ResourceServerResponse struct {
	Response ResourceServerWebSocketMessage
	Error    error
}

type ClientWebSocketMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type ResourceServerWebSocketMessage struct {
	ClientID int32           `json:"clientId"`
	Type     string          `json:"type"`
	Data     json.RawMessage `json:"data"`
}

type WebSocketData interface{}

func parseClientWebSocketMessage(message []byte) (ClientWebSocketMessage, error) {
	var wsMessage ClientWebSocketMessage
	if err := json.Unmarshal(message, &wsMessage); err != nil {
		return ClientWebSocketMessage{}, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	return wsMessage, nil
}

func parseResourceServerWebSocketMessage(message []byte) (ResourceServerWebSocketMessage, error) {
	var wsMessage ResourceServerWebSocketMessage
	if err := json.Unmarshal(message, &wsMessage); err != nil {
		return ResourceServerWebSocketMessage{}, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	return wsMessage, nil
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
	registerReesourceServerWebsocketHandler(http.DefaultServeMux)
	fmt.Println("register http server")
	registerClientWebsocketHandler(http.DefaultServeMux)
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
	waitings map[int32]chan ResourceServerResponse
	mu       sync.Mutex
}

var waitngResponse = WaitingResponse{
	waitings: make(map[int32]chan ResourceServerResponse),
}

var addWaiter = func(requestId int32, ch chan ResourceServerResponse) {
	waitngResponse.mu.Lock()
	defer waitngResponse.mu.Unlock()
	waitngResponse.waitings[requestId] = ch
}

var consumeWaiter = func(requestId int32) (chan ResourceServerResponse, error) {
	waitngResponse.mu.Lock()
	defer waitngResponse.mu.Unlock()
	ch, ok := waitngResponse.waitings[requestId]
	if !ok {
		return nil, fmt.Errorf("no waiter found for requestId: %d", requestId)
	}
	return ch, nil
}

var websocketUpgrader = websocket.Upgrader{}

func registerReesourceServerWebsocketHandler(serverMux *http.ServeMux) {
	serverMux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("websocket connected")
		c, err := websocketUpgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println("Error upgrading websocket: ", err)
			return
		}
		defer c.Close()

		go func() {
			fmt.Println("wait for offer in websocket loop")
			for {
				fromClient := <-toResourceServer
				fmt.Println("send message to resource server")

				jsonMsg, err := json.Marshal(fromClient.Request)
				if err != nil {
					fmt.Println("Error marshaling clientSessionDescription: ", err)
					continue
				}
				if err := c.WriteMessage(websocket.TextMessage, jsonMsg); err != nil {
					fmt.Println("Error emitting clientSessionDescription: ", err)
				}

				addWaiter(fromClient.Request.ClientID, fromClient.ResponseChan)
			}
		}()

		for {
			_, message, err := c.ReadMessage()
			fmt.Println("received answer")
			if err != nil {
				fmt.Println("Error reading message: ", err)
				break
			}

			wsMessage, err := parseResourceServerWebSocketMessage(message)
			if err != nil {
				fmt.Println("Error parsing message: ", err)
				continue
			}

			responseSocket, err := consumeWaiter(wsMessage.ClientID)
			if err != nil {
				fmt.Println("Error consuming waiter: ", err)
				return
			}

			fmt.Println("send answer to http server")
			responseSocket <- ResourceServerResponse{
				Response: wsMessage,
				Error:    nil,
			}
		}
	})
}

func registerClientWebsocketHandler(serverMux *http.ServeMux) {
	serverMux.HandleFunc("/client/ws", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("client offer")
		c, err := websocketUpgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println("Error upgrading websocket: ", err)
			return
		}
		defer c.Close()

		clinetId := rand.Int31()
		response := make(chan ResourceServerResponse)

		go func() {
			for {
				_, message, err := c.ReadMessage()
				if err != nil {
					fmt.Println("Error reading message: ", err)
					break
				}

				wsMessage, err := parseClientWebSocketMessage(message)
				if err != nil {
					fmt.Println("Error parsing message: ", err)
					continue
				}

				toResourceServer <- ResourceServerRequest{
					Request: ResourceServerWebSocketMessage{
						ClientID: clinetId,
						Type:     wsMessage.Type,
						Data:     wsMessage.Data,
					},
					ResponseChan: response,
				}
			}
		}()

		for {
			responseData := <-response
			if responseData.Error != nil {
				fmt.Println("Error processing request: ", err)
				continue
			}

			responseStr, err := json.Marshal(responseData.Response)
			if err != nil {
				fmt.Println("Error marshaling response: ", err)
				continue
			}
			if err := c.WriteMessage(websocket.TextMessage, responseStr); err != nil {
				fmt.Println("Error emitting response: ", err)
			}
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

// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

// Package main implements a simple TURN server
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"

	"github.com/pion/turn/v3"
	"github.com/rs/cors"
	"github.com/zishang520/socket.io/socket"
)

type SessionDescriptionRequest struct {
	SessionDescription string `json:"sessionDescription"`
}

type SessionDescriptionResponse struct {
	SessionDescription string `json:"sessionDescription"`
}

var (
	toWebSocket chan toWebSocketRequest
)

type toWebSocketRequest struct {
	offer        string
	responseChan chan toWebSocketResponse
}

type toWebSocketResponse struct {
	answer string
	error  error
}

type WebSocketSDRequest struct {
	Offer     string `json:"sessionDescription"`
	RequestId int32  `json:"requestId"`
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
	go runTurnServer(ctx)
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

func registerWebSocketHandler(serverMux *http.ServeMux) {
	io := socket.NewServer(nil, nil)
	serverMux.Handle("/socket.io/", io.ServeHandler(nil))

	io.On("connection", func(clients ...any) {
		client := clients[0].(*socket.Socket)
		fmt.Println("connected:", client.Id())
		client.Emit("debugMessage", "connected using client.Emit")

		client.On("serverSessionDescription", func(datas ...any) {
			if len(datas) != 2 {
				fmt.Printf("serverSessionDescription: invalid number of arguments: %d\n", len(datas))
				return
			}
			requestId, requestIdConversion := datas[0].(int32)
			if !requestIdConversion {
				fmt.Printf("serverSessionDescription: requestId is not int %v\n", datas[0])
				return
			}
			data, dataConversion := datas[1].(string)
			if !dataConversion {
				fmt.Printf("serverSessionDescription: data is not string %v\n", datas[1])
				return
			}
			fmt.Println("serverSessionDescription: ", data)

			responseSocket, err := consumeWaiter(requestId)
			if err != nil {
				fmt.Println("Error consuming waiter: ", err)
				return
			}

			responseSocket <- toWebSocketResponse{
				answer: data,
				error:  nil,
			}
		})

		go func() {
			for {
				fromClient := <-toWebSocket
				requestId := rand.Int31()

				if err := client.Emit("clientSessionDescription", WebSocketSDRequest{
					Offer:     fromClient.offer,
					RequestId: requestId,
				}); err != nil {
					fmt.Println("Error emitting clientSessionDescription: ", err)
				}
				addWaiter(requestId, fromClient.responseChan)
			}
		}()
	})
}

func registerHTTPHandler(serverMux *http.ServeMux) {
	serverMux.HandleFunc("/client/sd", func(w http.ResponseWriter, r *http.Request) {
		sessionDescription, err := decode[SessionDescriptionRequest](r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fmt.Println("sessionDescription: ", sessionDescription.SessionDescription)
		response := make(chan toWebSocketResponse)
		toWebSocket <- toWebSocketRequest{
			offer:        sessionDescription.SessionDescription,
			responseChan: response,
		}
		responseData := <-response
		if responseData.error != nil {
			http.Error(w, responseData.error.Error(), http.StatusInternalServerError)
			return
		}

		if err := encode(w, r, http.StatusOK, SessionDescriptionResponse{
			SessionDescription: responseData.answer,
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}

func runTurnServer(ctx context.Context) {
	publicIP := "3.34.13.104"
	port := flag.Int("port", 3478, "Listening port.")
	realm := flag.String("realm", "turn.i.juhyung.dev", "Realm (defaults to \"pion.ly\")")
	flag.Parse()

	udpListener, err := net.ListenPacket("udp4", "0.0.0.0:"+strconv.Itoa(*port))
	if err != nil {
		fmt.Printf("failed to create TURN server listener: %v\n", err)
		return
	}

	s, err := turn.NewServer(turn.ServerConfig{
		Realm: *realm,
		AuthHandler: func(username string, realm string, srcAddr net.Addr) ([]byte, bool) { // nolint: revive
			fmt.Println("username: ", username)
			return []byte(username), true
		},

		PacketConnConfigs: []turn.PacketConnConfig{
			{
				PacketConn: udpListener,
				RelayAddressGenerator: &turn.RelayAddressGeneratorStatic{
					RelayAddress: net.ParseIP(publicIP), // Claim that we are listening on IP passed by user (This should be your Public IP)
					Address:      "0.0.0.0",             // But actually be listening on every interface
				},
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

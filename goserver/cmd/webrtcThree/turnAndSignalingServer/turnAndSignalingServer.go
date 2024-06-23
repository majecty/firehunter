// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

// Package main implements a simple TURN server
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/pion/turn/v3"
	"github.com/rs/cors"
	"github.com/zishang520/socket.io/socket"
)

type SessionDescriptionRequest struct {
	SessionDescription string `json:"sessionDescription"`
}

func main() {
	go runTurnServer()

	io := socket.NewServer(nil, nil)
	http.Handle("/socket.io/", io.ServeHandler(nil))

	io.On("connection", func(clients ...any) {
		client := clients[0].(*socket.Socket)
		fmt.Println("connected:", client.Id())
		client.Emit("debugMessage", "connected using client.Emit")
	})
	fmt.Println("Hello, worlsd.")
	handler := cors.AllowAll().Handler(http.DefaultServeMux)

	http.HandleFunc("/client/sd", func(w http.ResponseWriter, r *http.Request) {
		sessionDescription := SessionDescriptionRequest{}
		defer r.Body.Close()

		if err := json.NewDecoder(r.Body).Decode(&sessionDescription); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fmt.Println("sessionDescription: ", sessionDescription.SessionDescription)
	})

	http.ListenAndServe(":8478", handler)
	fmt.Println("Server started on port 8478.")
}

func runTurnServer() {
	publicIP := "3.34.13.104"
	port := flag.Int("port", 3478, "Listening port.")
	realm := flag.String("realm", "turn.i.juhyung.dev", "Realm (defaults to \"pion.ly\")")
	flag.Parse()

	udpListener, err := net.ListenPacket("udp4", "0.0.0.0:"+strconv.Itoa(*port))
	if err != nil {
		log.Panicf("Failed to create TURN server listener: %s", err)
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
		log.Panic(err)
	}

	fmt.Println("Listening on", udpListener.LocalAddr())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	if err = s.Close(); err != nil {
		log.Panic(err)
	}
}

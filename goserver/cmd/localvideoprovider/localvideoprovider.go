package main

import (
	"bytes"
	"fmt"
	"image"
	"log"
	"net"
	"net/http"
	"os"

	g "github.com/AllenDang/giu"
	"github.com/skip2/go-qrcode"
)

var (
	currentIp        string = ""
	websiteQR        []byte
	websiteQRTexture *g.Texture
)

func main() {
	if _, err := os.Stat("./goserver-root"); os.IsNotExist(err) {
		currentDir, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Current directory:", currentDir)
		log.Fatal("goserver-root file not found in the current directory")
	}

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Fatal(err)
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			log.Println("Private IP address:", ipnet.IP.String())
			currentIp = ipnet.IP.String()
			break
		}
	}

	go func() {
		fs := http.FileServer(http.Dir("./resource"))
		http.Handle("/", fs)

		log.Println("Server started on port 8080")
		err = http.ListenAndServe(":8080", nil)
		if err != nil {
			log.Fatal(err)
		}
	}()
	giuMain()
}

func giuMain() {
	wnd := g.NewMasterWindow("Hello world", 600, 600, g.MasterWindowFlagsNotResizable)
	websiteQR, err := qrcode.Encode("http://"+currentIp+":8080", qrcode.Medium, 512)
	if err != nil {
		log.Fatal(err)
	}
	img, _, err := image.Decode(bytes.NewReader(websiteQR))
	if err != nil {
		log.Fatal(err)
	}
	g.EnqueueNewTextureFromRgba(img, func(t *g.Texture) {
		websiteQRTexture = t
	})
	wnd.Run(loop)
}

func loop() {
	var layout g.Layout

	layout = g.Layout{
		g.Label("Hello world from giu"),
		g.Label("Server started on port 8080"),
		g.Label("Local IP address: " + currentIp),
		g.Image(websiteQRTexture).Size(256, 256),
		g.Button("DONE").OnClick(func() {
			fmt.Println("Im sooooooo cute!!")
		}),
	}
	g.SingleWindow().Layout(layout)
}

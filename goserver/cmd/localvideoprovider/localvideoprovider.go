package main

import (
	"log"
	"net"
	"net/http"
	"os"
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
			break
		}
	}

	fs := http.FileServer(http.Dir("./resource"))
	http.Handle("/", fs)

	log.Println("Server started on port 8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}

package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
)

func main() {
	if err := checkDirectory(); err != nil {
		currentDir, currentDirErr := os.Getwd()
		if currentDirErr != nil {
			fmt.Printf("error getting current directory while finding goserver-root: %v", currentDirErr)
		}
		fmt.Println("Current directory:", currentDir)
		fmt.Println(err)
	}

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Fatal(err)
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			log.Println("Private IP address:", ipnet.IP.String())
			// currentIp = ipnet.IP.String()
			break
		}
	}

	fvideos := http.FileServer(http.Dir("./resource/"))
	http.Handle("/", fvideos)
	err = http.ListenAndServe("0.0.0.0:8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func checkDirectory() error {
	if err := checkCurrentDirectory(); err != nil {
		return fmt.Errorf("error checking current directory: %w", err)
	}

	if err := checkResourceDirectory(); err != nil {
		return fmt.Errorf("error checking resource directory: %w", err)
	}

	return nil
}

func checkCurrentDirectory() error {
	if _, err := os.Stat("./goserver-root"); os.IsNotExist(err) {
		return fmt.Errorf("goserver-root file not found in the current directory %w", err)
	}

	return nil
}

func checkResourceDirectory() error {
	if _, err := os.Stat("./resource"); os.IsNotExist(err) {
		return fmt.Errorf("resource directory not found: %w", err)
	}
	return nil
}

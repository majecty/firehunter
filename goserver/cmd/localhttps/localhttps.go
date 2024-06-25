package main

// use dns server 3.34.13.104
// dig 192-168-1-2.i.juhyung.dev @3.34.13.104

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/rs/cors"
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

	// https server 192-168-1-2.i.juhyung.dev:8443

	// http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	// 	w.Write([]byte("Hello, World!"))
	// })

	fvideos := http.FileServer(http.Dir("./resource/"))
	http.Handle("/videos/", http.StripPrefix("/videos/", fvideos))
	fs := http.FileServer(http.Dir("./resource/root"))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.URL.Path)
		if strings.HasPrefix(r.URL.Path, "/seventh") {
			http.ServeFile(w, r, "./resource/root/index.html")
			return
		}
		if r.URL.Path != "/" {
			// fs.ServeHTTP(w, r)
			// http.StripPrefix("/")
			fs.ServeHTTP(w, r)
		} else {
			http.ServeFile(w, r, "./resource/root/index.html")
		}
	})

	handler := cors.AllowAll().Handler(http.DefaultServeMux)

	if err := http.ListenAndServeTLS("0.0.0.0:8443", "./resource/i.juhyung.dev/fullchain.pem", "./resource/i.juhyung.dev/privkey.pem", handler); err != nil {
		fmt.Printf("error starting server: %v", err)
	}

	println("Hello, World!")
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

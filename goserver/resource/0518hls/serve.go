/*
Serve is a very simple static file server in go
Usage:
	-p="8100": port to serve on
	-d=".":    the directory of static files to host
Navigating to http://localhost:8100 will display the index.html or directory
listing file.
*/
package main

import (
	"flag"
	"log"
	"net/http"
)

func main() {
	port := flag.String("p", "8000", "port to serve on")
	directory := flag.String("d", ".", "the directory of static file to host")
	flag.Parse()

	
	fileServer := http.FileServer(http.Dir(*directory))
	corsEnabledFileServer := addCORS(fileServer)

	http.Handle("/", corsEnabledFileServer)

	log.Printf("Serving %s on HTTP port: %s\n", *directory, *port)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}

// Middleware to add CORS headers
func addCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			return
		}

		// Serve the next handler
		next.ServeHTTP(w, r)
	})
}


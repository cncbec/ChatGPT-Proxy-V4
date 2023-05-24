package main

import (
	"cncbec/ChatGPT-Proxy-V4/api"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	http.HandleFunc("/api", api.api)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port

	log.Println("Starting ChatGPT At " + addr)

	srv := &http.Server{
		Addr:         addr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}

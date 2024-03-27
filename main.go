package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/elazarl/goproxy"
)


func main() {
	setupProxy()
	setupCommandServer()
}


func setupProxy() {
	proxy := goproxy.NewProxyHttpServer()
	//proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)
	proxy.Verbose = true // Set to true to see all proxy traffic

	// Use a single handler function for all requests
	proxy.OnRequest().DoFunc(
		func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			// Log the fact a request is being made. This will log both HTTP and HTTPS requests.
			// For HTTPS, the request URL will be the CONNECT method host:port.
			log.Printf("Request made to: %s", r.URL.String())
			return r, nil // Proceed with the default operation (forward the request)
		},
	)

	log.Println("Starting proxy server on :8080...")
	log.Fatal(http.ListenAndServe(":8080", proxy))
}

func setupCommandServer() {
	vpnClient := NewOpenVPNClient() // Initialize your VPN client

	http.HandleFunc("/start-vpn", func(w http.ResponseWriter, r *http.Request) {
		configPath := r.URL.Query().Get("config")
		if configPath == "" {
			http.Error(w, "Missing config parameter", http.StatusBadRequest)
			return
		}

		if err := vpnClient.Start(configPath); err != nil {
			log.Printf("Failed to start VPN: %v", err)
			http.Error(w, "Failed to start VPN", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "VPN started with config: %s", configPath)
	})

	http.HandleFunc("/stop-vpn", func(w http.ResponseWriter, r *http.Request) {
		if err := vpnClient.Stop(); err != nil {
			log.Printf("Failed to stop VPN: %v", err)
			http.Error(w, "Failed to stop VPN", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "VPN stopped")
	})

	log.Println("Starting command server on :8081...")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatalf("Failed to start command server: %v", err)
	}
}
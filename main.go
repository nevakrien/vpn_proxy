package main

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"os"
	"strconv"

	"github.com/elazarl/goproxy"
)

//const CONFIGS_DIR=filepath.Join("external","ovpn_configs")
var CONFIGS_DIR=filepath.Join("dockers","proxy0","ovpn_configs")
var configPaths []string
var cur_config=-1

func main() {
	setupVars()
	go func(){setupProxy()}()
	setupCommandServer()
}

func setupVars(){
	//vpn configs
	files, err := os.ReadDir(CONFIGS_DIR)
    if err != nil {
        fmt.Println("Error reading configs directory:", err)
        return
    }

    for _, file := range files {
        // Check if it's a file, not a directory
        if !file.IsDir() {
            // Construct the full path and add it to the slice
            fullPath := filepath.Join(CONFIGS_DIR, file.Name())
            configPaths = append(configPaths, fullPath)
        }
    }
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

	http.HandleFunc("/start-vpn-name", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
	        http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	        return
	    }

		configPath := r.Header.Get("config")
		if configPath == "" {
			http.Error(w, "Missing config parameter", http.StatusBadRequest)
			return
		}

		configPath=filepath.Join(CONFIGS_DIR,configPath)

		if err := vpnClient.Start(configPath); err != nil {
			log.Printf("Failed to start VPN: %v", err)
			http.Error(w, "Failed to start VPN", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "VPN started with config: %s", configPath)
	})

	http.HandleFunc("/start-vpn-int", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
	        http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	        return
	    }

		idxStr := r.Header.Get("idx")
		// Convert idx from string to int
		idx, err := strconv.Atoi(idxStr)
		if err != nil {
			// Handle the error, maybe return an HTTP error response
			// For example, if idxStr is not an integer
			http.Error(w, "Error parsing idx to integer", http.StatusBadRequest)
			return
		}

		configPath:=configPaths[idx]
		if configPath == "" {
			http.Error(w, "Missing config parameter", http.StatusBadRequest)
			return
		}

		configPath=filepath.Join(CONFIGS_DIR,configPath)

		if err := vpnClient.Start(configPath); err != nil {
			log.Printf("Failed to start VPN: %v", err)
			http.Error(w, "Failed to start VPN", http.StatusInternalServerError)
			return
		}

		cur_config=idx
		fmt.Fprintf(w, "VPN started with config: %s", configPath)

	})

	http.HandleFunc("/stop-vpn", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
	        http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	        return
	    }

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
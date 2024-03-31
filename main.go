package main

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"os"
	"strconv"
	
	"sync"
	"context"
	"errors"

	"github.com/elazarl/goproxy"
	"net"
	"time"
)

var AUTH_PATH=filepath.Join(".","dockers","proxy0","auth.txt")
//var CONFIGS_DIR=filepath.Join("external","ovpn_configs")
var CONFIGS_DIR=filepath.Join(".","dockers","proxy0","ovpn_configs")
var configPaths []string
var cur_config=-1

func main() {
	setupVars()
	//go func(){setupProxy()}()
	setupCommandServer()
	setupProxy()
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

type onCloseConn struct {
    net.Conn
    lastActivity time.Time
    idleFlag     bool
    mutex        sync.Mutex
    ctx          context.Context
    cancel       context.CancelFunc
}

func (c *onCloseConn) MarkActivity() {
    c.mutex.Lock()
    defer c.mutex.Unlock()
    c.lastActivity = time.Now()
    c.idleFlag = false
}

func (c *onCloseConn) CheckIdle(timeout time.Duration) {
    c.mutex.Lock()
    defer c.mutex.Unlock()
    if time.Since(c.lastActivity) > timeout {
        c.idleFlag = true
    }
}

func (c *onCloseConn) IsIdle() bool {
    c.mutex.Lock()
    defer c.mutex.Unlock()
    return c.idleFlag
}

func (c *onCloseConn) Write(b []byte) (n int, err error) {
    if c.IsIdle() {
        // Handle attempt to write to an idle connection
        return 0, errors.New("attempt to write to idle connection")
    }
    n, err = c.Conn.Write(b)
    if err == nil {
        c.MarkActivity()
    }
    return n, err
}


func (c *onCloseConn) Close() error {
    log.Println("Connection closed")
    return c.Conn.Close()
}

func (c *onCloseConn) idleChecker(idleTimeout time.Duration) {
    checkInterval := time.NewTicker(100 * time.Millisecond) // Adjust check frequency as needed
    defer checkInterval.Stop()

    for {
        select {
        case <-c.ctx.Done():
            // Context canceled, stop the goroutine
            return
        case <-checkInterval.C:
            // Perform idle check
            c.mutex.Lock()
            if time.Since(c.lastActivity) > idleTimeout && !c.idleFlag {
                log.Println("Connection marked as idle")
                c.idleFlag = true
                c.mutex.Unlock()
                return // Optionally close the connection here or take other action
            }
            c.mutex.Unlock()
        }
    }
}


func setupProxy() {//lock *sync.RWMutex
	proxy := goproxy.NewProxyHttpServer()
	//proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)
	proxy.Verbose = true // Set to true to see all proxy traffic

	customDial := func(network, addr string) (net.Conn, error) {
        conn, err := net.Dial(network, addr)
        if err != nil {
            return nil, err
        }

        ctx, cancel := context.WithCancel(context.Background())
	    wrappedConn := &onCloseConn{
	        Conn:        conn,
	        lastActivity: time.Now(),
	        ctx:          ctx,
	        cancel:       cancel,
	    }

        log.Println("Connection made (TCP)")
        go wrappedConn.idleChecker(1000 * time.Millisecond) // Adjust timeout as needed

        //return &onCloseConn{Conn: conn}, nil
        return wrappedConn, nil
    }

    // Set the custom dial function for HTTP traffic
    proxy.Tr = &http.Transport{
    	Dial: customDial,
    	DisableKeepAlives :true,
    	MaxIdleConns :0,
    	MaxIdleConnsPerHost: 0, 
    	IdleConnTimeout:     1 * time.Millisecond,
    }

    proxy.ConnectDial = customDial

	//proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)
	// Use a single handler function for all requests
	proxy.OnRequest().DoFunc(
		func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			// Log the fact a request is being made. This will log both HTTP and HTTPS requests.
			// For HTTPS, the request URL will be the CONNECT method host:port.
			log.Printf("Request made to: %s", r.URL.String())
			return r, nil // Proceed with the default operation (forward the request)
		},
	)

	proxy.OnResponse().DoFunc(func(r*http.Response,ctx *goproxy.ProxyCtx)*http.Response {
		log.Printf("Response made")
		return r
	})

	log.Println("Starting proxy server on :8080...")
	// go func() {
	// 	for {
	// 		time.Sleep(1 * time.Second) // Adjust the interval as needed
	// 		if proxy.Tr != nil {
	// 			proxy.Tr.CloseIdleConnections()
	// 			log.Println("Idle connections closed")
	// 		}
	// 	}
	// }()
	log.Fatal(http.ListenAndServe(":8080", proxy))
}

func setupCommandServer() {
	vpnClient := NewOpenVPNClient(AUTH_PATH) // Initialize your VPN client

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

		//configPath=filepath.Join(CONFIGS_DIR,configPath)

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

	go func(){
		log.Println("Starting command server on :8081...")
		if err := http.ListenAndServe(":8081", nil); err != nil {
			log.Fatalf("Failed to start command server: %v", err)
		}
	}()
}

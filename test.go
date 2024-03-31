package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	//"strings"
	//"net"
	//"time"
)


func main() {
	// URL of a relatively stable page for testing
	testURL := "https://openai.com/"
	commandStr := "http://localhost:8081/start-vpn-int"
	//stopCommandStr := "http://localhost:8081/stop-vpn"
	proxyStr := "http://localhost:8080" // Ensure the proxy scheme is included
	
	set_proxy(commandStr)
	proxyURL, err := url.Parse(proxyStr)
	if err != nil {
		fmt.Println("Error parsing proxy URL:", err)
		os.Exit(1)
	}


	// Fetch directly
	fmt.Println("Fetching direct content...")
	directContent, err := fetchContent(testURL, nil)
	if err != nil {
		fmt.Println("Error fetching direct content:", err)
		os.Exit(1)
	}

	// Fetch via proxy
	fmt.Println("Fetching content via proxy...")
	proxyContent, err := fetchContent(testURL, &http.Transport{Proxy: http.ProxyURL(proxyURL)})
	if err != nil {
		fmt.Println("Error fetching proxy content:", err)
		os.Exit(1)
	}

	// Compare the contents
	if directContent == proxyContent {
		fmt.Println("Success: Direct content matches proxy content.")
	} else {
		fmt.Println("Failure: Content mismatch between direct and proxy fetch.")
	}

	//set_proxy(stopCommandStr)
	//testHangUp(proxyURL)
}

func set_proxy(commandStr string){//,proxyStr string) (*url.URL){
	// Set up the proxy 
	// proxyURL, err := url.Parse(proxyStr)
	// if err != nil {
	// 	fmt.Println("Error parsing proxy URL:", err)
	// 	os.Exit(1)
	// }
	// Create a new request
	req, err := http.NewRequest("POST", commandStr, nil) // nil body since you're not sending form data
	if err != nil {
		fmt.Println("Error creating POST request:", err)
		os.Exit(1)
	}

	// Set idx in the header
	req.Header.Set("idx", "0")

	// Send the request
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making POST request to the command server:", err)
		os.Exit(1)
	}
	defer response.Body.Close()

	// Check response status
	if response.StatusCode != http.StatusOK {
		// Read the response body
		bodyBytes, err := ioutil.ReadAll(response.Body)
		if err != nil {
			fmt.Println("Failed to read response body:", err)
			os.Exit(1)
		}

		bodyString := string(bodyBytes)
		fmt.Printf("Failed to start VPN, status code: %d, response: %s\n", response.StatusCode, bodyString)
	} else {
		fmt.Println("VPN started successfully")
	}

	//return proxyURL
}

// fetchContent makes an HTTP GET request to the specified URL and returns the response body as a string.
// It accepts an optional http.Transport to allow for proxying.
func fetchContent(urlStr string, transport *http.Transport) (string, error) {
	client := &http.Client{}
	if transport != nil {
		client.Transport = transport
	}

	resp, err := client.Get(urlStr)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}


// func startLocalTCPServer() {
//     listener, err := net.Listen("tcp", "localhost:9999")
//     if err != nil {
//         fmt.Println("Failed to start server:", err)
//         os.Exit(1)
//     }
//     defer listener.Close()
//     fmt.Println("Server listening on localhost:9999")

//     // Accept a single connection then close it
//     conn, err := listener.Accept()
//     if err != nil {
//         fmt.Println("Failed to accept connection:", err)
//         return
//     }
//     fmt.Println("Accepted connection from:", conn.RemoteAddr().String())
//     conn.Close()
// }

// func openAndCloseTCPConnectionThroughProxy(proxyURL *url.URL, targetAddr string) error {
//     // This is the proxy address
//     proxyAddr := proxyURL.Host

//     // Directly dial the proxy
//     conn, err := net.Dial("tcp", proxyAddr)
//     if err != nil {
//         return fmt.Errorf("failed to dial proxy: %v", err)
//     }
//     defer conn.Close()

//     // Send a CONNECT request to the proxy for the target address
//     fmt.Fprintf(conn, "CONNECT %s HTTP/1.1\r\nHost: %s\r\n\r\n", targetAddr, targetAddr)

//     // Read the proxy's response
//     buffer := make([]byte, 4096)
//     _, err = conn.Read(buffer)
//     if err != nil {
//         return fmt.Errorf("failed to read response from proxy: %v", err)
//     }
//     fmt.Println("Opened and closed a TCP connection through the proxy to:", targetAddr)

//     return nil
// }

// func testHangUp(proxyURL *url.URL) {
//     // Start the local TCP server in a goroutine
//     go startLocalTCPServer()

//     // Small delay to ensure the server starts before the client tries to connect
//     time.Sleep(1 * time.Second)

//     // Target address is the local server
//     targetAddr := "localhost:9999"

//     // Connect to the local server through the proxy and close the connection
//     err := openAndCloseTCPConnectionThroughProxy(proxyURL, targetAddr)
//     if err != nil {
//         fmt.Println("Error:", err)
//         os.Exit(1)
//     }

// }
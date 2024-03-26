package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

func main() {
	// URL of a relatively stable page for testing
	testURL := "https://openai.com/"

	// Set up the proxy correctly
	proxyStr := "http://localhost:8080" // Ensure the proxy scheme is included
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

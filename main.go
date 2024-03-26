package main

import (
	"log"
	"net/http"

	"github.com/elazarl/goproxy"
)

func main() {
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

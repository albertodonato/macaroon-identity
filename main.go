package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"gopkg.in/macaroon-bakery.v2-unstable/httpbakery"
)

func main() {
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)
	s := NewService("localhost:8080", logger)
	goPanicOnError(s.Start)

	serverEndpoint := mustServe(func(endpoint string) (http.Handler, error) {
		return targetService(endpoint, s.Endpoint(), &s.KeyPair.Public)
	})
	resp, err := clientRequest(newClient(), serverEndpoint)
	if err != nil {
		log.Fatalf("client failed: %v", err)
	}
	fmt.Printf("client success: %q\n", resp)
}

func mustServe(newHandler func(string) (http.Handler, error)) (endpointURL string) {
	endpoint, err := serve(newHandler)
	if err != nil {
		log.Fatalf("cannot serve: %v", err)
	}
	return endpoint
}

func serve(newHandler func(string) (http.Handler, error)) (endpointURL string, err error) {
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", fmt.Errorf("cannot listen: %v", err)
	}
	endpointURL = "http://" + listener.Addr().String() + "/gold"
	handler, err := newHandler(endpointURL)
	if err != nil {
		return "", fmt.Errorf("cannot start handler: %v", err)
	}
	go http.Serve(listener, handler)
	return endpointURL, nil
}

func newClient() *httpbakery.Client {
	c := httpbakery.NewClient()
	c.AddInteractor(httpbakery.WebBrowserInteractor{})
	return c
}

func goPanicOnError(f func() error) {
	go func() {
		if err := f(); err != nil {
			panic(err)
		}
	}()
}

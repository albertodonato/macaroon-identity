package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"gopkg.in/macaroon-bakery.v2-unstable/httpbakery"

	"github.com/albertodonato/macaroon-identity/service"
)

func main() {
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)
	s := service.NewAuthService("localhost:8080", logger)
	if err := s.Start(true); err != nil {
		panic(err)
	}

	t := NewTargetService("localhost:0", s.Endpoint(), &s.KeyPair.Public, logger)
	if err := t.Start(true); err != nil {
		panic(err)
	}

	resp, err := clientRequest(newClient(), t.Endpoint())
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

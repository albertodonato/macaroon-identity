package main

import (
	"log"
	"os"

	"github.com/albertodonato/macaroon-identity/service"
)

func main() {
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)
	s := service.NewAuthService("localhost:0", logger)
	if err := s.Start(true); err != nil {
		panic(err)
	}

	t := NewTargetService("localhost:0", s.Endpoint(), &s.KeyPair.Public, logger)
	if err := t.Start(true); err != nil {
		panic(err)
	}

	resp, err := clientRequest("GET", t.Endpoint(), logger)
	if err == nil {
		logger.Printf("client response: %s", resp)
	} else {
		logger.Fatalf("client error: %v", err)
	}
}

package main

import (
	"log"
	"os"

	"github.com/albertodonato/macaroon-identity/service"
)

func main() {
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)
	s := service.NewAuthService("localhost:0", logger)
	s.Checker.AddCreds(map[string]string{"foo": "bar", "baz": "bza"})
	if err := s.Start(true); err != nil {
		panic(err)
	}

	t := NewTargetService("localhost:0", s.Endpoint(), &s.KeyPair.Public, logger)
	if err := t.Start(true); err != nil {
		panic(err)
	}

	clientRequest("GET", t.Endpoint(), Credentials{Username: "foo", Password: "bar"}, logger)
	clientRequest("GET", t.Endpoint(), Credentials{Username: "foo", Password: "invalid"}, logger)
	clientRequest("GET", t.Endpoint(), Credentials{Username: "baz", Password: "bza"}, logger)
}

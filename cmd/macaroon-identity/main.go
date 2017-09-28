package main

import (
	"flag"
	"log"
	"os"

	"github.com/albertodonato/macaroon-identity/service"
)

func main() {
	endpoint := flag.String("endpoint", "localhost:8080", "service endpoint")
	flag.Parse()

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)
	s := service.NewAuthService(*endpoint, logger)
	if err := s.Start(false); err != nil {
		panic(err)
	}
}

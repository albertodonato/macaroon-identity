package main

import (
	"flag"
	"log"
	"os"

	"github.com/albertodonato/macaroon-identity/service"
)

type flags struct {
	Endpoint string
	CredsFile string
}

func main() {
	flags := parseFlags()
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)
	s := service.NewAuthService(flags.Endpoint, logger)
	if err := s.Checker.LoadCreds(flags.CredsFile); err != nil {
		panic(err)
	}
	if err := s.Start(false); err != nil {
		panic(err)
	}
}

func parseFlags() flags {
	f := flags{
		Endpoint: *flag.String("endpoint", "localhost:8080", "service endpoint"),
		CredsFile: *flag.String("creds", "credentials.csv", "CSV file with credentials (username and password)"),
	}
	flag.Parse()
	return f
}

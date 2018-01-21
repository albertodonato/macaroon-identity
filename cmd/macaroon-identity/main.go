package main

import (
	"flag"
	"log"
	"os"

	"github.com/albertodonato/macaroon-identity/authservice"
)

type flags struct {
	Endpoint  string
	CredsFile string
	KeyFile   string
}

func main() {
	flags := parseFlags()
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)
	keyPair, err := authservice.GetKeyPair(flags.KeyFile)
	if err != nil {
		panic(err)
	}
	s := authservice.NewAuthService(flags.Endpoint, logger, keyPair)
	if err := s.Checker.LoadCreds(flags.CredsFile); err != nil {
		panic(err)
	}
	if err := s.Start(false); err != nil {
		panic(err)
	}
}

func parseFlags() *flags {
	endpoint := flag.String("endpoint", "localhost:8081", "service endpoint")
	credsFile := flag.String("creds", "credentials.csv", "CSV file with credentials (username and password)")
	keyFile := flag.String("keyfile", "", "JSON file containing the service public/private key pair.")
	flag.Parse()
	return &flags{
		Endpoint:  *endpoint,
		CredsFile: *credsFile,
		KeyFile:   *keyFile,
	}
}

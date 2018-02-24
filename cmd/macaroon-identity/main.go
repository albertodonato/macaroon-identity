package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/albertodonato/macaroon-identity/authservice"
)

type flags struct {
	Endpoint     string
	CredsFile    string
	KeyFile      string
	AuthValidity time.Duration
}

func main() {
	flags := parseFlags()
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)
	keyPair, err := authservice.GetKeyPair(flags.KeyFile)
	if err != nil {
		panic(err)
	}
	s := authservice.NewAuthService(flags.Endpoint, logger, keyPair, flags.AuthValidity)
	if err := s.Checker.LoadCreds(flags.CredsFile); err != nil {
		panic(err)
	}
	if err := s.Start(false); err != nil {
		panic(err)
	}
}

func parseFlags() *flags {
	endpoint := flag.String("endpoint", "localhost:8081", "service endpoint")
	credsFile := flag.String(
		"creds", "credentials.csv",
		"CSV file with credentials as (username,password[,group1 group2...])")
	keyFile := flag.String("keyfile", "", "JSON file containing the service public/private key pair")
	authValidity := flag.Duration("auth-validity", 5*time.Minute, "Duration of macaroon validity")
	flag.Parse()
	return &flags{
		Endpoint:     *endpoint,
		CredsFile:    *credsFile,
		KeyFile:      *keyFile,
		AuthValidity: *authValidity,
	}
}

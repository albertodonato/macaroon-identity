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

func parseFlags() (f flags) {
	flag.StringVar(&f.Endpoint, "endpoint", "localhost:8081", "service endpoint")
	flag.StringVar(&f.CredsFile, "creds", "credentials.csv",
		"CSV file with credentials as (username,password[,group1 group2...])")
	flag.StringVar(&f.KeyFile, "keyfile", "",
		"JSON file containing the service public/private key pair")
	flag.DurationVar(&f.AuthValidity, "auth-validity", 5*time.Minute,
		"Duration of macaroon validity")
	flag.Parse()
	return
}

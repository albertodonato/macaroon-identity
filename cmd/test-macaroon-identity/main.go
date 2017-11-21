package main

import (
	"flag"
	"log"
	"os"

	"github.com/albertodonato/macaroon-identity/service"

	"github.com/juju/loggo"
)

type flags struct {
	NoRequests bool
}

func main() {
	flags := parseFlags()

	if loggoLevel := os.Getenv("LOGGO"); loggoLevel != "" {
		loggo.ConfigureLoggers("<root>=" + loggoLevel)
	}

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)
	s := service.NewAuthService("localhost:0", logger)
	s.Checker.AddCreds(map[string]string{"foo": "bar", "baz": "bza"})
	if err := s.Start(true); err != nil {
		panic(err)
	}

	t := NewTargetService("localhost:0", s.Endpoint(), &s.KeyPair.Public, logger)
	if err := t.Start(!flags.NoRequests); err != nil {
		panic(err)
	}

	if !flags.NoRequests {
		makeTestRequests(t.Endpoint(), logger)
	}
}

func makeTestRequests(endpoint string, logger *log.Logger) {
	testCredentials := []Credentials{
		{Username: "foo", Password: "bar"},
		{Username: "foo", Password: "invalid"},
		{Username: "baz", Password: "bza"},
	}
	for _, credentials := range testCredentials {
		clientRequest("GET", endpoint, credentials, logger)
	}
}

func parseFlags() *flags {
	noRequests := flag.Bool("noreq", false, "don't perform test requests")
	flag.Parse()
	return &flags{
		NoRequests: *noRequests,
	}
}

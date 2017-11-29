package main

import (
	"flag"
	"log"
	"os"

	"github.com/juju/loggo"
	"gopkg.in/macaroon-bakery.v2/bakery"

	"github.com/albertodonato/macaroon-identity/service"
)

type flags struct {
	NoRequests bool
}

// Sample user/password credentials.
var sampleCredentials = map[string]string{
	"user1": "pass1",
	"user2": "pass2",
	"user3": "pass3",
}

// Sample user/groups mapping.
var sampleGroups = map[string][]string{
	"user1": {"group1", "group3"},
	"user2": {"group2"},
	"user3": {"group3"},
}

// Groups required by the target service for authenticating a user. A user can
// belong to any of the specified groups.
var requiredGroups = []string{"group1", "group2"}

func main() {
	flags := parseFlags()

	if loggoLevel := os.Getenv("LOGGO"); loggoLevel != "" {
		loggo.ConfigureLoggers("<root>=" + loggoLevel)
	}

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)
	s := service.NewAuthService("localhost:0", logger, bakery.MustGenerateKey())
	s.Checker.AddCreds(sampleCredentials)
	s.Checker.AddGroups(sampleGroups)
	if err := s.Start(true); err != nil {
		panic(err)
	}

	t := NewTargetService("localhost:0", s.Endpoint(), &s.KeyPair.Public, requiredGroups, logger)
	if err := t.Start(!flags.NoRequests); err != nil {
		panic(err)
	}

	if !flags.NoRequests {
		makeTestRequests(t.Endpoint(), logger)
	}
}

func makeTestRequests(endpoint string, logger *log.Logger) {
	testCredentials := []Credentials{
		{Username: "user1", Password: "pass1"},
		{Username: "user1", Password: "invalid"}, // invalid password
		{Username: "user2", Password: "pass2"},
		{Username: "user3", Password: "pass3"}, // valid creds but not in groups
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

package main

import (
	"flag"
	"log"
	"os"
	"regexp"

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
	s := setupAuthService(logger)
	t := setupTargetService(logger, s, !flags.NoRequests)

	if !flags.NoRequests {
		makeTestRequests(t.Endpoint(), logger)
	}
}

func parseFlags() *flags {
	noRequests := flag.Bool("noreq", false, "don't perform test requests")
	flag.Parse()
	return &flags{
		NoRequests: *noRequests,
	}
}

func errorExit(logger *log.Logger, format string, args ...interface{}) {
	logger.Printf(format+"\n", args...)
	os.Exit(1)
}

func setupAuthService(logger *log.Logger) *service.AuthService {
	s := service.NewAuthService("localhost:0", logger, bakery.MustGenerateKey())
	s.Checker.AddCreds(sampleCredentials)
	s.Checker.AddGroups(sampleGroups)
	if err := s.Start(true); err != nil {
		panic(err)
	}

	return s
}

func setupTargetService(logger *log.Logger, authService *service.AuthService, background bool) *TargetService {
	t := NewTargetService(
		"localhost:0", authService.Endpoint(), &authService.KeyPair.Public, requiredGroups, logger)
	if err := t.Start(background); err != nil {
		panic(err)
	}

	return t
}

type testScenario struct {
	Credentials       Credentials
	ResponseCode      int
	ErrorStringRegexp string
}

func makeTestRequests(endpoint string, logger *log.Logger) {
	testScenarios := []testScenario{
		{
			Credentials:  Credentials{Username: "user1", Password: "pass1"},
			ResponseCode: 200,
		},
		{
			// invalid password
			Credentials:       Credentials{Username: "user1", Password: "invalid"},
			ErrorStringRegexp: `invalid credentials`,
		},
		{
			Credentials:  Credentials{Username: "user2", Password: "pass2"},
			ResponseCode: 200,
		},
		{
			// valid credentials but not in groups
			Credentials:       Credentials{Username: "user3", Password: "pass3"},
			ErrorStringRegexp: `user not in required group\(s\)`,
		},
	}
	for _, scenario := range testScenarios {
		resp, err := clientRequest("GET", endpoint, scenario.Credentials, logger)
		if scenario.ResponseCode != 0 {
			if resp == nil {
				errorExit(
					logger,
					"expected response code %d, instead got error: %v",
					scenario.ResponseCode, err)
			}
			if resp.StatusCode != scenario.ResponseCode {
				errorExit(
					logger,
					"expected resposne code %d, instead got: %d",
					scenario.ResponseCode, resp.StatusCode)
			}
		}
		if scenario.ErrorStringRegexp != "" {
			if err == nil {
				errorExit(
					logger,
					`expected error maching regexp "%s", instead got nil`,
					scenario.ErrorStringRegexp)
			}

			if matched, _ := regexp.MatchString(scenario.ErrorStringRegexp, err.Error()); !matched {
				errorExit(
					logger,
					`error didn't match expected regexp "%s", instead got: %v`,
					scenario.ErrorStringRegexp, err)
			}
		}
	}
}

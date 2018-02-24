package main

import (
	"flag"
	"log"
	"os"

	"github.com/juju/loggo"
)

type flags struct {
	Username string
	Password string
	Path     string
	LogLevel string
}

func main() {
	flags := parseFlags()
	makeRequest := flags.Username != ""

	if flags.LogLevel != "" {
		loggo.ConfigureLoggers("<root>=" + flags.LogLevel)
	}

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)
	logSetup(logger)

	s := setupAuthService(logger)
	t := setupTargetService(logger, s, makeRequest)
	if makeRequest {
		creds := Credentials{Username: flags.Username, Password: flags.Password}
		makeTestRequest(logger, creds, t.Endpoint())
	}
}

func parseFlags() (f flags) {
	flag.StringVar(&f.Username, "auth-username", "", "authentication username")
	flag.StringVar(&f.Password, "auth-password", "", "authentication password")
	flag.StringVar(&f.Path, "path", "/", "request path")
	flag.StringVar(&f.LogLevel, "loglevel", "", "log level")
	flag.Parse()
	return
}

func logSetup(logger *log.Logger) {
	logger.Printf("valid credentials: %q", sampleCredentials)
	logger.Printf("user/group mapping: %q", sampleGroups)
	logger.Printf("required groups: %q", requiredGroups)
	logger.Printf("macaroon validity: %q", macaroonValidity)
}

func makeTestRequest(logger *log.Logger, creds Credentials, path string) {
	logger.Printf("cli  - req: GET %s %q", path, creds)
	statusCode, content, err := clientRequest("GET", path, creds)
	if err == nil {
		logger.Printf("cli  - resp: %d - %s", statusCode, content)
	} else {
		logger.Printf("cli  - err: %v", err)
	}
}

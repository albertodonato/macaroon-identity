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
	logUserGroupSetup(logger)

	s := setupAuthService(logger)
	t := setupTargetService(logger, s, makeRequest)
	if makeRequest {
		creds := Credentials{Username: flags.Username, Password: flags.Password}
		logger.Printf("cli  - req: GET %s %q", t.Endpoint(), creds)
		statusCode, content, err := clientRequest("GET", t.Endpoint(), creds)
		if err == nil {
			logger.Printf("cli  - resp: %d - %s", statusCode, content)
		} else {
			logger.Printf("cli  - err: %v", err)
		}
	}
}

func parseFlags() *flags {
	username := flag.String("username", "", "authentication username")
	password := flag.String("password", "", "authentication password")
	path := flag.String("path", "/", "request path")
	logLevel := flag.String("loglevel", "", "log level")
	flag.Parse()
	return &flags{
		Username: *username,
		Password: *password,
		Path:     *path,
		LogLevel: *logLevel,
	}
}

func logUserGroupSetup(logger *log.Logger) {
	logger.Printf("valid credentials: %q", sampleCredentials)
	logger.Printf("user/group mapping: %q", sampleGroups)
	logger.Printf("required groups: %q", requiredGroups)
}

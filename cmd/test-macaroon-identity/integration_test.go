// Integration tests.

package main

import (
	"log"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/albertodonato/macaroon-identity/authservice"
)

type IntegrationTestSuite struct {
	suite.Suite

	authService   *authservice.AuthService
	targetService *TargetService
}

func (s *IntegrationTestSuite) SetupSuite() {
	logger := log.New(voidWriter{}, "", 0)
	s.authService = setupAuthService(logger)
	s.targetService = setupTargetService(logger, s.authService, true)
}

// If credentials are valid the request to the target service succeeds.
func (s *IntegrationTestSuite) TestValidCredentials() {
	creds := Credentials{Username: "user1", Password: "pass1"}
	code, content, err := clientRequest("GET", s.targetService.Endpoint(), creds)
	s.Equal(200, code)
	s.Equal(`you requested URL "/"`, content)
	s.Nil(err)
}

// If credentials are invalid the request fails.
func (s *IntegrationTestSuite) TestInValidCredentials() {
	creds := Credentials{Username: "user1", Password: "invalid"}
	_, _, err := clientRequest("GET", s.targetService.Endpoint(), creds)
	s.Regexp(`invalid credentials`, err.Error())
}

// If Credentials are valid but the user is not part of required group, the request fails.
func (s *IntegrationTestSuite) TestNotGroupMember() {
	creds := Credentials{Username: "user3", Password: "pass3"}
	_, _, err := clientRequest("GET", s.targetService.Endpoint(), creds)
	s.Regexp(`user not in required group\(s\)`, err.Error())
}

func TestIntegration(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

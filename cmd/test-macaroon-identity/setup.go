package main

import (
	"log"

	"gopkg.in/macaroon-bakery.v2/bakery"

	"github.com/albertodonato/macaroon-identity/service"
)

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
var requiredGroups = []string{
	"group1",
	"group2",
}

// A writer that doesn't do anything.
type voidWriter struct{}

func (l voidWriter) Write(p []byte) (int, error) {
	return len(p), nil
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

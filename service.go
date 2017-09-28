package main

import (
	"fmt"
	"log"
	"net/http"

	"golang.org/x/net/context"

	"gopkg.in/errgo.v1"
	"gopkg.in/macaroon-bakery.v2-unstable/bakery"
	"gopkg.in/macaroon-bakery.v2-unstable/bakery/checkers"
	"gopkg.in/macaroon-bakery.v2-unstable/httpbakery"

	"github.com/juju/idmclient/params"
)

type Service struct {
	ListenAddr string
	KeyPair    *bakery.KeyPair

	logger *log.Logger
	mux    http.Handler
}

func NewService(listenAddr string, logger *log.Logger) *Service {
	key := bakery.MustGenerateKey()
	discharger := httpbakery.NewDischarger(
		httpbakery.DischargerParams{
			Key:     key,
			Checker: httpbakery.ThirdPartyCaveatCheckerFunc(thirdPartyChecker),
		})

	mux := http.NewServeMux()
	discharger.AddMuxHandlers(mux, "/")
	return &Service{
		ListenAddr: listenAddr,
		KeyPair:    key,
		logger:     logger,
		mux:        mux,
	}
}

func (s *Service) Endpoint() string {
	return "http://" + s.ListenAddr
}

func (s *Service) Start() error {
	s.logger.Printf("Authentication service running at %s", s.Endpoint())
	return http.ListenAndServe(s.ListenAddr, s.mux)
}

func thirdPartyChecker(ctx context.Context, req *http.Request, info *bakery.ThirdPartyCaveatInfo, token *httpbakery.DischargeToken) ([]checkers.Caveat, error) {
	cond, args, err := checkers.ParseCaveat(string(info.Condition))
	if err != nil {
		return nil, errgo.WithCausef(err, params.ErrBadRequest, "cannot parse caveat %q", info.Condition)
	}
	fmt.Println(cond, args)
	return []checkers.Caveat{httpbakery.SameClientIPAddrCaveat(req)}, nil
}

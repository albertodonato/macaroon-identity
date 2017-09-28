package service

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

type AuthService struct {
	HTTPService

	KeyPair *bakery.KeyPair
}

func NewAuthService(listenAddr string, logger *log.Logger) *AuthService {
	key := bakery.MustGenerateKey()
	discharger := httpbakery.NewDischarger(
		httpbakery.DischargerParams{
			Key:     key,
			Checker: httpbakery.ThirdPartyCaveatCheckerFunc(thirdPartyChecker),
		})

	mux := http.NewServeMux()
	discharger.AddMuxHandlers(mux, "/")
	return &AuthService{
		HTTPService: HTTPService{
			Name:       "authentication service",
			ListenAddr: listenAddr,
			Logger:     logger,
			Mux:        mux,
		},
		KeyPair: key,
	}
}

func thirdPartyChecker(ctx context.Context, req *http.Request, info *bakery.ThirdPartyCaveatInfo, token *httpbakery.DischargeToken) ([]checkers.Caveat, error) {
	cond, args, err := checkers.ParseCaveat(string(info.Condition))
	if err != nil {
		return nil, errgo.WithCausef(err, params.ErrBadRequest, "cannot parse caveat %q", info.Condition)
	}
	fmt.Println(">>>", cond, args)
	return []checkers.Caveat{httpbakery.SameClientIPAddrCaveat(req)}, nil
}

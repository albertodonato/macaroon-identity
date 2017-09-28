package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/context"

	"gopkg.in/errgo.v1"
	"gopkg.in/macaroon-bakery.v2-unstable/bakery"
	"gopkg.in/macaroon-bakery.v2-unstable/bakery/checkers"
	"gopkg.in/macaroon-bakery.v2-unstable/httpbakery"

	"github.com/albertodonato/macaroon-identity/service"
)

type TargetService struct {
	service.HTTPService

	authEndpoint string
	authKey      *bakery.PublicKey
	keyPair      *bakery.KeyPair
	bakery       *bakery.Bakery
}

func NewTargetService(endpoint string, authEndpoint string, authKey *bakery.PublicKey, logger *log.Logger) *TargetService {
	key := bakery.MustGenerateKey()

	locator := httpbakery.NewThirdPartyLocator(nil, nil)
	locator.AllowInsecure()
	b := bakery.New(bakery.BakeryParams{
		Key:      key,
		Location: endpoint,
		Locator:  locator,
		Checker:  httpbakery.NewChecker(),
		Authorizer: authorizer{
			thirdPartyLocation: authEndpoint,
		},
	})
	mux := http.NewServeMux()
	t := TargetService{
		HTTPService: service.HTTPService{
			Name:       "target service",
			ListenAddr: endpoint,
			Logger:     logger,
			Mux:        mux,
		},
		authEndpoint: authEndpoint,
		authKey:      authKey,
		keyPair:      key,
		bakery:       b,
	}
	mux.Handle("/", t.auth(http.HandlerFunc(t.serveURL)))
	return &t

}

func (t *TargetService) serveURL(w http.ResponseWriter, req *http.Request) {
	t.LogRequest(req)
	fmt.Fprintf(w, `you requested URL "%s"`, req.URL.Path)
}

// auth wraps the given handler with a handler that provides authorization by
// inspecting the HTTP request to decide what authorization is required.
func (t *TargetService) auth(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := httpbakery.ContextWithRequest(context.TODO(), req)
		ops, err := opsForRequest(req)
		if err != nil {
			t.Fail(w, http.StatusInternalServerError, "%v", err)
			return
		}
		authChecker := t.bakery.Checker.Auth(httpbakery.RequestMacaroons(req)...)
		if _, err = authChecker.Allow(ctx, ops...); err != nil {
			t.writeError(ctx, w, req, err)
			return
		}
		h.ServeHTTP(w, req)
	})
}

// opsForRequest returns the required operations implied by the given HTTP
// request.
func opsForRequest(req *http.Request) ([]bakery.Op, error) {
	if !strings.HasPrefix(req.URL.Path, "/") {
		return nil, errgo.Newf("bad path")
	}
	elems := strings.Split(req.URL.Path, "/")
	if len(elems) < 2 {
		return nil, errgo.Newf("bad path")
	}
	return []bakery.Op{{
		Entity: elems[1],
		Action: req.Method,
	}}, nil
}

// writeError writes an error to w in response to req. If the error was
// generated because of a required macaroon that the client does not have, we
// mint a macaroon that, when discharged, will grant the client the right to
// execute the given operation.
func (t *TargetService) writeError(ctx context.Context, w http.ResponseWriter, req *http.Request, verr error) {
	derr, ok := errgo.Cause(verr).(*bakery.DischargeRequiredError)
	if !ok {
		t.Fail(w, http.StatusForbidden, "%v", verr)
		return
	}
	// Mint an appropriate macaroon and send it back to the client.
	m, err := t.bakery.Oven.NewMacaroon(ctx, httpbakery.RequestVersion(req), time.Now().Add(5*time.Minute), derr.Caveats, derr.Ops...)
	if err != nil {
		t.Fail(w, http.StatusInternalServerError, "cannot mint macaroon: %v", err)
		return
	}

	herr := httpbakery.NewDischargeRequiredError(m, "/", derr, req)
	herr.(*httpbakery.Error).Info.CookieNameSuffix = "auth"
	httpbakery.WriteError(ctx, w, herr)
}

type authorizer struct {
	thirdPartyLocation string
}

// Authorize implements bakery.Authorizer.Authorize by
// allowing anyone to do anything if a third party
// approves it.
func (a authorizer) Authorize(ctx context.Context, id bakery.Identity, ops []bakery.Op) (allowed []bool, caveats []checkers.Caveat, err error) {
	allowed = make([]bool, len(ops))
	for i := range allowed {
		allowed[i] = true
	}
	caveats = []checkers.Caveat{{
		Location:  a.thirdPartyLocation,
		Condition: "access-allowed",
	}}
	return
}

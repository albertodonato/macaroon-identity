package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/context"
	"gopkg.in/errgo.v1"

	"gopkg.in/macaroon-bakery.v2-unstable/bakery"
	"gopkg.in/macaroon-bakery.v2-unstable/bakery/checkers"
	"gopkg.in/macaroon-bakery.v2-unstable/httpbakery"
)

type targetServiceHandler struct {
	bakery       *bakery.Bakery
	authEndpoint string
	endpoint     string
	mux          *http.ServeMux
}

// targetService implements a "target service", representing
// an arbitrary web service that wants to delegate authorization
// to third parties.
//
func targetService(endpoint, authEndpoint string, authPK *bakery.PublicKey) (http.Handler, error) {
	key, err := bakery.GenerateKey()
	if err != nil {
		return nil, err
	}
	pkLocator := httpbakery.NewThirdPartyLocator(nil, nil)
	pkLocator.AllowInsecure()
	b := bakery.New(bakery.BakeryParams{
		Key:      key,
		Location: endpoint,
		Locator:  pkLocator,
		Checker:  httpbakery.NewChecker(),
		Authorizer: authorizer{
			thirdPartyLocation: authEndpoint,
		},
	})
	mux := http.NewServeMux()
	srv := &targetServiceHandler{
		bakery:       b,
		authEndpoint: authEndpoint,
	}
	mux.Handle("/gold/", srv.auth(http.HandlerFunc(srv.serveGold)))
	mux.Handle("/silver/", srv.auth(http.HandlerFunc(srv.serveSilver)))
	return mux, nil
}

func (srv *targetServiceHandler) serveGold(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "all is golden")
}

func (srv *targetServiceHandler) serveSilver(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "every cloud has a silver lining")
}

// auth wraps the given handler with a handler that provides
// authorization by inspecting the HTTP request
// to decide what authorization is required.
func (srv *targetServiceHandler) auth(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := httpbakery.ContextWithRequest(context.TODO(), req)
		ops, err := opsForRequest(req)
		if err != nil {
			fail(w, http.StatusInternalServerError, "%v", err)
			return
		}
		authChecker := srv.bakery.Checker.Auth(httpbakery.RequestMacaroons(req)...)
		if _, err = authChecker.Allow(ctx, ops...); err != nil {
			srv.writeError(ctx, w, req, err)
			return
		}
		h.ServeHTTP(w, req)
	})
}

// opsForRequest returns the required operations
// implied by the given HTTP request.
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
// generated because of a required macaroon that the client does not
// have, we mint a macaroon that, when discharged, will grant the client
// the right to execute the given operation.
func (srv *targetServiceHandler) writeError(ctx context.Context, w http.ResponseWriter, req *http.Request, verr error) {
	derr, ok := errgo.Cause(verr).(*bakery.DischargeRequiredError)
	if !ok {
		fail(w, http.StatusForbidden, "%v", verr)
		return
	}
	// Mint an appropriate macaroon and send it back to the client.
	m, err := srv.bakery.Oven.NewMacaroon(ctx, httpbakery.RequestVersion(req), time.Now().Add(5*time.Minute), derr.Caveats, derr.Ops...)
	if err != nil {
		fail(w, http.StatusInternalServerError, "cannot mint macaroon: %v", err)
		return
	}

	herr := httpbakery.NewDischargeRequiredError(m, "/", derr, req)
	herr.(*httpbakery.Error).Info.CookieNameSuffix = "auth"
	httpbakery.WriteError(ctx, w, herr)
}

func fail(w http.ResponseWriter, code int, msg string, args ...interface{}) {
	http.Error(w, fmt.Sprintf(msg, args...), code)
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

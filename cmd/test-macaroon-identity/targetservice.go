package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"golang.org/x/net/context"

	"gopkg.in/juju/idmclient.v1"
	"gopkg.in/macaroon-bakery.v2/bakery"
	"gopkg.in/macaroon-bakery.v2/bakery/identchecker"
	"gopkg.in/macaroon-bakery.v2/httpbakery"

	"github.com/albertodonato/macaroon-identity/service"
)

const authLifeSpan time.Duration = 5 * time.Minute

// TargetService is an HTTP service which requires macaroon-based authentication.
type TargetService struct {
	service.HTTPService

	RequiredGroups []string

	authEndpoint string
	authKey      *bakery.PublicKey
	keyPair      *bakery.KeyPair
	bakery       *identchecker.Bakery
}

// NewTargetService returns a TargetService instance.
func NewTargetService(endpoint string, authEndpoint string, authKey *bakery.PublicKey, requiredGroups []string, logger *log.Logger) *TargetService {
	key := bakery.MustGenerateKey()

	locator := httpbakery.NewThirdPartyLocator(nil, nil)
	locator.AllowInsecure()

	idmClient, _ := idmclient.New(idmclient.NewParams{
		BaseURL: authEndpoint,
	})
	b := identchecker.NewBakery(identchecker.BakeryParams{
		Key:            key,
		Location:       endpoint,
		Locator:        locator,
		Checker:        httpbakery.NewChecker(),
		IdentityClient: idmClient,
		Authorizer: identchecker.ACLAuthorizer{
			GetACL: func(ctx context.Context, op bakery.Op) ([]string, bool, error) {
				return []string{identchecker.Everyone}, false, nil
			},
		},
	})
	mux := http.NewServeMux()
	t := TargetService{
		HTTPService: service.HTTPService{
			Name:       "serv",
			ListenAddr: endpoint,
			Logger:     logger,
			Mux:        mux,
		},
		RequiredGroups: requiredGroups,
		authEndpoint:   authEndpoint,
		authKey:        authKey,
		keyPair:        key,
		bakery:         b,
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
		ops := opsForRequest(req)
		authChecker := t.bakery.Checker.Auth(httpbakery.RequestMacaroons(req)...)
		if _, err := authChecker.Allow(ctx, ops...); err != nil {
			oven := httpbakery.Oven{Oven: t.bakery.Oven}
			httpbakery.WriteError(ctx, w, oven.Error(ctx, req, err))
			return
		}
		h.ServeHTTP(w, req)
	})
}

// opsForRequest returns the required operations implied by the given HTTP
// request.
func opsForRequest(r *http.Request) []bakery.Op {
	return []bakery.Op{{
		Entity: r.URL.Path,
		Action: r.Method,
	}}
}

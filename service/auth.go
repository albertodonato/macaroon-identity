package service

import (
	"fmt"
	"log"
	"net/http"

	"golang.org/x/net/context"

	"gopkg.in/macaroon-bakery.v2-unstable/bakery"
	"gopkg.in/macaroon-bakery.v2-unstable/bakery/checkers"
	"gopkg.in/macaroon-bakery.v2-unstable/httpbakery"
	"gopkg.in/macaroon-bakery.v2-unstable/httpbakery/form"

	"github.com/juju/httprequest"
)

const formURL string = "/form"
const formTokenContent string = "ok"

type loginResponse struct {
	Token *httpbakery.DischargeToken `json:"token"`
}

// AuthService is an HTTP service for authentication using macaroons.
type AuthService struct {
	HTTPService

	KeyPair *bakery.KeyPair
	Checker CredentialsChecker
}

// NewAuthService returns an AuthService
func NewAuthService(listenAddr string, logger *log.Logger) *AuthService {
	key := bakery.MustGenerateKey()
	discharger := httpbakery.NewDischarger(
		httpbakery.DischargerParams{
			Key:     key,
			Checker: httpbakery.ThirdPartyCaveatCheckerFunc(thirdPartyChecker),
		})

	mux := http.NewServeMux()
	discharger.AddMuxHandlers(mux, "/")
	s := AuthService{
		HTTPService: HTTPService{
			Name:       "auth",
			ListenAddr: listenAddr,
			Logger:     logger,
			Mux:        mux,
		},
		KeyPair: key,
		Checker: NewCredentialsChecker(),
	}

	mux.Handle(formURL, http.HandlerFunc(s.formHandler))
	return &s
}

func thirdPartyChecker(ctx context.Context, req *http.Request, info *bakery.ThirdPartyCaveatInfo, token *httpbakery.DischargeToken) ([]checkers.Caveat, error) {
	if token == nil {
		err := httpbakery.NewInteractionRequiredError(nil, req)
		err.SetInteraction("form", form.InteractionInfo{URL: formURL})
		return nil, err
	}

	if token.Kind != "form" || string(token.Value) != formTokenContent {
		return nil, fmt.Errorf("invalid token %#v", token)
	}

	_, _, err := checkers.ParseCaveat(string(info.Condition))
	if err != nil {
		return nil, fmt.Errorf("cannot parse caveat %q: %s", info.Condition, err)
	}
	return []checkers.Caveat{httpbakery.SameClientIPAddrCaveat(req)}, nil
}

func (s *AuthService) formHandler(w http.ResponseWriter, req *http.Request) {
	s.LogRequest(req)
	switch req.Method {
	case "GET":
		httprequest.WriteJSON(w, http.StatusOK, schemaResponse)
	case "POST":
		params := httprequest.Params{
			Response: w,
			Request:  req,
			Context:  context.TODO(),
		}
		loginRequest := form.LoginRequest{}
		if err := httprequest.Unmarshal(params, &loginRequest); err != nil {
			s.bakeryFail(w, "can't unmarshal login request")
			return
		}

		form, err := fieldsChecker.Coerce(loginRequest.Body.Form, nil)
		if err != nil {
			s.bakeryFail(w, "invalid login form data: %v", err)
			return
		}

		if !s.Checker.Check(form) {
			s.bakeryFail(w, "invalid credentials")
			return
		}

		loginResponse := loginResponse{
			Token: &httpbakery.DischargeToken{
				Kind:  "form",
				Value: []byte(formTokenContent),
			},
		}
		httprequest.WriteJSON(w, http.StatusOK, loginResponse)

	default:
		s.Fail(w, http.StatusMethodNotAllowed, "%s method not allowed", req.Method)
		return
	}
}

func (s *AuthService) bakeryFail(w http.ResponseWriter, msg string, args ...interface{}) {
	httpbakery.WriteError(context.TODO(), w, fmt.Errorf(msg, args...))
}

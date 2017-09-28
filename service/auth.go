package service

import (
	"log"
	"net/http"

	"golang.org/x/net/context"

	"gopkg.in/errgo.v1"
	"gopkg.in/macaroon-bakery.v2-unstable/bakery"
	"gopkg.in/macaroon-bakery.v2-unstable/bakery/checkers"
	"gopkg.in/macaroon-bakery.v2-unstable/httpbakery"
	"gopkg.in/macaroon-bakery.v2-unstable/httpbakery/form"

	"github.com/juju/httprequest"
	"github.com/juju/idmclient/params"
)

const formURL string = "/form"

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
	s := AuthService{
		HTTPService: HTTPService{
			Name:       "authentication service",
			ListenAddr: listenAddr,
			Logger:     logger,
			Mux:        mux,
		},
		KeyPair: key,
	}

	mux.Handle(formURL, http.HandlerFunc(s.formHandler))
	return &s
}

func thirdPartyChecker(ctx context.Context, req *http.Request, info *bakery.ThirdPartyCaveatInfo, token *httpbakery.DischargeToken) ([]checkers.Caveat, error) {
	_, _, err := checkers.ParseCaveat(string(info.Condition))
	if err != nil {
		return nil, errgo.WithCausef(err, params.ErrBadRequest, "cannot parse caveat %q", info.Condition)
	}
	return []checkers.Caveat{httpbakery.SameClientIPAddrCaveat(req)}, nil
}

func (s *AuthService) formHandler(w http.ResponseWriter, req *http.Request) {
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
			s.Fail(w, http.StatusBadRequest, "can't unmarshal login request") // XXX return a proper json error
		}

		form, err := fieldsChecker.Coerce(loginRequest.Body.Form, nil)
		if err != nil {
			s.Fail(w, http.StatusBadRequest, "invalid login form data: %v", err) // XXX return a proper json error
			return
		}

		m := form.(map[string]interface{})
		username := m["username"].(string)
		password := m["password"].(string)
		s.Logger.Printf("login data: %s %s", username, password) // XXX handle authentication
	default:
		s.Fail(w, http.StatusMethodNotAllowed, "%s method not allowed", req.Method)
	}
}

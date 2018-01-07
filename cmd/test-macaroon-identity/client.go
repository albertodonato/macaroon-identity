package main

import (
	"io/ioutil"
	"net/http"

	schemaform "gopkg.in/juju/environschema.v1/form"
	"gopkg.in/macaroon-bakery.v2/httpbakery"
	"gopkg.in/macaroon-bakery.v2/httpbakery/form"
)

// Credentials for a user
type Credentials struct {
	Username string
	Password string
}

// BatchFiller is a form.Filler which uses the provided Credentials.
type BatchFiller struct {
	Credentials Credentials
}

// Fill fills the Form with the filler credentials
func (f *BatchFiller) Fill(form schemaform.Form) (map[string]interface{}, error) {
	return map[string]interface{}{
		"username": f.Credentials.Username,
		"password": f.Credentials.Password,
	}, nil
}

func clientRequest(method string, endpoint string, creds Credentials) (statusCode int, content string, err error) {
	client := newClient(creds)
	req, err := http.NewRequest(method, endpoint, nil)
	if err != nil {
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	statusCode = resp.StatusCode
	data, err := ioutil.ReadAll(resp.Body)
	if err == nil {
		content = string(data)
	}
	return
}

func newClient(creds Credentials) *httpbakery.Client {
	c := httpbakery.NewClient()
	c.AddInteractor(form.Interactor{Filler: &BatchFiller{Credentials: creds}})
	return c
}

package main

import (
	"io/ioutil"
	"log"
	"net/http"

	schemaform "gopkg.in/juju/environschema.v1/form"
	"gopkg.in/macaroon-bakery.v2-unstable/httpbakery"
	"gopkg.in/macaroon-bakery.v2-unstable/httpbakery/form"
)

type Credentials struct {
	Username string
	Password string
}

type BatchFiller struct {
	Credentials Credentials
}

func (f *BatchFiller) Fill(form schemaform.Form) (map[string]interface{}, error) {
	return map[string]interface{}{
		"username": f.Credentials.Username,
		"password": f.Credentials.Password,
	}, nil
}

func clientRequest(method string, endpoint string, creds Credentials, logger *log.Logger) error {
	client := newClient(creds)
	req, err := http.NewRequest(method, endpoint, nil)
	if err != nil {
		return err
	}

	logger.Printf("cli  - %s %s with creds %q", method, endpoint, creds)
	resp, err := client.Do(req)
	if err != nil {
		logger.Printf("cli  - got error: %v", err)
		return err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err == nil {
		logger.Printf("cli  - got response: %s", string(data))
	} else {
		logger.Printf("cli  - got error: %v", err)
	}
	return err
}

func newClient(creds Credentials) *httpbakery.Client {
	c := httpbakery.NewClient()
	c.AddInteractor(form.Interactor{Filler: &BatchFiller{Credentials: creds}})
	return c
}

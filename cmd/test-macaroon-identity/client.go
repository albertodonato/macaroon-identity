package main

import (
	"io/ioutil"
	"log"
	"net/http"

	"gopkg.in/macaroon-bakery.v2-unstable/httpbakery"
)

func clientRequest(method string, endpoint string, logger *log.Logger) (string, error) {
	client := newClient()
	req, err := http.NewRequest(method, endpoint, nil)
	if err != nil {
		return "", err
	}

	logger.Printf("client requesting %s %s", method, endpoint)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func newClient() *httpbakery.Client {
	c := httpbakery.NewClient()
	c.AddInteractor(httpbakery.WebBrowserInteractor{})
	return c
}

package main

import (
	"log"
	"net"
	"net/http"
)

type HTTPService struct {
	Name       string
	ListenAddr string

	logger   *log.Logger
	mux      http.Handler
	listener net.Listener
}

func (s *HTTPService) Endpoint() string {
	if s.listener == nil {
		return ""
	}

	return "http://" + s.listener.Addr().String()
}

func (s *HTTPService) Start() error {
	listener, err := net.Listen("tcp", s.ListenAddr)
	if err != nil {
		return err
	}
	s.listener = listener
	if s.Name != "" {
		s.logger.Printf("%s running at %s", s.Name, s.Endpoint())
	}
	go func() {
		err := http.Serve(s.listener, s.mux)
		if err != nil {
			panic(err)
		}
	}()
	
	return nil
}

package service

import (
	"fmt"
	"log"
	"net"
	"net/http"
)

// HTTPService represent an HTTP service.
type HTTPService struct {
	Name       string
	ListenAddr string
	Logger     *log.Logger
	Mux        http.Handler

	listener net.Listener
}

// Endpoint returns the URL of the service.
func (s *HTTPService) Endpoint() string {
	if s.listener == nil {
		return ""
	}

	return "http://" + s.listener.Addr().String()
}

// Start starts the servicem either in background or foreground.
func (s *HTTPService) Start(background bool) error {
	listener, err := net.Listen("tcp", s.ListenAddr)
	if err != nil {
		return err
	}
	s.listener = listener
	if s.Name != "" {
		s.Logger.Printf("%s - running at %s", s.Name, s.Endpoint())
	}

	if !background {
		return http.Serve(s.listener, s.Mux)
	}

	go func() {
		err := http.Serve(s.listener, s.Mux)
		if err != nil {
			panic(err)
		}
	}()
	return nil
}

// LogRequest logs an HTTP request.
func (s *HTTPService) LogRequest(req *http.Request) {
	s.Logger.Printf("%s - %s %s", s.Name, req.Method, req.URL.Path)
}

// Fail returns an HTTP error with the specified message
func (s *HTTPService) Fail(w http.ResponseWriter, code int, msg string, args ...interface{}) {
	http.Error(w, fmt.Sprintf(msg, args...), code)
}

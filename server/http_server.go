package server

import (
	"errors"
	"net/http"
)

// StartHTTPServer starts a generic HTTP server and returns it plus an error channel.
func StartHTTPServer(addr string, handler http.Handler) (*http.Server, <-chan error) {
	httpServer := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	errCh := make(chan error, 1)
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	return httpServer, errCh
}

package server

import (
	"fmt"
	"io/fs"
	"net/http"
	"path"

	webassets "aeibi/web"
)

// NewFrontendHandler serves the embedded frontend build output from web/dist.
func NewFrontendHandler() (http.Handler, error) {
	distFS, err := fs.Sub(webassets.DistFS, "dist")
	if err != nil {
		return nil, fmt.Errorf("sub embedded dist fs: %w", err)
	}

	fileServer := http.FileServer(http.FS(distFS))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if path.Ext(path.Clean("/"+r.URL.Path)) == "" {
			http.ServeFileFS(w, r, distFS, "index.html")
			return
		}

		fileServer.ServeHTTP(w, r)
	}), nil
}

// StartFrontendServer starts an HTTP server for the embedded frontend assets.
func StartFrontendServer(addr string) (*http.Server, <-chan error, error) {
	handler, err := NewFrontendHandler()
	if err != nil {
		return nil, nil, err
	}

	httpServer, errCh := StartHTTPServer(addr, handler)
	return httpServer, errCh, nil
}

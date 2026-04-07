package bootstrap

import (
"net/http"
"time"
)

type HTTPServer struct {
inner *http.Server
}

func NewHTTPServer(addr string, handler http.Handler) *HTTPServer {
return &HTTPServer{inner: &http.Server{Addr: addr, Handler: handler, ReadHeaderTimeout: 5 * time.Second}}
}

func (s *HTTPServer) Start() error { return s.inner.ListenAndServe() }

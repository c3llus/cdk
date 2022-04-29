package http

import (
	"net"
	"net/http"

	"github.com/gorilla/mux"
)

type Service struct {
	router *Router
	server *http.Server
}

type ServerHandler interface {
	RegisterHandler(r *Router)
}

// New to create a new http service
func New() *Service {

	// creates router
	r := &Router{
		router: mux.NewRouter(),
	}

	svc := &Service{
		server: &http.Server{
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
		},
		router: r,
	}
	return svc
}

// RegisterService execute services' RegisterHandler
func (svc *Service) RegisterService(services ...ServerHandler) {
	for _, s := range services {
		s.RegisterHandler(svc.router)
	}
	svc.server.Handler = svc.router.router
}

func (svc *Service) Serve(ls net.Listener) error {
	return svc.Server().Serve(ls)
}

func (svc Service) Server() *http.Server {
	return svc.server
}

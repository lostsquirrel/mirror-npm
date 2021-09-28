package app

import (
	"context"
	"log"
	"mirror-npm/handlers"
	"mirror-npm/utils"
	"net/http"
	"time"
)

type Config struct {
	addr string
}

type Instance struct {
	httpServer *http.Server
	handler    http.HandlerFunc
	config     Config
}

func NewInstance() *Instance {

	s := &Instance{
		// just in case you need some setup here
		handler: handlers.Handler,
		config: Config{
			addr: utils.GetAddr(),
		},
	}

	return s
}

func (s *Instance) Start() { // Startup all dependencies

	s.httpServer = &http.Server{Addr: s.config.addr, Handler: s.handler}

	err := s.httpServer.ListenAndServe()
	if err != http.ErrServerClosed {
		log.Println("Http Server stopped unexpected")
		s.Shutdown()
	} else {
		log.Println("Http Server stopped")
	}
}

func (s *Instance) Shutdown() {
	if s.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		err := s.httpServer.Shutdown(ctx)
		if err != nil {
			log.Println("Failed to shutdown http server gracefully")
		} else {
			s.httpServer = nil
		}
	}
}

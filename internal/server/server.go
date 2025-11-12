package server

import (
	"fmt"
	"ndk/internal/config"
	"net/http"
)

type Server struct {
	cfg *config.Config
}

func New(cfg *config.Config) *Server {
	return &Server{cfg: cfg}
}

func (s *Server) Start() error {
	r := NewRouter()
	return http.ListenAndServe(fmt.Sprintf(":%s", s.cfg.ServerPort), r)
}

package server

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"parserTest/internal/lamoda/config"
	"parserTest/internal/lamoda/parser"
)

type (
	Server struct {
		Debug  *http.Server
		Cfg    *config.Config
		Parser *parser.Parser
	}
)

func New(parser *parser.Parser, cfg *config.Config) *Server {
	return &Server{
		Debug: &http.Server{
			Addr: cfg.PrivatePort,
		},
		// Public:   new(fasthttp.Server),
		Cfg:    cfg,
		Parser: parser,
	}
}

func (s *Server) ParseLamoda(ctx context.Context, mainWg *sync.WaitGroup) {
	if err := s.Parser.ParseLamodaBySku(ctx); err != nil {
		fmt.Println(err)
	}
	mainWg.Done()
}

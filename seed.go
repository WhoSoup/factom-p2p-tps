package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"
)

type SeedServer struct {
	port  string
	seeds []string
}

func NewSeedServer(port, seeds string) *SeedServer {
	srv := new(SeedServer)
	srv.port = port

	split := strings.Split(seeds, ",")
	srv.seeds = make([]string, 0, len(split))
	for _, s := range split {
		s = strings.TrimSpace(s)
		if s != "" {
			srv.seeds = append(srv.seeds, s)
		}
	}
	return srv
}

func (s *SeedServer) Run() {
	mux := http.NewServeMux()
	mux.HandleFunc("/seed.txt", func(rw http.ResponseWriter, req *http.Request) {
		for _, s := range s.seeds {
			fmt.Fprintln(rw, s)
		}
	})
	log.Info().Str("url", fmt.Sprintf("http://localhost:%s/seed.txt", s.port)).Msg("Starting seed server")
	log.Error().Err(http.ListenAndServe(fmt.Sprintf(":%s", s.port), mux))
}

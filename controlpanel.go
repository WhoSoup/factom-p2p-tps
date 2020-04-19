package main

import (
	"fmt"
	"net/http"
)

type ControlPanel struct {
	port string
}

func NewControlPanel(port string) (*ControlPanel, error) {
	cp := new(ControlPanel)
	cp.port = port

	return cp, nil
}

func (cp *ControlPanel) Launch() error {
	mux := http.NewServeMux()

	return http.ListenAndServe(fmt.Sprintf(":%s", cp.port), mux)
}

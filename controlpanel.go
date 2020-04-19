package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"sync"

	"github.com/WhoSoup/factom-p2p-tps/network"
)

type ControlPanel struct {
	port       string
	n          network.Network
	template   *template.Template
	netEnabled bool
	enabler    sync.Once
}

func NewControlPanel(port string, n network.Network) (*ControlPanel, error) {

	template, err := template.ParseGlob("templates/*.html")
	if err != nil {
		return nil, err
	}

	cp := new(ControlPanel)
	cp.port = port
	cp.template = template
	cp.n = n

	return cp, nil
}

func (cp *ControlPanel) Launch() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/", cp.index)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	mux.HandleFunc("/enable", cp.enable)

	return http.ListenAndServe(fmt.Sprintf(":%s", cp.port), mux)
}

var validProtocols = []string{"p2p1-v9", "p2p2-v9", "p2p2-v10", "p2p2-v11"}

type settings struct {
	Name, P2PPort, Protocol, Seed, SeedStart, SeedPort, SeedContent string
}

func (cp *ControlPanel) verify(s settings) error {
	if len(s.Name) == 0 {
		return fmt.Errorf("name empty")
	}

	if _, err := strconv.Atoi(s.P2PPort); err != nil {
		return err
	}

	found := false
	for _, valid := range validProtocols {
		if s.Protocol == valid {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("invalid protocol specified \"%s\"", s.Protocol)
	}

	if len(s.Seed) == 0 {
		return fmt.Errorf("no seed server specified")
	}

	if s.SeedStart == "true" {
		if _, err := strconv.Atoi(s.SeedPort); err != nil {
			return err
		}
		if len(s.SeedContent) == 0 {
			return fmt.Errorf("no seed server specified")
		}
	}

	return nil
}

func (cp *ControlPanel) index(rw http.ResponseWriter, r *http.Request) {
	tpl := template.Must(template.ParseGlob("templates/*.html"))
	if err := tpl.ExecuteTemplate(rw, "index.html", map[string]interface{}{
		"enabled": cp.netEnabled,
	}); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
}

func (cp *ControlPanel) enable(rw http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	set := settings{
		Name:        r.FormValue("name"),
		P2PPort:     r.FormValue("p2pport"),
		Protocol:    r.FormValue("protocol"),
		Seed:        r.FormValue("seed"),
		SeedStart:   r.FormValue("seed-start"),
		SeedPort:    r.FormValue("seed-port"),
		SeedContent: r.FormValue("seed-content"),
	}
	if err := cp.verify(set); err != nil {
		http.Error(rw, err.Error(), http.StatusNotAcceptable)
		return
	}

	cp.enabler.Do(func() {
		cp.netEnabled = true
		go cp.n.Start()
	})

	http.Redirect(rw, r, "/", http.StatusSeeOther)
}

package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"sync"

	"github.com/WhoSoup/factom-p2p-tps/app"

	"github.com/WhoSoup/factom-p2p-tps/network"
	"github.com/rs/zerolog/log"
)

type ControlPanel struct {
	port       string
	n          network.Network
	template   *template.Template
	netEnabled bool
	eps        int
	enabler    sync.Once
	app        *app.App
}

func NewControlPanel(port string) (*ControlPanel, error) {

	template, err := template.ParseGlob("templates/*.html")
	if err != nil {
		return nil, err
	}

	cp := new(ControlPanel)
	cp.port = port
	cp.template = template
	cp.app = app.NewApp()
	return cp, nil
}

func (cp *ControlPanel) createNetwork(s settings) error {
	if err := cp.verify(s); err != nil {
		return err
	}

	var n network.Network
	switch s.Protocol {
	case "p2p1-v9":
		n = network.NewV9()
	case "p2p2-v9":
		n = network.NewV10(9)
	case "p2p2-v10":
		n = network.NewV10(10)
	case "p2p2-v11":
		n = network.NewV10(11)
	default:
		log.Fatal().Msg("protocol verify fail")
	}

	n.Init(s.Name, s.P2PPort, s.Seed)

	cp.n = n
	return nil
}

func (cp *ControlPanel) startSeed(s settings) {
	fmt.Println(s.SeedStart, s.SeedPort, s.SeedContent)
	if s.SeedStart == "1" {
		srv := NewSeedServer(s.SeedPort, s.SeedContent)
		go srv.Run()
	}
}

func (cp *ControlPanel) Launch() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/", cp.index)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	mux.HandleFunc("/enable", cp.enable)
	mux.HandleFunc("/peers", cp.peers)
	mux.HandleFunc("/report", cp.report)

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

	if s.SeedStart == "1" {
		if _, err := strconv.Atoi(s.SeedPort); err != nil {
			return err
		}
		if len(s.SeedContent) == 0 {
			return fmt.Errorf("no seed server specified")
		}
	}

	return nil
}

func (cp *ControlPanel) exec(templ string, rw http.ResponseWriter, data interface{}) {
	tpl := template.Must(template.ParseGlob("templates/*.html"))
	if err := tpl.ExecuteTemplate(rw, templ, data); err != nil {
		log.Error().Err(err).Str("template", templ).Msg("executing template")
	}
}

func (cp *ControlPanel) index(rw http.ResponseWriter, r *http.Request) {
	cp.exec("index.html", rw, map[string]interface{}{
		"enabled": cp.netEnabled,
		"eps":     cp.eps,
		"feds":    28,
		"audits":  26,
	})
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

	if err := cp.createNetwork(set); err != nil {
		http.Error(rw, err.Error(), http.StatusNotAcceptable)
		return
	}
	cp.startSeed(set)

	cp.enabler.Do(func() {
		cp.app.SetNetwork(cp.n)
		cp.netEnabled = true
		go cp.n.Start()
		go cp.app.Launch()
	})

	http.Redirect(rw, r, "/", http.StatusSeeOther)
}

func (cp *ControlPanel) peers(rw http.ResponseWriter, r *http.Request) {
	var p []string
	if cp.n != nil {
		p = cp.n.Peers()
	}
	cp.exec("peers.html", rw, p)
}

func (cp *ControlPanel) report(rw http.ResponseWriter, r *http.Request) {
	cp.exec("report.html", rw, cp.app.Stats())
}

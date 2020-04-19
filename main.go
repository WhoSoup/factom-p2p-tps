package main

import (
	"flag"
	"os"
	"strconv"

	"github.com/WhoSoup/factom-p2p-max/network"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05", NoColor: true})

	var seedServer, seedPort string

	p2p1 := flag.Bool("p2p1", false, "enable to use the original factom p2p codebase. limited to v9")
	version := flag.Int("v", 10, "the protocol version to use. 9, 10, or 11")
	name := flag.String("name", "", "the name of this specific node")
	port := flag.String("port", "8111", "the port to use for this client (if running multiple nodes on one machine)")
	seed := flag.String("seed", "", "the url of the seed server")
	flag.StringVar(&seedServer, "seedserver", "", "if this is set, a seed server is started containing the addresses listed (comma separated)")
	flag.StringVar(&seedPort, "seedport", "8112", "the port of the seed server")
	flag.Parse()

	if _, err := strconv.Atoi(*port); err != nil {
		log.Fatal().Str("port", *port).Msg("port must be a number")
	}

	if seedServer != "" {
		srv := NewSeedServer(seedPort, seedServer)
		go srv.Run()
	}

	var n network.Network
	if *p2p1 {
		n = network.NewV9()
	} else {
		switch *version {
		case 9, 10, 11:
			n = network.NewV10(*version)
		default:
			log.Fatal().Msg("this version is not available")
		}
	}

	n.Init(*name, *port, *seed)

	cp, err := NewControlPanel(*port)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to start control panel")
	}
	log.Info().Msgf("Control panel started: http://localhost:%s/", *port)
	log.Fatal().Err(cp.Launch()).Msg("control panel shut down")
}

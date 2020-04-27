package main

import (
	"flag"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05", NoColor: true})

	//	var seedServer, seedPort string

	//p2p1 := flag.Bool("p2p1", false, "enable to use the original factom p2p codebase. limited to v9")
	//version := flag.Int("v", 10, "the protocol version to use. 9, 10, or 11")
	//name := flag.String("name", "", "the name of this specific node")
	port := flag.String("port", "7999", "the port for the control panel")
	host := flag.Bool("host", false, "enable to expose the host functionality")
	bcast := flag.Int("broadcast", 16, "number of peers to send broadcasts to")
	//p2pport := flag.String("p2pport", "8111", "the port to use for this client (if running multiple nodes on one machine)")
	//seed := flag.String("seed", "", "the url of the seed server")
	///	flag.StringVar(&seedServer, "seedserver", "", "if this is set, a seed server is started containing the addresses listed (comma separated)")
	//	flag.StringVar(&seedPort, "seedport", "8112", "the port of the seed server")
	flag.Parse()

	cp, err := NewControlPanel(*port, *host, *bcast)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to start control panel")
	}
	log.Info().Msgf("Control panel started: http://localhost:%s/", *port)
	log.Fatal().Err(cp.Launch()).Msg("control panel shut down")
}

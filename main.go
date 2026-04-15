package main

import (
	"flag"
	"log"

	"github.com/ArditZubaku/kvx/config"
	"github.com/ArditZubaku/kvx/server"
)

func main() {
	setupFlags()
	log.Println("Starting the kvx server...")
	log.Fatal(server.RunAsyncTCPServer())
}

func setupFlags() {
	flag.StringVar(&config.Host, "host", "0.0.0.0", "Host for the kvx server")
	flag.IntVar(&config.Port, "port", 7379, "Port for the kvx server")
	flag.IntVar(&config.KeysLimit, "keys-limit", 1_000_000, "The limit of up to how many keys the server can hold")
	flag.Parse()
}

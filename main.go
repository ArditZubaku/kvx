package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/ArditZubaku/kvx/config"
	"github.com/ArditZubaku/kvx/server"
)

func main() {
	setupFlags()
	log.Println("Starting the kvx server...")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	var wg sync.WaitGroup
	wg.Go(func() {
		if err := server.RunAsyncTCPServer(); err != nil {
			log.Printf("Error running the async TCP server -> %+v\n", err)
		}
	})
	wg.Go(func() {
		server.WaitForSignal(sigChan)
	})

	wg.Wait()
}

func setupFlags() {
	flag.StringVar(&config.Host, "host", "0.0.0.0", "Host for the kvx server")
	flag.IntVar(&config.Port, "port", 7379, "Port for the kvx server")
	flag.IntVar(&config.KeysLimit, "keys-limit", 1_000, "The limit of up to how many keys the server can hold")
	flag.StringVar(&config.AofFile, "aof-file", "./kvx-master.aof", "The AppendOnlyFile where to write the dump")
	flag.Float64Var(&config.EvictionRatio, "evic-ratio", 0.4, "The ratio at which to evict keys")
	flag.StringVar(&config.EvictionStrategy, "evic-strategy", "all-keys-random", "The strategy based on which to evict keys")
	flag.Parse()
}

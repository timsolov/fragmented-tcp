package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/timsolov/fragmented-tcp/conf"
	"github.com/timsolov/fragmented-tcp/server"
)

var (
	bindAddr string
)

// init function will run automatically on application startups so we don't need to call it from anywhere.
// also this logic can be implemented by cobra package but it's not necessary for that little project.
func init() {
	flag.StringVar(&bindAddr, "bindAddr", ":2000", "Bind addr for listening connections on.")
	flag.Parse()
}

func main() {
	config := conf.New()

	log := config.LOG()

	srv := server.NewServer(bindAddr, log)
	defer srv.Stop()
	log.Infof("the server is running on %s", bindAddr)

	// wait for interruption
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	<-interrupt
}

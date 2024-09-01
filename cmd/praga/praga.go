package main

import (
	"flag"
	"github.com/cocreators-ee/praga/backend"
)

var config = flag.String("config", "praga.yaml", "Path to praga yaml configuration file")

func main() {
	flag.Parse()
	ok, c := backend.LoadConfig(*config)
	if !ok {
		return
	}

	srv := backend.NewServer(c)
	srv.Start()
}

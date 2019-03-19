package main

import (
	"flag"

	"github.com/TimeForCoin/Server/app"
)


func main() {
	configFile := flag.String("c", "config/default.yaml", "Config file")
	flag.Parse()
	app.Run(*configFile)
}

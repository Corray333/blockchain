package main

import (
	"github.com/Corray333/blockchain/internal/client"
	"github.com/Corray333/blockchain/internal/config"
)

func main() {
	config.LoadConfig()
	client.NewServer(config.CFG.PortServer, config.CFG.PortClient).Run()
}

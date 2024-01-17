package main

import (
	"github.com/Corray333/blockchain/internal/app"
	"github.com/Corray333/blockchain/internal/client"
)

func main() {
	go app.CreateApp().Run()
	go client.NewServer(app.Application.Config.PortServer, app.Application.Config.PortClient).Run()
}

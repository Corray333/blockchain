package app

import (
	"log/slog"
	"net"

	"github.com/Corray333/blockchain/internal/blockchain"
	"github.com/Corray333/blockchain/internal/config"
	"github.com/Corray333/blockchain/internal/wallet"
	"github.com/joho/godotenv"
)

type App struct {
	Blockchain blockchain.Blockchain
	ServerP2P  ServerP2P
	ServerHTTP ServerHTTP
	Config     config.Config
}

func CreateApp() *App {
	godotenv.Load("../.env")
	wallet.InitializeWallet()
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error(err.Error(), "process", "config")
		panic(err)
	}
	// TODO: update server ports
	return &App{
		Blockchain: *blockchain.NewBlockchain(),
		Config:     *cfg,
		ServerP2P: ServerP2P{
			port:        cfg.PortP2P,
			connections: make(map[string]net.Conn),
			masterNode:  nil,
			isMaster:    false,
		},
		ServerHTTP: ServerHTTP{
			port: cfg.PortHTTP,
		},
	}
}

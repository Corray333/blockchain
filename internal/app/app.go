package app

import (
	"log/slog"

	"github.com/Corray333/blockchain/internal/app/node"
	"github.com/Corray333/blockchain/internal/blockchain"
	"github.com/Corray333/blockchain/internal/config"
	"github.com/Corray333/blockchain/internal/person"
	"github.com/Corray333/blockchain/internal/wallet"
	"github.com/joho/godotenv"
)

type App struct {
	Blockchain blockchain.Blockchain
	Wallet     wallet.Wallet
	ServerP2P  ServerP2P
	ServerHTTP ServerHTTP
	Config     config.Config
	UpToDate   bool
	Persons    []person.Person
}

func CreateApp() *App {
	godotenv.Load("../.env")
	wallet.InitializeWallet()
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error(err.Error(), "process", "config")
		panic(err)
	}
	return &App{
		Blockchain: *blockchain.NewBlockchain(),
		Config:     *cfg,
		ServerP2P: ServerP2P{
			port:        cfg.PortP2P,
			connections: make(map[string]node.Node),
			walletsBL:   make(map[string]struct{}),
		},
		ServerHTTP: ServerHTTP{
			port: cfg.PortHTTP,
		},
		UpToDate: false,
		Persons:  make([]person.Person, 25000),
	}
}

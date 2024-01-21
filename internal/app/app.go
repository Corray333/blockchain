package app

import (
	"log/slog"

	"github.com/Corray333/blockchain/internal/app/node"
	"github.com/Corray333/blockchain/internal/blockchain"
	"github.com/Corray333/blockchain/internal/config"
	"github.com/Corray333/blockchain/internal/person"
	"github.com/Corray333/blockchain/internal/wallet"
)

var Application *App

type App struct {
	Blockchain blockchain.Blockchain
	Wallet     wallet.Wallet
	ServerP2P  ServerP2P
	Config     config.Config
	UpToDate   bool
	Persons    []person.Person
}

func CreateApp(recoveryPhrase string) *App {
	// godotenv.Load("../.env")
	err := wallet.InitializeWallet(recoveryPhrase)
	if err != nil {
		slog.Error(err.Error(), "process", "wallet")
		return nil
	}
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error(err.Error(), "process", "config")
		return nil
	}

	Application = &App{
		Blockchain: *blockchain.NewBlockchain(),
		Config:     *cfg,
		ServerP2P: ServerP2P{
			port:        cfg.PortP2P,
			connections: make(map[string]node.Node),
			walletsBL:   make(map[string]struct{}),
		},
		UpToDate: false,
		Persons:  make([]person.Person, 25000),
	}
	return Application
}

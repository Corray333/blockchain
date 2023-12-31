package app

import (
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
	return &App{
		Blockchain: *blockchain.NewBlockchain(),
	}
}

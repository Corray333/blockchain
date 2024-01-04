package app

import (
	"encoding/json"
	"log/slog"
	"net"
	"time"

	"github.com/Corray333/blockchain/internal/blockchain"
	"github.com/Corray333/blockchain/internal/config"
	"github.com/Corray333/blockchain/internal/wallet"
	"github.com/joho/godotenv"
)

const (
	HeartRate = 3 * time.Second
)

type App struct {
	Blockchain blockchain.Blockchain
	Wallet     wallet.Wallet
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
	return &App{
		Blockchain: *blockchain.NewBlockchain(),
		Config:     *cfg,
		ServerP2P: ServerP2P{
			port:        cfg.PortP2P,
			connections: make(map[string]Node),
			masterNode:  nil,
			status:      Follower,
		},
		ServerHTTP: ServerHTTP{
			port: cfg.PortHTTP,
		},
	}
}

func (a *App) Heartbeat() {
	for {
		a.ServerP2P.heartbeat = false
		time.Sleep(HeartRate)
		if a.ServerP2P.status == Master {
			for addr := range a.ServerP2P.connections {
				data := map[string]interface{}{"query": "06"}
				marshalled, err := json.Marshal(data)
				if err != nil {
					slog.Error(err.Error(), "process", "heartbeat")
				}
				conn, err := net.Dial("tcp", addr)
				if err != nil {
					slog.Error(err.Error(), "process", "heartbeat")
				}
				conn.Write(marshalled)
			}
		} else {
			if !a.ServerP2P.heartbeat {
				a.ServerP2P.status = Candidate
				a.ServerP2P.masterNode = nil
				a.StartElection()
			}
		}
	}
}

func (a *App) StartElection() error {
	// TODO: implement election algorithm
	return nil
}

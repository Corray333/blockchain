package app

import (
	"encoding/json"
	"log/slog"
	"net"
	"sync/atomic"
	"time"

	"github.com/Corray333/blockchain/internal/blockchain"
	"github.com/Corray333/blockchain/internal/config"
	"github.com/Corray333/blockchain/internal/wallet"
	"github.com/joho/godotenv"
)

const (
	HeartRate    = 2 * time.Second
	ElectionTime = 5 * time.Second
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
			port:          cfg.PortP2P,
			connections:   make(map[string]Node),
			masterNode:    "",
			currentVote:   "",
			status:        Follower,
			heartbeat:     0,
			walletsBL:     make(map[string]struct{}),
			connectionsBL: make(map[string]struct{}),
		},
		ServerHTTP: ServerHTTP{
			port: cfg.PortHTTP,
		},
	}
}

func (a *App) Heartbeat() {
	for {
		atomic.StoreInt32(&a.ServerP2P.heartbeat, 0)
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
			if a.ServerP2P.heartbeat == 0 {
				a.ServerP2P.status = Candidate
				a.ServerP2P.masterNode = ""
				a.StartElection()
			}
		}
	}
}

func (a *App) StartElection() error {
	for addr := range a.ServerP2P.connections {
		data := map[string]interface{}{"query": "04"}
		marshalled, err := json.Marshal(data)
		if err != nil {
			slog.Error(err.Error(), "process", "election")
			return err
		}
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			slog.Error(err.Error(), "process", "election")
			return err
		}
		if _, err := conn.Write(marshalled); err != nil {
			slog.Error(err.Error(), "process", "election")
			return err
		}
	}
	return nil
}

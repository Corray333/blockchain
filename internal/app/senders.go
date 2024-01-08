package app

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/Corray333/blockchain/internal/blockchain"
)

func SendTransactionToNetwork(a *App, tx blockchain.Transaction) error {
	for addr := range a.ServerP2P.connections {
		err := SendTransaction(a, addr, tx)
		if err != nil {
			slog.Error(err.Error(), "type", "blockchain", "process", "send transaction")
		}
	}
	return nil
}

func SendTransaction(a *App, to string, tx blockchain.Transaction) error {
	query := struct {
		Query     string    `json:"query"`
		PKH       [20]byte  `json:"pkh"`
		Data      []byte    `json:"data"`
		Sign      []byte    `json:"sign"`
		PublicKey []byte    `json:"publicKey"`
		Timestamp time.Time `json:"timestamp"`
	}{
		Query:     "03",
		PKH:       tx.GetOutput().GetPKH(),
		Data:      tx.GetOutput().GetData(),
		Sign:      tx.GetSign(),
		PublicKey: tx.GetPublicKey(),
		Timestamp: tx.GetTimestamp(),
	}
	conn, err := net.Dial("tcp", to)
	if err != nil {
		return fmt.Errorf("error while dialing: %s", err.Error())
	}
	defer conn.Close()
	marshalled, err := json.Marshal(query)
	if err != nil {
		return fmt.Errorf("error while marshalling: %s", err.Error())
	}
	if _, err = conn.Write(marshalled); err != nil {
		return fmt.Errorf("error while writing: %s", err.Error())
	}
	// Handle response
	buf := make([]byte, 128)
	n, err := conn.Read(buf)
	if err != nil {
		return fmt.Errorf("error while reading from querier: %s", err.Error())
	}
	if string(buf[:n]) != "ok" {
		return fmt.Errorf("error while sending transaction")
	}
	return nil
}

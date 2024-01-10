package app

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/Corray333/blockchain/internal/blockchain"
	"github.com/Corray333/blockchain/internal/wallet"
)

func SendAllNodes(a *App, conn net.Conn) error {
	if len(a.ServerP2P.connections) != 0 {
		resp := ""
		for k := range a.ServerP2P.connections {
			resp += k + "|"
		}
		if _, err := conn.Write([]byte(resp[:len(resp)-1])); err != nil {
			return fmt.Errorf("error while writing to querier: %s", err.Error())
		}
	} else {
		if _, err := conn.Write([]byte(GetOutboundIP() + ":" + strconv.Itoa(a.ServerP2P.port) + "|")); err != nil {
			return fmt.Errorf("error while writing to querier: %s", err.Error())
		}
	}
	return nil
}

func AddNewNode(a *App, wallet string, from string, conn net.Conn, lastBlock string) error {
	defer conn.Close()
	if _, ok := a.ServerP2P.walletsBL[wallet]; ok {
		if _, err := conn.Write([]byte("forbidden")); err != nil {
			return fmt.Errorf("error while writing to querier: %s", err.Error())
		}
		return fmt.Errorf("wallet is in black list")
	}
	if lastBlock != fmt.Sprintf("%x", a.Blockchain.GetLastBlock()) {
		if _, err := conn.Write([]byte("not up to date")); err != nil {
			return fmt.Errorf("error while writing to querier: %s", err.Error())
		}
		return fmt.Errorf("node is not up to date")
	}
	a.ServerP2P.connections[from] = Node{
		wallet: wallet,
	}
	if _, err := conn.Write([]byte("ok")); err != nil {
		return fmt.Errorf("error while writing to querier: %s", err.Error())
	}
	slog.Info("new node: " + from)
	return nil
}

func NotifyAllAboutNewNode(a *App, from string) error {
	query := map[string]interface{}{
		"query":   "02", // message about new node
		"address": from,
	}
	marshalled, err := json.Marshal(query)
	if err != nil {
		return fmt.Errorf("error while marshaling query: %s", err.Error())
	}
	for k := range a.ServerP2P.connections {
		NotifyAboutNewNode(k, marshalled)
	}
	return nil
}

func NotifyAboutNewNode(addr string, query []byte) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("error while dialing to querier: %s", err.Error())
	}
	if _, err := conn.Write(query); err != nil {
		return fmt.Errorf("error while writing to querier: %s", err.Error())
	}
	// Handle response
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return fmt.Errorf("error while reading from querier: %s", err.Error())
	}
	if string(buf[:n]) != "ok" {
		return fmt.Errorf("bad response")
	}
	return nil
}

func NewTransaction(a *App, req []byte, conn net.Conn) error {
	query := struct {
		Query     string    `json:"query"`
		PKH       [20]byte  `json:"pkh"`
		Data      []byte    `json:"data"`
		Sign      []byte    `json:"sign"`
		PublicKey []byte    `json:"publicKey"`
		Timestamp time.Time `json:"timestamp"`
	}{}
	// TODO: transaction validation
	err := json.Unmarshal(req, &query)
	if err != nil {
		return fmt.Errorf("error while unmarshaling request: %s", err.Error())
	}
	tx := blockchain.NewTransaction(query.PKH, query.Data, query.PublicKey, query.Timestamp)
	tx.SetSign(query.Sign)
	if err := a.Blockchain.NewTransaction(tx); err != nil {
		pkh := [20]byte{}
		hash := sha256.Sum256(query.PublicKey)
		copy(pkh[:], hash[:])
		a.ServerP2P.walletsBL[wallet.GenerateWalletAddress(pkh)] = struct{}{}
		if _, err := conn.Write([]byte("forbidden")); err != nil {
			return fmt.Errorf("error while writing to querier: %s", err.Error())
		}
		return fmt.Errorf("error while adding transaction to blockchain:" + err.Error())
	}
	if _, err := conn.Write([]byte("ok")); err != nil {
		return fmt.Errorf("error while writing to querier: %s", err.Error())
	}
	err = a.Blockchain.NewTransaction(tx)
	if err != nil {
		return fmt.Errorf("error while adding transaction to blockchain:" + err.Error())
	}
	slog.Info("new transaction", "transaction", tx.String())
	return nil
}

func NotifyAboutNewBlock(a *App, b *blockchain.Block) error {
	query := map[string]interface{}{
		"query":      "08", // query to verify new block
		"block_hash": b.Hash(),
		"timestamp":  b.GetTimestamp(),
	}
	marshalled, err := json.Marshal(query)
	if err != nil {
		return fmt.Errorf("error while marshaling query: %s", err.Error())
	}
	for k := range a.ServerP2P.connections {
		conn, err := net.Dial("tcp", k)
		if err != nil {
			return fmt.Errorf("error while dialing to querier: %s", err.Error())
		}
		if _, err := conn.Write(marshalled); err != nil {
			return fmt.Errorf("error while writing to querier: %s", err.Error())
		}
	}
	return nil
}

func SendAllBlocks(a *App, conn net.Conn) error {
	if !a.UpToDate {
		if _, err := conn.Write([]byte("not up to date")); err != nil {
			if _, err := conn.Write([]byte("error")); err != nil {
				return fmt.Errorf("error while writing to querier: %s", err.Error())
			}
			return fmt.Errorf("error while writing to querier: %s", err.Error())
		}
		return fmt.Errorf("node is not up to date")
	}
	entries, err := os.ReadDir("../../store")
	if err != nil {
		return fmt.Errorf("error while reading directory: %s", err.Error())
	}
	for i := 0; i < len(entries); i++ {
		f, err := os.ReadFile("../../store/" + entries[i].Name())
		if err != nil {
			if _, err := conn.Write([]byte("error")); err != nil {
				return fmt.Errorf("error while writing to querier: %s", err.Error())
			}
			return fmt.Errorf("error while reading file: %s", err.Error())
		}
		query := struct {
			Query string `json:"query"`
			Data  []byte `json:"data"`
		}{
			Query: "06", // new block
			Data:  f,
		}
		marshalled, err := json.Marshal(query)
		if err != nil {
			if _, err := conn.Write([]byte("error")); err != nil {
				return fmt.Errorf("error while writing to querier: %s", err.Error())
			}
			return fmt.Errorf("error while marshaling query: %s", err.Error())
		}
		if _, err := conn.Write(marshalled); err != nil {
			return fmt.Errorf("error while writing to querier: %s", err.Error())
		}
	}
	return nil
}

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

// HandleRequest function handles request from new node to get all the nodes in network.
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

// AddNewNode function handles request from new node to be added in list. isUpToDate is false by default and will be changed by verifying with another request.
func AddNewNode(a *App, wallet string, from string, conn net.Conn, lastBlock string) error {
	defer conn.Close()
	if _, ok := a.ServerP2P.walletsBL[wallet]; ok {
		if _, err := conn.Write([]byte("forbidden")); err != nil {
			return fmt.Errorf("error while writing to querier: %s", err.Error())
		}
		return fmt.Errorf("wallet is in black list")
	}
	a.ServerP2P.connections[from] = Node{
		wallet:     wallet,
		isUpToDate: true,
	}
	if lastBlock != fmt.Sprintf("%x", a.Blockchain.GetLastBlock()) {
		a.ServerP2P.connections[from] = Node{
			wallet:     a.ServerP2P.connections[from].wallet,
			isUpToDate: false,
		}
		if _, err := conn.Write([]byte("not up to date")); err != nil {
			return fmt.Errorf("error while writing to querier: %s", err.Error())
		}
		return fmt.Errorf("node is not up to date")
	}
	if _, err := conn.Write([]byte("ok")); err != nil {
		return fmt.Errorf("error while writing to querier: %s", err.Error())
	}
	slog.Info("new node: " + from)
	return nil
}

// NewTransaction function handles request from node to commit new transaction.
func NewTransaction(a *App, req []byte, conn net.Conn) error {
	query := struct {
		Query     string    `json:"query"`
		PKH       [20]byte  `json:"pkh"`
		Data      []byte    `json:"data"`
		Sign      []byte    `json:"sign"`
		PublicKey []byte    `json:"publicKey"`
		Timestamp time.Time `json:"timestamp"`
		LastBlock [32]byte  `json:"lastBlock"`
	}{}
	// TODO: transaction validation
	err := json.Unmarshal(req, &query)
	if err != nil {
		return fmt.Errorf("error while unmarshaling request: %s", err.Error())
	}
	if a.Blockchain.GetLastBlock() != query.LastBlock {
		if _, err := conn.Write([]byte("not up to date")); err != nil {
			return fmt.Errorf("error while writing to querier: %s", err.Error())
		}
		return fmt.Errorf("node is not up to date")
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

// SendAllBlocks function handles request from new node to get all the blocks in blockchain. Blocks will be sent from the last block to the first one.
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
	for i := len(entries) - 1; i > -1; i-- {
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
			Query: "04", // new block
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

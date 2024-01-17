package handlers

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/Corray333/blockchain/internal/app/node"
	"github.com/Corray333/blockchain/internal/blockchain"
	"github.com/Corray333/blockchain/internal/helpers"
	"github.com/Corray333/blockchain/internal/wallet"
)

// HandleRequest function handles request from new node to get all the nodes in network.
func SendAllNodes(conns map[string]node.Node, port int, conn net.Conn) error {
	if len(conns) != 0 {
		resp := ""
		for k := range conns {
			resp += k + "|"
		}
		if _, err := conn.Write([]byte(resp[:len(resp)-1])); err != nil {
			return fmt.Errorf("error while writing to querier: %s", err.Error())
		}
	} else {
		if _, err := conn.Write([]byte(helpers.GetOutboundIP() + ":" + strconv.Itoa(port) + "|")); err != nil {
			return fmt.Errorf("error while writing to querier: %s", err.Error())
		}
	}
	return nil
}

// AddNewNode function handles request from new node to be added in list. isUpToDate is false by default and will be changed by verifying with another request.
func AddNewNode(lastBlockLocal [32]byte, conns map[string]node.Node, blackList map[string]struct{}, wallet string, from string, conn net.Conn, lastBlock string) error {
	defer conn.Close()
	if _, ok := blackList[wallet]; ok {
		if _, err := conn.Write([]byte("forbidden")); err != nil {
			return fmt.Errorf("error while writing to querier: %s", err.Error())
		}
		return fmt.Errorf("wallet is in black list")
	}
	newNode := node.Node{}
	newNode.SetWallet(wallet)
	conns[from] = newNode
	if lastBlock != fmt.Sprintf("%x", lastBlockLocal) {
		newNode := node.Node{}
		newNode.SetWallet(conns[from].GetWallet())
		newNode.IsNotUpToDate()
		conns[from] = newNode
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
func NewTransaction(bchain *blockchain.Blockchain, blackList map[string]struct{}, req []byte, conn net.Conn) error {
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
	if bchain.GetLastBlock() != query.LastBlock {
		if _, err := conn.Write([]byte("not up to date")); err != nil {
			return fmt.Errorf("error while writing to querier: %s", err.Error())
		}
		return fmt.Errorf("node is not up to date")
	}
	tx := blockchain.NewTransaction(query.PKH, query.Data, query.PublicKey, query.Timestamp)
	tx.Sign = query.Sign
	if err := bchain.NewTransaction(tx); err != nil {
		pkh := [20]byte{}
		hash := sha256.Sum256(query.PublicKey)
		copy(pkh[:], hash[:])
		blackList[wallet.GenerateWalletAddress(pkh)] = struct{}{}
		if _, err := conn.Write([]byte("forbidden")); err != nil {
			return fmt.Errorf("error while writing to querier: %s", err.Error())
		}
		return fmt.Errorf("error while adding transaction to blockchain:" + err.Error())
	}
	if _, err := conn.Write([]byte("ok")); err != nil {
		return fmt.Errorf("error while writing to querier: %s", err.Error())
	}
	err = bchain.NewTransaction(tx)
	if err != nil {
		return fmt.Errorf("error while adding transaction to blockchain:" + err.Error())
	}
	slog.Info("new transaction", "transaction", tx.String())
	return nil
}

// SendAllBlocks function handles request from new node to get all the blocks in blockchain. Blocks will be sent from the last block to the first one.
func SendAllBlocks(isUpToDate bool, conn net.Conn) error {
	if !isUpToDate {
		if _, err := conn.Write([]byte("not up to date")); err != nil {
			if _, err := conn.Write([]byte("error")); err != nil {
				return fmt.Errorf("error while writing to querier: %s", err.Error())
			}
			return fmt.Errorf("error while writing to querier: %s", err.Error())
		}
		return fmt.Errorf("node is not up to date")
	}
	entries, err := os.ReadDir("../../store/blocks")
	if err != nil {
		return fmt.Errorf("error while reading directory: %s", err.Error())
	}
	for i := len(entries) - 1; i > -1; i-- {
		f, err := os.ReadFile("../../store/blocks" + entries[i].Name())
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

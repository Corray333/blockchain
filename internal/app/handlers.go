package app

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net"

	"github.com/Corray333/blockchain/internal/blockchain"
)

func SendAllNodes(a *App, from string) error {
	if from == "" {
		return fmt.Errorf("error while sending all nodes: from is empty")
	}
	if len(a.ServerP2P.connections) != 0 {
		resp := ""
		for k := range a.ServerP2P.connections {
			resp += k + "|"
		}
		conn, err := net.Dial("tcp", from)
		if err != nil {
			return fmt.Errorf("error while dialing to querier: %s", err.Error())
		}
		if _, err := conn.Write([]byte(resp[:len(resp)-1])); err != nil {
			return fmt.Errorf("error while writing to querier: %s", err.Error())
		}
	} else {
		conn, err := net.Dial("tcp", from)
		if err != nil {
			return fmt.Errorf("error while dialing to querier: %s", err.Error())
		}
		if _, err := conn.Write([]byte("")); err != nil {
			return fmt.Errorf("error while writing to querier: %s", err.Error())
		}
	}

	// Handle response
	// buf := make([]byte, 1024)
	// n, err := conn.Read(buf)
	// if err != nil {
	// 	return fmt.Errorf("error while reading from querier: %s", err.Error())
	// }
	// var data map[string]interface{}
	// if err := json.Unmarshal(buf[:n], &data); err != nil {
	// 	return fmt.Errorf("error while unmarshaling data: %s", err.Error())
	// }
	// if data["query"] != "00" {
	// 	return SendAllNodes(a, from)
	// }
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
	var data map[string]interface{}
	if err := json.Unmarshal(buf[:n], &data); err != nil {
		return fmt.Errorf("error while unmarshaling data: %s", err.Error())
	}
	if data["query"] != "00" {
		return NotifyAboutNewNode(addr, query)
	}
	return nil
}

func AddNode(a *App, data map[string]interface{}) error {
	addr, ok := data["address"].(string)
	if !ok {
		return fmt.Errorf("error while casting address to string")
	}
	a.ServerP2P.connections[addr] = Node{
		status: Follower,
		wallet: data["wallet"].(string),
	}
	return nil
}

// func Vote(a *App, from string) error {
// 	conn, err := net.Dial("tcp", from)
// 	if err != nil {
// 		return fmt.Errorf("error while dialing to querier: %s", err.Error())
// 	}
// 	query := map[string]interface{}{
// 		"query": "05",  // query to set a vote
// 		"vote":  false, // vote against
// 	}
// 	if a.ServerP2P.currentVote == "" && a.ServerP2P.status == Follower { // if still didn't vote and is a follower
// 		query["vote"] = true
// 		a.ServerP2P.currentVote = from
// 		if entry, ok := a.ServerP2P.connections[from]; ok {
// 			entry.status = Candidate
// 			a.ServerP2P.connections[from] = entry
// 		}
// 	}
// 	marshlled, err := json.Marshal(query)
// 	if err != nil {
// 		return fmt.Errorf("error while marshaling query: %s", err.Error())
// 	}
// 	if _, err := conn.Write(marshlled); err != nil {
// 		return fmt.Errorf("error while writing to querier: %s", err.Error())
// 	}
// 	return nil
// }

// func GetVotes(a *App) error {
// 	// TODO: optimize
// 	query := map[string]interface{}{
// 		"query": "04", // query to get a vote
// 	}
// 	marshalled, err := json.Marshal(query)
// 	if err != nil {
// 		return fmt.Errorf("error while marshaling query: %s", err.Error())
// 	}
// 	for k := range a.ServerP2P.connections {
// 		conn, err := net.Dial("tcp", k)
// 		if err != nil {
// 			slog.Error(fmt.Sprintf("error while dialing to querier: %s", err.Error()))
// 			continue
// 		}
// 		if _, err := conn.Write(marshalled); err != nil {
// 			slog.Error(fmt.Sprintf("error while writing to querier: %s", err.Error()))
// 			continue
// 		}
// 	}
// 	return nil
// }

// func RecieveVote(a *App, data map[string]interface{}) error {
// 	vote, ok := data["vote"].(bool)
// 	if !ok {
// 		return fmt.Errorf("error while casting vote to bool")
// 	}
// 	if vote {
// 		a.ServerP2P.votesFor++
// 	} else {
// 		a.ServerP2P.votesAgainst++
// 	}
// 	return nil
// }

// func RecieveHeartbeat(a *App) {
// 	mu := sync.Mutex{}
// 	mu.Lock()
// 	atomic.StoreInt32(&a.ServerP2P.heartbeat, 1)
// 	mu.Unlock()
// }

func NewTransaction(a *App, data map[string]interface{}) error {
	// TODO: transaction validation
	var pkh [20]byte
	copy(pkh[:], data["pkh"].([]byte))
	tx := blockchain.NewTransaction(pkh, sha256.Sum256([]byte(data["data"].(string))), []byte(data["data"].(string)), []byte(data["sign"].(string)), []byte(data["public_key"].(string)))
	if err := a.Blockchain.NewTransaction(tx); err != nil {
		return fmt.Errorf("error while adding transaction to blockchain:" + err.Error())
	}
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

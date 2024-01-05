package app

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"sync/atomic"

	"github.com/Corray333/blockchain/internal/blockchain"
)

func SendAllNodes(conns map[string]Node, from string, req []byte) error {
	resp := ""
	for k := range conns {
		resp += k + "|"
	}
	conn, err := net.Dial("tcp", from)
	if err != nil {
		return fmt.Errorf("error while dialing to querier: %s", err.Error())
	}
	if _, err := conn.Write([]byte(resp[:len(resp)-1])); err != nil {
		return fmt.Errorf("error while writing to querier: %s", err.Error())
	}

	return nil
}

func NotifyAboutNewNode(conns map[string]Node, from string) error {
	query := map[string]interface{}{
		"query":   "02", // message about new node
		"address": from,
	}
	for k := range conns {
		marshalled, err := json.Marshal(query)
		if err != nil {
			return fmt.Errorf("error while marshaling query: %s", err.Error())
		}
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

func AddNode(conns map[string]Node, data map[string]interface{}) error {
	addr, ok := data["address"].(string)
	if !ok {
		return fmt.Errorf("error while casting address to string")
	}
	conns[addr] = Node{
		status: Follower,
		wallet: data["wallet"].(string),
	}
	return nil
}

func Vote(a *App, from string) error {
	conn, err := net.Dial("tcp", from)
	if err != nil {
		return fmt.Errorf("error while dialing to querier: %s", err.Error())
	}
	query := map[string]interface{}{
		"query": "05",  // query to set a vote
		"vote":  false, // vote against
	}
	if a.ServerP2P.currentVote == "" && a.ServerP2P.status == Follower { // if still didn't vote and is a follower
		query["vote"] = true
		a.ServerP2P.currentVote = from
		if entry, ok := a.ServerP2P.connections[from]; ok {
			entry.status = Candidate
			a.ServerP2P.connections[from] = entry
		}
	}
	marshlled, err := json.Marshal(query)
	if err != nil {
		return fmt.Errorf("error while marshaling query: %s", err.Error())
	}
	if _, err := conn.Write(marshlled); err != nil {
		return fmt.Errorf("error while writing to querier: %s", err.Error())
	}
	return nil
}

func GetVotes(a *App) error {
	// TODO: optimize
	query := map[string]interface{}{
		"query": "04", // query to get a vote
	}
	marshalled, err := json.Marshal(query)
	if err != nil {
		return fmt.Errorf("error while marshaling query: %s", err.Error())
	}
	for k := range a.ServerP2P.connections {
		conn, err := net.Dial("tcp", k)
		if err != nil {
			slog.Error(fmt.Sprintf("error while dialing to querier: %s", err.Error()))
			continue
		}
		if _, err := conn.Write(marshalled); err != nil {
			slog.Error(fmt.Sprintf("error while writing to querier: %s", err.Error()))
			continue
		}
	}
	return nil
}

func RecieveVote(a *App, data map[string]interface{}) error {
	vote, ok := data["vote"].(bool)
	if !ok {
		return fmt.Errorf("error while casting vote to bool")
	}
	if vote {
		a.ServerP2P.votesFor++
	} else {
		a.ServerP2P.votesAgainst++
	}
	return nil
}

func RecieveHeartbeat(a *App) {
	mu := sync.Mutex{}
	mu.Lock()
	atomic.StoreInt32(&a.ServerP2P.heartbeat, 1)
	mu.Unlock()
}

func NewTransaction(a *App, data map[string]interface{}) error {
	// TODO: transaction validation
	var pkh [20]byte
	copy(pkh[:], data["pkh"].([]byte))
	tx := blockchain.NewTransaction(pkh, sha256.Sum256([]byte(data["data"].(string))), []byte(data["data"].(string)), []byte(data["sign"].(string)), []byte(data["public_key"].(string)))
	mu := sync.Mutex{}
	mu.Lock()
	if err := a.Blockchain.NewTransaction(tx); err != nil {
		mu.Unlock()
		return fmt.Errorf("error while adding transaction to blockchain:" + err.Error())
	}
	mu.Unlock()
	return nil
}

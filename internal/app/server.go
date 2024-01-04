package app

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"strconv"
	"time"

	"github.com/Corray333/blockchain/internal/wallet"
)

const (
	Follower = iota
	Candidate
	Master
)

type Node struct {
	status uint8
	wallet string
}

func (n *Node) GetWallet() string {
	return n.wallet
}

func (n *Node) SetWallet(address string) {
	n.wallet = address
}

func (n *Node) SetStatus(status uint8) {
	n.status = status
}

func (n *Node) GetStatus() uint8 {
	return n.status
}

type ServerP2P struct {
	port        int
	connections map[string]Node
	masterNode  net.Conn
	status      uint8
	heartbeat   bool
}

type ServerHTTP struct {
	port int
}

func (a *App) Run() {
	listener, err := net.Listen("tcp", "0.0.0.0:"+strconv.Itoa(a.ServerP2P.port))
	if err != nil {
		slog.Error("falat error while starting server:" + err.Error())
		panic(err)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			slog.Error("error while accepting connection:" + err.Error())
			continue
		}
		conn.SetDeadline(time.UnixMicro(0))
		go a.handleConnection(conn)
	}
}

func (a *App) handleConnection(conn net.Conn) {
	slog.Info("new connection from " + conn.RemoteAddr().String())
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		slog.Error("error while reading from connection:" + err.Error())
		return
	}
	err = a.HandleRequest(conn, buffer[:n])
	if err != nil {
		slog.Error("error while handling request:" + err.Error())
		return
	}
}

func (a *App) ConnectWithBootnodes() error {
	for _, v := range a.Config.BootNodes {
		err := a.ConnectDirectly(v)
		if err != nil {
			slog.Error(err.Error())
			continue
		}
		return nil
	}
	return errors.New("error while connecting to network: all the boot nodes are unavalible, try to connect directly")
}

func (a *App) ConnectDirectly(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return errors.New("error while connecting to network:" + err.Error())
	}
	defer conn.Close()
	// Code 01 is for network-service commands
	query := map[string]interface{}{
		"query":  "01",
		"wallet": wallet.GetAddress(),
	}
	bytesQuery, err := json.Marshal(query)
	if err != nil {
		return errors.New("error while connecting to network:" + err.Error())
	}
	if _, err := conn.Write(bytesQuery); err != nil {
		return errors.New("error while connecting to network:" + err.Error())
	}
	return nil
	// buf := make([]byte, 4096)
	// k := 0
	// for {
	// 	n, err := conn.Read(buf[k:])
	// 	if err != nil {
	// 		return errors.New("error while connecting to network:" + err.Error())
	// 	}
	// 	splited := strings.Split(string(buf[:n]), "|")
	// 	for _, v := range splited {
	// 		a.ServerP2P.connections[v] = nil
	// 	}
	// 	if len(splited[len(splited)-1]) < 20 {
	// 		delete(a.ServerP2P.connections, splited[len(splited)-1])
	// 		k = len(splited[len(splited)-1])
	// 	}
	// }
}

package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net"
	"strconv"
	"strings"

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

type ServerP2P struct {
	port        int             // port of the node for P2P connection
	connections map[string]Node // map of nodes
	// masterNode    string              // master node IP address
	// currentVote   string              // current vote
	// votesFor      int                 // number of votes for
	// votesAgainst  int                 // number of votes against
	// status        uint8               // 0 - follower, 1 - candidate, 2 - master
	// heartbeat     int32               // 0 - no heartbeat, 1 - heartbeat, int32 for atomic
	walletsBL     map[string]struct{} // black list of wallets
	connectionsBL map[string]struct{} // black list of IP addresses
}

type ServerHTTP struct {
	port int
}

func GetOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}

func (a *App) Run() {
	listener, err := net.Listen("tcp", GetOutboundIP()+":"+strconv.Itoa(a.ServerP2P.port))
	fmt.Printf("Starting server %s", listener.Addr().String())
	if err != nil {
		slog.Error("error while starting server:" + err.Error())
		panic(err)
	}
	go a.ConnectWithBootnodes()
	for {
		conn, err := listener.Accept()
		if err != nil {
			slog.Error("error while accepting connection:" + err.Error())
			continue
		}
		go a.handleConnection(conn)
	}
}

func (a *App) handleConnection(conn net.Conn) {
	defer conn.Close()
	slog.Info("new connection from " + conn.RemoteAddr().String())
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		slog.Error("error while reading from connection:" + err.Error())
		return
	}
	slog.Info("received request from " + conn.RemoteAddr().String() + ": " + string(buffer[:n]))
	err = a.HandleRequest(conn, buffer[:n])
	if err != nil {
		slog.Error("error while handling request:" + err.Error())
		return
	}
}

func (a *App) ConnectWithBootnodes() error {
	if len(a.Config.BootNodes) == 0 {
		return errors.New("error while connecting to network: no boot nodes")
	}
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
		"from":   GetOutboundIP() + ":" + strconv.Itoa(a.ServerP2P.port),
	}
	bytesQuery, err := json.Marshal(query)
	if err != nil {
		return errors.New("error while connecting to network:" + err.Error())
	}
	if _, err := conn.Write(bytesQuery); err != nil {
		return errors.New("error while connecting to network:" + err.Error())
	}
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return errors.New("error while connecting to network:" + err.Error())
	}
	splited := strings.Split(string(buf[:n]), "|")
	for _, v := range splited {
		if v == "" {
			continue
		}
		a.ServerP2P.connections[v] = Node{}
	}
	slog.Info("connected to network: " + string(buf[:n]) + "...")
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

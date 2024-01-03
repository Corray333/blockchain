package app

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"strconv"
	"strings"
	"time"
)

type ServerP2P struct {
	port        int
	connections map[string]net.Conn
	masterNode  net.Conn
	isMaster    bool
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
	a.ServerP2P.connections[conn.RemoteAddr().String()] = conn
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
	conn.SetDeadline(time.Now().Add(time.Second * 5))
	// TODO: make list of query types with codes
	// Code 01 is for network-service commands
	query := map[string]interface{}{
		"type":  "01",
		"query": "01",
	}
	bytesQuery, err := json.Marshal(query)
	if err != nil {
		return errors.New("error while connecting to network:" + err.Error())
	}
	conn.Write(bytesQuery)
	buf := make([]byte, 4096)
	k := 0
	for {
		n, err := conn.Read(buf[k:])
		if err != nil {
			return errors.New("error while connecting to network:" + err.Error())
		}
		splited := strings.Split(string(buf[:n]), "|")
		for _, v := range splited {
			a.ServerP2P.connections[v] = nil
		}
		if len(splited[len(splited)-1]) < 20 {
			delete(a.ServerP2P.connections, splited[len(splited)-1])
			k = len(splited[len(splited)-1])
		}
	}
}

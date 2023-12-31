package app

import (
	"log/slog"
	"net"
	"strconv"
	"strings"
	"time"
)

type ServerP2P struct {
	port        int
	connections map[string]net.Conn
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
	err = a.Blockchain.HandleRequest(string(buffer[:n]))
	if err != nil {
		slog.Error("error while handling request:" + err.Error())
		return
	}
}

func (a *App) ConnectToNetwork() error {
bootNodesLoop:
	for _, v := range a.Config.BootNodes {
		conn, err := net.Dial("tcp", v)
		if err != nil {
			slog.Error("error while connecting to network:" + err.Error())
			continue
		}
		// TODO: make list of query types with codes
		// Code 01 is for network-service commands
		conn.Write([]byte(`"01", "connect to network"`))
		buf := make([]byte, 4096)
		k := 0
		for {
			n, err := conn.Read(buf[k:])
			if err != nil {
				slog.Error("error while reading from boot node " + v + ": " + err.Error())
				continue bootNodesLoop
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
	return nil
}

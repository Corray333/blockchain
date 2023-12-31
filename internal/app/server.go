package app

import (
	"log/slog"
	"net"
	"strconv"
	"time"
)

type ServerP2P struct {
	port        int
	connections map[string]net.Conn
}

type ServerHTTP struct {
	port int
}

func (s *ServerP2P) Run() {
	listener, err := net.Listen("tcp", "0.0.0.0:"+strconv.Itoa(s.port))
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
		go s.handleConnection(conn)
	}
}

func (s *ServerP2P) handleConnection(conn net.Conn) {
	s.connections[conn.RemoteAddr().String()] = conn
	slog.Info("new connection from " + conn.RemoteAddr().String())
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		slog.Error("error while reading from connection:" + err.Error())
		return
	}

}

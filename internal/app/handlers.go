package app

import (
	"encoding/json"
	"fmt"
	"net"
)

func SendAllNodes(conns map[string]Node, conn net.Conn, req []byte) error {
	resp := ""
	for k := range conns {
		resp += k + "|"
	}
	if _, err := conn.Write([]byte(resp[:len(resp)-1])); err != nil {
		return fmt.Errorf("error while writing to querier: %s", err.Error())
	}

	return nil
}

func NotifyAboutNewNode(conns map[string]Node, conn net.Conn) error {
	query := map[string]interface{}{
		"query": "02", // message about new node
		"data": map[string]interface{}{
			"addr": conn.RemoteAddr().String(),
		},
	}
	for k := range conns {
		bytesQuery, err := json.Marshal(query)
		if err != nil {
			return fmt.Errorf("error while marshaling query: %s", err.Error())
		}
		conn, err := net.Dial("tcp", k)
		if err != nil {
			return fmt.Errorf("error while dialing to querier: %s", err.Error())
		}
		if _, err := conn.Write(bytesQuery); err != nil {
			return fmt.Errorf("error while writing to querier: %s", err.Error())
		}
	}
	return nil
}

func AddNode(conns map[string]Node, conn net.Conn, data map[string]interface{}) error {
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

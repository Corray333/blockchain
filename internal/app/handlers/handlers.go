package handlers

import (
	"encoding/json"
	"fmt"
	"net"
)

func SendAllNodes(conns map[string]net.Conn, conn net.Conn, req []byte) error {
	resp := ""
	for k := range conns {
		resp += k + "|"
	}
	if _, err := conn.Write([]byte(resp[:len(resp)-1])); err != nil {
		return fmt.Errorf("error while writing to querier: %s", err.Error())
	}

	return nil
}

func NotifyAboutNewNode(conns map[string]net.Conn, conn net.Conn) error {
	for k := range conns {
		query := map[string]interface{}{
			"type":  "01",
			"query": "02", // message about new node
			"data": map[string]interface{}{
				"addr": conn.RemoteAddr().String(),
			},
		}
		bytesQuery, err := json.Marshal(query)
		if err != nil {
			return fmt.Errorf("error while marshaling query: %s", err.Error())
		}
		conns[k].Write(bytesQuery)
	}
	return nil
}

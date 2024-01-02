package app

import (
	"encoding/json"
	"fmt"
	"net"
)

func (a *App) HandleRequest(from string, req []byte) error {
	var data map[string]interface{}
	err := json.Unmarshal(req, &data)
	if err != nil {
		return fmt.Errorf("error while unmarshaling request: %s", err.Error())
	}
	switch data["type"] {
	case "01": // network-service command
		switch data["query"] {
		case "01": // query to get all nodes in network
			resp := ""
			for k := range a.ServerP2P.connections {
				resp += k + "|"
			}
			conn, err := net.Dial("tcp", from)
			if err != nil {
				return fmt.Errorf("error while connecting to querier: %s", err.Error())
			}
			defer conn.Close()
			if _, err := conn.Write([]byte(resp[:len(resp)-1])); err != nil {
				return fmt.Errorf("error while writing to querier: %s", err.Error())
			}
			for k := range a.ServerP2P.connections {
				query := map[string]interface{}{
					"type":  "01",
					"query": "02", // message about new node
				}
				bytesQuery, err := json.Marshal(query)
				if err != nil {
					return fmt.Errorf("error while marshaling query: %s", err.Error())
				}
				a.ServerP2P.connections[k].Write(bytesQuery)
			}
			return nil
		case "02": // message about new node

		}
	case "02": // consensus command
	case "03": // blockchain command
	}
	return nil
}

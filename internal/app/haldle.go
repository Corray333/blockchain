package app

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/Corray333/blockchain/internal/app/handlers"
)

func (a *App) HandleRequest(conn net.Conn, req []byte) error {
	var data map[string]interface{}
	err := json.Unmarshal(req, &data)
	if err != nil {
		return fmt.Errorf("error while unmarshaling request: %s", err.Error())
	}
	// Redirect if this is not a master node
	if a.ServerP2P.isMaster {
		data["addr"] = conn.RemoteAddr().String()
		req, err = json.Marshal(data)
		if err != nil {
			return fmt.Errorf("error while marshaling request: %s", err.Error())
		}
		if _, err := a.ServerP2P.masterNode.Write(req); err != nil {
			return fmt.Errorf("error while writing to master node: %s", err.Error())
		}
		return nil
	}
	switch data["type"] {
	case "01": // network-service command
		switch data["query"] {
		case "01": // query to get all nodes in network
			handlers.SendAllNodes(a.ServerP2P.connections, conn, req)
		case "02": // message about new node
			handlers.NotifyAboutNewNode(a.ServerP2P.connections, conn)
		}
	case "02": // consensus command
	case "03": // blockchain command
	}
	return nil
}

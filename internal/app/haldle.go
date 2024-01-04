package app

import (
	"encoding/json"
	"fmt"
	"net"
)

func (a *App) HandleRequest(conn net.Conn, req []byte) error {
	var data map[string]interface{}
	err := json.Unmarshal(req, &data)
	if err != nil {
		return fmt.Errorf("error while unmarshaling request: %s", err.Error())
	}
	switch data["query"] {
	case "01": // query to get all nodes in network
		// Redirect if this is not a master node
		if a.ServerP2P.status != Master {
			err := Redirect(data, conn, a.ServerP2P.masterNode)
			return err
		}
		err := SendAllNodes(a.ServerP2P.connections, conn, req)
		if err != nil {
			return fmt.Errorf("error while sending all nodes: %s", err.Error())
		}
		err = NotifyAboutNewNode(a.ServerP2P.connections, conn)
		if err != nil {
			return fmt.Errorf("error while notifying about new node: %s", err.Error())
		}
	case "02": // message about new node
		// Redirect if this is not a master node
		if a.ServerP2P.status != Master {
			err := Redirect(data, conn, a.ServerP2P.masterNode)
			return err
		}
		if err := NotifyAboutNewNode(a.ServerP2P.connections, conn); err != nil {
			return fmt.Errorf("error while notifying about new node: %s", err.Error())
		}
		return nil
	case "03": // query to add a new node
		return AddNode(a.ServerP2P.connections, conn, data)
	case "04": // query to vote
	case "05": // query to get a vote
	case "06": // heartbeat
	case "07": // commit
	case "08": // query to verify new block
	}
	return nil
}

func Redirect(data map[string]interface{}, conn net.Conn, master net.Conn) error {
	data["address"] = conn.RemoteAddr().String()
	req, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("error while marshaling request: %s", err.Error())
	}
	if _, err := master.Write(req); err != nil {
		return fmt.Errorf("error while writing to master node: %s", err.Error())
	}
	return nil
}

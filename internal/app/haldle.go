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
	case "00": // verify that query is done
	case "01": // query to get all nodes in network
		if err := SendAllNodes(a, data["from"].(string)); err != nil {
			return fmt.Errorf("error while sending all nodes: %s", err.Error())
		}
	case "02": // message about new node
		if err := NotifyAllAboutNewNode(a, conn.RemoteAddr().String()); err != nil {
			return fmt.Errorf("error while notifying about new node: %s", err.Error())
		}
	case "03": // query to add a new node
		return AddNode(a, data)
	case "04": // commit transaction
		return NewTransaction(a, data)
	case "05": // query to verify new block
	}
	return nil
}

// func Redirect(data map[string]interface{}, addr string, master string) error {
// 	data["address"] = addr
// 	req, err := json.Marshal(data)
// 	if err != nil {
// 		return fmt.Errorf("error while marshaling request: %s", err.Error())
// 	}
// 	conn, err := net.Dial("tcp", master)
// 	if err != nil {
// 		return fmt.Errorf("error while dialing to master node: %s", err.Error())
// 	}
// 	if _, err := conn.Write(req); err != nil {
// 		return fmt.Errorf("error while writing to master node: %s", err.Error())
// 	}
// 	return nil
// }

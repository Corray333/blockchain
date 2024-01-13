package app

import (
	"encoding/json"
	"fmt"
	"net"
)

// HandleRequest function handles request from other node.
func (a *App) HandleRequest(conn net.Conn, req []byte) error {
	var data map[string]interface{}
	err := json.Unmarshal(req, &data)
	if err != nil {
		return fmt.Errorf("error while unmarshaling request: %s", err.Error())
	}
	switch data["query"] {
	case "01": // query to get all nodes in network
		if err := SendAllNodes(a, conn); err != nil {
			return fmt.Errorf("error while sending all nodes: %s", err.Error())
		}
	case "02": // message about new node
		if err := AddNewNode(a, data["wallet"].(string), data["from"].(string), conn, data["lastBlock"].(string)); err != nil {
			return fmt.Errorf("error while notifying about new node: %s", err.Error())
		}
	case "03": // commit transaction
		if err := NewTransaction(a, req, conn); err != nil {
			return fmt.Errorf("error while committing transaction: %s", err.Error())
		}
	case "04":
	}
	return nil
}

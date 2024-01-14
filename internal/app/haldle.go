package app

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/Corray333/blockchain/internal/app/handlers"
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
		if err := handlers.SendAllNodes(a.ServerP2P.connections, a.ServerP2P.port, conn); err != nil {
			return fmt.Errorf("error while sending all nodes: %s", err.Error())
		}
	case "02": // message about new node
		if err := handlers.AddNewNode(a.Blockchain.GetLastBlock(), a.ServerP2P.connections, a.ServerP2P.walletsBL, data["wallet"].(string), data["from"].(string), conn, data["lastBlock"].(string)); err != nil {
			return fmt.Errorf("error while notifying about new node: %s", err.Error())
		}
	case "03": // commit transaction
		if err := handlers.NewTransaction(&a.Blockchain, a.ServerP2P.walletsBL, req, conn); err != nil {
			return fmt.Errorf("error while committing transaction: %s", err.Error())
		}
	case "04": // query to get all blocks
		if err := handlers.SendAllBlocks(a.UpToDate, conn); err != nil {
			return fmt.Errorf("error while sending all blocks: %s", err.Error())
		}
	}
	return nil
}

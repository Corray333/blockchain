package node

import "github.com/Corray333/blockchain/internal/person"

type Node struct {
	wallet     string
	isUpToDate bool
	person     person.Person
}

func (n Node) GetWallet() string {
	return n.wallet
}
func (n Node) GetIsUpToDate() bool {
	return n.isUpToDate
}

func (n *Node) SetWallet(wallet string) {
	n.wallet = wallet
}
func (n *Node) IsUpToDate() {
	n.isUpToDate = true
}
func (n *Node) IsNotUpToDate() {
	n.isUpToDate = false
}

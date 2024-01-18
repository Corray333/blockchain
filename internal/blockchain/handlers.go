package blockchain

import "fmt"

func (b *Blockchain) GetLastTransaction() *Transaction {
	if len(b.TransactionPool) == 0 {
		return nil
	}
	return &b.TransactionPool[len(b.TransactionPool)-1]
}

func (b *Blockchain) GetTransactionByHash(hash string) *Transaction {
	for _, tx := range b.TransactionPool {
		if fmt.Sprintf("%x", tx.Hash()) == hash {
			return &tx
		}
	}
	return nil
}

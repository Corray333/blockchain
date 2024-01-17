package blockchain

func (b *Blockchain) GetLastTransaction() Transaction {
	return b.transactionPool[len(b.transactionPool)-1]
}

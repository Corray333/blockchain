package blockchain

import (
	"crypto/sha256"
	"fmt"
	"time"
)

// Block structure represents a block in a blockchain
type Block struct {
	Prev         [32]byte
	Root         [32]byte
	Transactions []Transaction
	Timestamp    time.Time
}

func GetMerkleRoot(hashes [][32]byte) [32]byte {
	if len(hashes) == 1 {
		return hashes[0]
	}
	if len(hashes)%2 == 1 {
		hashes = append(hashes, hashes[len(hashes)-1])
	}
	for i := 0; i < len(hashes)/2; i++ {
		hashes[i] = sha256.Sum256(append(hashes[2*i][:], hashes[2*i+1][:]...))
	}
	return GetMerkleRoot(hashes[:len(hashes)/2])
}

// NewBlock creates a new block
func NewBlock(prev [32]byte, transactionPool *[]Transaction) *Block {
	length := len(*transactionPool)
	b := &Block{
		Prev:         prev,
		Root:         [32]byte{},
		Transactions: make([]Transaction, length),
		Timestamp:    time.Now(),
	}
	copy(b.Transactions, *transactionPool)
	*transactionPool = []Transaction{}
	hashes := make([][32]byte, length)
	for i := range hashes {
		hashes[i] = b.Transactions[i].Hash()
	}
	b.Root = GetMerkleRoot(hashes)
	return b
}

// Block.String turns block into a string, placing
func (b Block) String() string {
	return fmt.Sprintf("prev hash: %x,\nmerkle root: %x,\ntimestamp: %s\n", b.Prev[:], b.Root[:], b.Timestamp.String())
}

// Block.Hash gets the hash of the block
func (b Block) Hash() [32]byte {
	return sha256.Sum256([]byte(b.String()))
}

// Input structure represents a transaction input
type Input struct {
	Hash [32]byte
	ID   uint8
}

// Output structure represents a transaction output
type Output struct {
	PKH    [20]byte
	Amount uint32
}

// Inputs is an alias for a slice of Input: []Input
type Inputs []Input

// Inputs.String turns Inputs to a string
func (inputs Inputs) String() string {
	res := ""
	for _, str := range inputs {
		res += fmt.Sprintf("%x", str.Hash[:])
	}
	return res
}

// Outputs is an alias for a slice of Output: []Output
type Outputs []Output

// Outputs.String turns Outputs to a string
func (outputs Outputs) String() string {
	res := ""
	for _, str := range outputs {
		res += fmt.Sprintf("%x%d", str.PKH[:], str.Amount)
	}
	return res
}

// Transact structure represents a transaction in blockchain
type Transaction struct {
	Inputs    Inputs
	Outputs   Outputs
	Sign      [32]byte
	PublicKey [32]byte
}

// Transaction.String turns Transaction into a string
func (tx Transaction) String() string {
	return fmt.Sprintf("%s%s%x%x", tx.Inputs.String(), tx.Outputs.String(), tx.Sign[:], tx.PublicKey[:])
}

// Transaction.Hash returns the hash of the transaction
func (tx Transaction) Hash() [32]byte {
	return sha256.Sum256([]byte(tx.String()))
}

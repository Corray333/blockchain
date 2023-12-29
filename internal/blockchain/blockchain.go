package blockchain

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/crypto/secp256k1"
)

type Blockchain struct {
	blockList       [][32]byte
	transactionPool []Transaction
}

func NewBlockchain() *Blockchain {
	return &Blockchain{
		blockList:       [][32]byte{},
		transactionPool: []Transaction{},
	}
}

func (b *Blockchain) PrintTransactions() {
	for _, t := range b.transactionPool {
		fmt.Println(t.String())
	}
}

func (b *Blockchain) NewTransaction(tx Transaction) error {
	ok := secp256k1.VerifySignature(tx.publicKey, tx.output.token.hash[:], tx.sign)
	if !ok {
		slog.Info("error while verifying transaction from " + fmt.Sprintf("%x", tx.input))
		return errors.New("error while verifying transaction from " + fmt.Sprintf("%x", tx.input))
	}
	b.transactionPool = append(b.transactionPool, tx)
	return nil
}

// CreateBlock creates a new block
func (b *Blockchain) CreateBlock() *Block {
	block := &Block{
		prev:         b.blockList[len(b.blockList)],
		root:         [32]byte{},
		transactions: make([]Transaction, len(b.transactionPool)),
		timestamp:    time.Now(),
	}
	copy(block.transactions, b.transactionPool)
	b.transactionPool = []Transaction{}
	hashes := make([][32]byte, len(b.transactionPool))
	for i := range hashes {
		hashes[i] = block.transactions[i].Hash()
	}
	block.root = GetMerkleRoot(hashes)
	return block
}

// Block structure represents a block in a blockchain
type Block struct {
	prev          [32]byte
	root          [32]byte
	transactions  []Transaction
	timestamp     time.Time
	creatorAdress string
}

// Block.Save function saves a block in store folder
func (b Block) Save() error {
	hash := b.Hash()
	file, err := os.Create(fmt.Sprintf("../../store/%x.blk", hash))
	if err != nil {
		// TODO: choose to log error emediately or just to return it
		slog.Error(fmt.Sprintf(`error while saving new block with hash "%x": %e`, hash, err))
		return err
	}
	n, err := file.Write([]byte(b.String()))
	if err != nil {
		// TODO: choose to log error emediately or just to return it
		slog.Error(fmt.Sprintf(`error while saving new block with hash "%x": %e`, hash, err))
		return err
	}
	slog.Error(fmt.Sprintf(`%d bytes are written in file "%s" with block "%x info`, n, file.Name(), hash))
	return nil
}

// GetMerkleRoot returns the merkle root got by transactions hashes
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

// Block.String turns block into a string, placing
func (b Block) String() string {
	return fmt.Sprintf("prev hash: %x\nmerkle root: %x\ntimestamp: %s\n", b.prev[:], b.root[:], b.timestamp.String())
}

// Block.Hash gets the hash of the block
func (b Block) Hash() [32]byte {
	return sha256.Sum256([]byte(b.String()))
}

type Token struct {
	hash [32]byte
	data []byte
}

// Output structure represents a transaction output
type Output struct {
	pkh   [20]byte
	token Token
}

func (o Output) String() string {
	return fmt.Sprintf("pkh: %x, token: %x, description: %s", o.pkh, o.token.hash, string(o.token.data))
}

// Transact structure represents a transaction in blockchain
type Transaction struct {
	input     [32]byte
	output    Output
	sign      []byte
	publicKey []byte
}

func NewTransaction(input [32]byte, pkh [20]byte, hash [32]byte, data []byte, sign []byte, publicKey []byte) Transaction {
	return Transaction{
		input: [32]byte{},
		output: Output{
			pkh: pkh,
			token: Token{
				hash: hash,
				data: data,
			},
		},
		sign:      sign,
		publicKey: publicKey,
	}
}

// Transaction.String turns Transaction into a string
func (tx Transaction) String() string {
	return fmt.Sprintf("input: %s\noutput: %s\nsign: %x\npublic key: %x", tx.input, tx.output.String(), tx.sign, tx.publicKey[:])
}

// Transaction.Hash returns the hash of the transaction
func (tx Transaction) Hash() [32]byte {
	return sha256.Sum256([]byte(tx.String()))
}

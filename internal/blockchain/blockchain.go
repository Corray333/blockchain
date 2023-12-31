package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Corray333/blockchain/internal/wallet"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
)

// Blockchain structure represents a blockchain
type Blockchain struct {
	blockList       [][32]byte
	transactionPool []Transaction
}

// NewBlockchain returnc a new blockchain
func NewBlockchain() *Blockchain {
	return &Blockchain{
		blockList: [][32]byte{
			sha256.Sum256(nil),
		},
		transactionPool: []Transaction{},
	}
}

// Blockchain.GetTransactionsString returns a string with all transactions
func (b *Block) GetTransactionsString() string {
	res := ""
	for _, t := range b.transactions {
		res += fmt.Sprintf("%x|%x|%x|%s|%x|%x|", t.input, t.output.pkh, t.output.token.hash, t.output.token.data, t.publicKey, t.sign)
	}
	return res
}

// Blockchain.PrintTransactions prints all transactions in transaction pool
func (b *Blockchain) PrintTransactions() {
	fmt.Println("====================\tTransactions\t====================")
	for i, t := range b.transactionPool {
		fmt.Printf("====================\tTransaction %d\t====================\n", i)
		fmt.Println(t.String())
	}
	fmt.Println("====================\tTransactions\t====================")
}

// Blockchain.NewTransaction creates a new transaction in transaction pool
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
		prev:         b.blockList[len(b.blockList)-1],
		root:         [32]byte{},
		transactions: make([]Transaction, len(b.transactionPool)),
		timestamp:    time.Now(),
	}
	copy(block.transactions, b.transactionPool)
	b.transactionPool = []Transaction{}
	hashes := make([][32]byte, len(block.transactions))
	for i := range hashes {
		hashes[i] = block.transactions[i].Hash()
	}
	block.root = GetMerkleRoot(hashes)
	block.creatorAdress = wallet.GetAddress()
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
	file, err := os.Create(fmt.Sprintf("../store/%x.blk", hash))
	if err != nil {
		return err
	}
	res := fmt.Sprintf("%x|%x|%d|%s\n", b.prev, b.root, b.timestamp.UnixMicro(), b.creatorAdress)
	// TODO: save transactions with compressed public keys
	res += b.GetTransactionsString()
	n, err := file.Write([]byte(res))
	if err != nil {
		// TODO: choose to log error emediately or just to return it
		return err
	}
	// TODO: save creator adress
	slog.Info(fmt.Sprintf(`%d bytes are written in file "%s" with block "%x info`, n, file.Name(), hash))
	return nil
}

func LoadBlock(hash [32]byte) (*Block, error) {
	file, err := os.Open(fmt.Sprintf("../store/%x.blk", hash))
	if err != nil {
		return nil, fmt.Errorf(`error while loading block with hash "%x": %s`, hash, err.Error())
	}
	var block Block
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf(`error while loading block with hash "%x": %s`, hash, err.Error())
	}
	splitted := strings.Split(string(data), "\n")
	blockData := strings.Split(splitted[0], "|")
	transactions := strings.Split(splitted[1], "|")
	prev, err := hex.DecodeString(blockData[0])
	if err != nil {
		return nil, fmt.Errorf(`error while loading block with hash "%x": %s`, hash, err.Error())
	}
	copy(block.prev[:], prev)
	root, err := hex.DecodeString(blockData[1])
	if err != nil {
		return nil, fmt.Errorf(`error while loading block with hash "%x": %s`, hash, err.Error())
	}
	copy(block.root[:], root)
	microSeconds, err := strconv.ParseInt(blockData[2], 10, 64)
	if err != nil {
		return nil, fmt.Errorf(`error while loading block with hash "%x": %s`, hash, err.Error())
	}
	block.timestamp = time.UnixMicro(microSeconds)
	if err != nil {
		return nil, fmt.Errorf(`error while loading block with hash "%x": %s`, hash, err.Error())
	}
	block.creatorAdress = blockData[3]
	block.transactions = []Transaction{}
	// TODO: load transactions
	for i := 0; i < len(transactions)-1; i += 6 {
		var transaction Transaction
		input, err := hex.DecodeString(transactions[i])
		if err != nil {
			return nil, fmt.Errorf(`error while loading block with hash "%x": %s`, hash, err.Error())
		}
		pkh, err := hex.DecodeString(transactions[i+1])
		if err != nil {
			return nil, fmt.Errorf(`error while loading block with hash "%x": %s`, hash, err.Error())
		}
		hash, err := hex.DecodeString(transactions[i+2])
		if err != nil {
			return nil, fmt.Errorf(`error while loading block with hash "%x": %s`, hash, err.Error())
		}
		pub, err := hex.DecodeString(transactions[i+4])
		if err != nil {
			return nil, fmt.Errorf(`error while loading block with hash "%x": %s`, hash, err.Error())
		}
		sign, err := hex.DecodeString(transactions[i+5])
		if err != nil {
			return nil, fmt.Errorf(`error while loading block with hash "%x": %s`, hash, err.Error())
		}
		copy(transaction.input[:], input)
		copy(transaction.output.pkh[:], pkh)
		copy(transaction.output.token.hash[:], hash)
		transaction.output.token.data = []byte(transactions[i+3])
		transaction.publicKey = make([]byte, 65)
		transaction.sign = make([]byte, 64)
		copy(transaction.publicKey[:], pub)
		copy(transaction.sign[:], sign)
		block.transactions = append(block.transactions, transaction)
	}
	// TODO: check data format
	return &block, nil
}

// GetMerkleRoot returns the merkle root got by transactions hashes
func GetMerkleRoot(hashes [][32]byte) [32]byte {
	if len(hashes) == 0 {
		return [32]byte{}
	}
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
	return fmt.Sprintf("prev hash: %x\nmerkle root: %x\ntimestamp: %s\ncreator: %s\n", b.prev[:], b.root[:], b.timestamp.String(), b.creatorAdress)
}

// Block.PrintTransactions prints all transactions in transaction pool
func (b *Block) PrintTransactions() {
	fmt.Println("====================\tTransactions\t====================")
	for i, t := range b.transactions {
		fmt.Printf("====================\tTransaction %d\t====================\n", i)
		fmt.Println(t.String())
	}
	fmt.Println("====================\tTransactions\t====================")
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

func (tx Transaction) String() string {
	return fmt.Sprintf("input: %x\noutput: %s\nsign: %x\npublic key: %x", tx.input, tx.output.String(), tx.sign, tx.publicKey[:])
}

// Transaction.Hash returns the hash of the transaction
func (tx Transaction) Hash() [32]byte {
	return sha256.Sum256([]byte(tx.String()))
}

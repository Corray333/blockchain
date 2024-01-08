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
	"sync"
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
		res += fmt.Sprintf("%x|%s|%x|%x|", t.output.pkh, t.output.data, t.publicKey, t.sign)
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
	hash := tx.Hash()
	ok := secp256k1.VerifySignature(tx.publicKey, hash[:], tx.sign[:64])
	if !ok {
		return errors.New("error while verifying transaction from " + fmt.Sprintf("%x", tx.publicKey))
	}
	mu := sync.Mutex{}
	mu.Lock()
	b.transactionPool = append(b.transactionPool, tx)
	if len(b.transactionPool) > 2000 {
		b.CreateBlock()
	}
	mu.Unlock()
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

func (b Block) GetTimestamp() time.Time {
	return b.timestamp
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
		copy(transaction.output.pkh[:], pkh)
		transaction.output.data = []byte(transactions[i+3])
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

func (b Block) StringForHash() string {
	return fmt.Sprintf("%x%x%s", b.prev[:], b.root[:], b.creatorAdress)
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
	return sha256.Sum256([]byte(b.StringForHash()))
}

// Output structure represents a transaction output
type Output struct {
	pkh  [20]byte
	data []byte
}

func (o Output) GetPKH() [20]byte {
	return o.pkh
}
func (o Output) GetData() []byte {
	return o.data
}

func (o Output) String() string {
	return fmt.Sprintf("pkh: %x, data: %s", o.pkh, string(o.data))
}

// Transact structure represents a transaction in blockchain
type Transaction struct {
	output    Output
	sign      []byte
	publicKey []byte
	timestamp time.Time
}

func (t Transaction) GetPublicKey() []byte {
	return t.publicKey
}
func (t Transaction) GetSign() []byte {
	return t.sign
}
func (t *Transaction) SetSign(sign []byte) {
	t.sign = sign
}
func (t Transaction) GetOutput() Output {
	return t.output
}
func (t Transaction) GetTimestamp() time.Time {
	return t.timestamp
}

// NewTransaction creates a new transaction
//
// pkh - public key hash of the receiver
//
// hash - hash of the string made of pkh+data+timestamp
//
// data - data of the transaction: marhsalled json
//
// sign - sign of the hash
//
// publicKey - public key of the sender
func NewTransaction(pkh [20]byte, data []byte, publicKey []byte, timestamp time.Time) Transaction {
	return Transaction{
		output: Output{
			pkh:  pkh,
			data: data,
		},
		publicKey: publicKey,
		timestamp: timestamp,
	}
}

func (tx Transaction) String() string {
	return fmt.Sprintf("output: %s\nsign: %x\npublic key: %x\ntimestamp: %s", tx.output.String(), tx.sign, tx.publicKey[:], tx.timestamp.String())
}

// Transaction.Hash returns the hash of the transaction
func (tx Transaction) Hash() [32]byte {
	return sha256.Sum256([]byte(string(tx.output.pkh[:]) + string(tx.output.data) + tx.timestamp.Format(time.RFC3339Nano)))
}

func (tx *Transaction) Sign(private []byte) error {
	hash := tx.Hash()
	sign, err := secp256k1.Sign(hash[:], private)
	if err != nil {
		return fmt.Errorf("error while signing transaction: %s", err.Error())
	}
	tx.sign = sign
	return nil
}

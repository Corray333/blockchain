package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
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
	BlockList       [][32]byte
	TransactionPool []Transaction
}

func (b Blockchain) GetLastBlock() [32]byte {
	return b.BlockList[len(b.BlockList)-1]
}

// NewBlockchain returnc a new blockchain
func NewBlockchain() *Blockchain {
	return &Blockchain{
		BlockList: [][32]byte{
			sha256.Sum256(nil),
		},
		TransactionPool: []Transaction{},
	}
}

// Blockchain.GetTransactionsString returns a string with all Transactions
func (b *Block) GetTransactionsString() string {
	res := ""
	for _, t := range b.Transactions {
		res += fmt.Sprintf("%x|%s|%x|%x|", t.Output.PKH, t.Output.Data, t.PublicKey, t.Sign)
	}
	return res
}

// Blockchain.PrintTransactions prints all Transactions in transaction pool
func (b *Blockchain) PrintTransactions() {
	fmt.Println("====================\tTransactions\t====================")
	for i, t := range b.TransactionPool {
		fmt.Printf("====================\tTransaction %d\t====================\n", i)
		fmt.Println(t.String())
	}
	fmt.Println("====================\tTransactions\t====================")
}

// Blockchain.NewTransaction creates a new transaction in transaction pool
func (b *Blockchain) NewTransaction(tx Transaction) error {
	hash := tx.Hash()
	ok := secp256k1.VerifySignature(tx.PublicKey, hash[:], tx.Sign[:64])
	if !ok {
		return errors.New("error while verifying transaction from " + fmt.Sprintf("%x", tx.PublicKey))
	}
	mu := sync.Mutex{}
	mu.Lock()
	b.TransactionPool = append(b.TransactionPool, tx)
	if len(b.TransactionPool) > 2000 {
		b.CreateBlock()
	}
	mu.Unlock()
	return nil
}

// CreateBlock creates a new block
func (b *Blockchain) CreateBlock() *Block {
	block := &Block{
		Prev:         b.BlockList[len(b.BlockList)-1],
		Root:         [32]byte{},
		Transactions: make([]Transaction, len(b.TransactionPool)),
		Timestamp:    time.Now(),
	}
	copy(block.Transactions, b.TransactionPool)
	b.TransactionPool = []Transaction{}
	hashes := make([][32]byte, len(block.Transactions))
	for i := range hashes {
		hashes[i] = block.Transactions[i].Hash()
	}
	block.Root = GetMerkleRoot(hashes)
	block.CreatorAddress = wallet.GetAddress()
	return block
}

// Block structure represents a block in a blockchain
type Block struct {
	Prev           [32]byte
	Root           [32]byte
	Transactions   []Transaction
	Timestamp      time.Time
	CreatorAddress string
	Level          int
}

func (b Block) GetTimestamp() time.Time {
	return b.Timestamp
}

func (b Block) GetLevel() int {
	return b.Level
}

// Block.Save function saves a block in store folder
func (b Block) Save() error {
	hash := b.Hash()
	file, err := os.Create(fmt.Sprintf("../store/blocks/%d-%x.blk", b.GetLevel(), hash))
	if err != nil {
		return err
	}
	res := fmt.Sprintf("%x|%x|%d|%s|%d\n", b.Prev, b.Root, b.Timestamp.UnixMicro(), b.CreatorAddress, b.Level)
	// TODO: save Transactions with compressed public keys
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

func LoadBlock(Data []byte) (*Block, error) {
	var block Block
	splitted := strings.Split(string(Data), "\n")
	blockData := strings.Split(splitted[0], "|")
	Transactions := strings.Split(splitted[1], "|")
	Prev, err := hex.DecodeString(blockData[0])
	if err != nil {
		return nil, fmt.Errorf(`error while loading block: %s`, err.Error())
	}
	copy(block.Prev[:], Prev)
	Root, err := hex.DecodeString(blockData[1])
	if err != nil {
		return nil, fmt.Errorf(`error while loading block: %s`, err.Error())
	}
	copy(block.Root[:], Root)
	microSeconds, err := strconv.ParseInt(blockData[2], 10, 64)
	if err != nil {
		return nil, fmt.Errorf(`error while loading block: %s`, err.Error())
	}
	block.Timestamp = time.UnixMicro(microSeconds)
	if err != nil {
		return nil, fmt.Errorf(`error while loading block: %s`, err.Error())
	}
	block.CreatorAddress = blockData[3]
	block.Level, err = strconv.Atoi(blockData[4])
	if err != nil {
		return nil, fmt.Errorf(`error while loading block: %s`, err.Error())
	}
	block.Transactions = []Transaction{}
	// TODO: load Transactions
	for i := 0; i < len(Transactions)-1; i += 6 {
		var transaction Transaction
		PKH, err := hex.DecodeString(Transactions[i+1])
		if err != nil {
			return nil, fmt.Errorf(`error while loading block: %s`, err.Error())
		}
		hash, err := hex.DecodeString(Transactions[i+2])
		if err != nil {
			return nil, fmt.Errorf(`error while loading block with hash "%x": %s`, hash, err.Error())
		}
		pub, err := hex.DecodeString(Transactions[i+4])
		if err != nil {
			return nil, fmt.Errorf(`error while loading block with hash "%x": %s`, hash, err.Error())
		}
		Sign, err := hex.DecodeString(Transactions[i+5])
		if err != nil {
			return nil, fmt.Errorf(`error while loading block with hash "%x": %s`, hash, err.Error())
		}
		copy(transaction.Output.PKH[:], PKH)
		transaction.Output.Data = Transactions[i+3]
		transaction.PublicKey = make([]byte, 65)
		transaction.Sign = make([]byte, 64)
		copy(transaction.PublicKey[:], pub)
		copy(transaction.Sign[:], Sign)
		block.Transactions = append(block.Transactions, transaction)
	}
	// TODO: check Data format
	return &block, nil
}

// GetMerkleRoot returns the merkle Root got by Transactions hashes
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
	return fmt.Sprintf("Prev hash: %x\nmerkle Root: %x\nTimestamp: %s\ncreator: %s\n", b.Prev[:], b.Root[:], b.Timestamp.String(), b.CreatorAddress)
}

func (b Block) StringForHash() string {
	return fmt.Sprintf("%x%x%s", b.Prev[:], b.Root[:], b.CreatorAddress)
}

// Block.PrintTransactions prints all Transactions in transaction pool
func (b *Block) PrintTransactions() {
	fmt.Println("====================\tTransactions\t====================")
	for i, t := range b.Transactions {
		fmt.Printf("====================\tTransaction %d\t====================\n", i)
		fmt.Println(t.String())
	}
	fmt.Println("====================\tTransactions\t====================")
}

// Block.Hash gets the hash of the block
func (b Block) Hash() [32]byte {
	return sha256.Sum256([]byte(b.StringForHash()))
}

// Output structure represents a transaction Output
type Output struct {
	PKH  [20]byte
	Data string
}

func (o Output) GetPKH() [20]byte {
	return o.PKH
}
func (o Output) GetData() string {
	return o.Data
}

func (o Output) String() string {
	return fmt.Sprintf("PKH: %x, Data: %s", o.PKH, string(o.Data))
}

// Transact structure represents a transaction in blockchain
type Transaction struct {
	Output    Output
	Sign      []byte
	PublicKey []byte
	Timestamp time.Time
}

// NewTransaction creates a new transaction
//
// # PKH - public key hash of the receiver
//
// hash - hash of the string made of PKH+Data+Timestamp
//
// Data - Data of the transaction: marhsalled json
//
// # Sign - Sign of the hash
//
// PublicKey - public key of the sender
func NewTransaction(PKH [20]byte, Data []byte, PublicKey []byte, Timestamp time.Time) Transaction {
	return Transaction{
		Output: Output{
			PKH:  PKH,
			Data: string(Data),
		},
		PublicKey: PublicKey,
		Timestamp: Timestamp,
	}
}

func (tx Transaction) String() string {
	return fmt.Sprintf("Output: %s\nSign: %x\npublic key: %x\nTimestamp: %s", tx.Output.String(), tx.Sign, tx.PublicKey[:], tx.Timestamp.String())
}

// Transaction.Hash returns the hash of the transaction
func (tx Transaction) Hash() [32]byte {
	return sha256.Sum256([]byte(string(tx.Output.PKH[:]) + string(tx.Output.Data) + tx.Timestamp.Format(time.RFC3339Nano)))
}

func (tx *Transaction) CreateSign(private []byte) error {
	hash := tx.Hash()
	Sign, err := secp256k1.Sign(hash[:], private)
	if err != nil {
		return fmt.Errorf("error while Signing transaction: %s", err.Error())
	}
	tx.Sign = Sign
	return nil
}

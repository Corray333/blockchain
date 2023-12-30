package main

import (
	"encoding/hex"

	"github.com/Corray333/blockchain/internal/blockchain"
)

func main() {
	// app := app.CreateApp()
	// fmt.Println(wallet.GetWallet().String())
	// hash := sha256.Sum256(nil)
	// sign, _ := secp256k1.Sign(hash[:], wallet.GetPrivateKey())
	// tx := blockchain.NewTransaction([32]byte{1, 2, 3}, [20]byte{1, 2, 3, 4}, hash, []byte("Hello, world!"), sign[:len(sign)-1], wallet.GetPublicKey())
	// app.Blockchain.NewTransaction(tx)
	// app.Blockchain.PrintTransactions()
	// b := app.Blockchain.CreateBlock()
	// b.Save()
	a := "159dfd6832cf077ec27d0e6d7923ba352157a2661a7ac51a9a2e4b7569a0abb6"
	b, _ := hex.DecodeString(a)
	var f [32]byte
	copy(f[:], b)
	block := blockchain.LoadBlock(f)
	block.PrintTransactions()

}

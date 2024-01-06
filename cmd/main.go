package main

import (
	"github.com/Corray333/blockchain/internal/app"
)

func main() {
	app := app.CreateApp()
	app.Run()
	// fmt.Println(wallet.GetWallet().String())
	// hash := sha256.Sum256(nil)
	// sign, _ := secp256k1.Sign(hash[:], wallet.GetPrivateKey())
	// tx := blockchain.NewTransaction([20]byte{1, 2, 3, 4}, hash, []byte("Hello, world!"), sign[:len(sign)-1], wallet.GetPublicKey())
	// app.Blockchain.NewTransaction(tx)
	// app.Blockchain.PrintTransactions()
	// b := app.Blockchain.CreateBlock()
	// b.Save()
}

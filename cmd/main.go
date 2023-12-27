package main

import (
	"fmt"

	"github.com/Corray333/blockchain/internal/wallet"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load("../.env")
	wallet.InitializeWallet()
	w := wallet.GetWallet()
	fmt.Println(len(string(w.PublicKey)))
	fmt.Printf("%x\n", w.PKH)
	fmt.Printf("%x\n", w.Address)
}

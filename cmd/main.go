package main

import (
	"fmt"

	"github.com/Corray333/blockchain/internal/wallet"
	"github.com/joho/godotenv"
)

// func GeneratePKH(public ecdsa.PublicKey) {
// 	text, _ := x509.MarshalPKIXPublicKey(public)
// 	hash := sha256.Sum256([]byte(pkh))
// 	pkh = append(pkh, hash[:4]...)
// 	fmt.Println(len(base58.Encode([]byte(pkh))))
// }

func main() {
	godotenv.Load("../.env")
	wallet.InitializeWallet()
	w := wallet.GetWallet()
	fmt.Println(len(string(w.PublicKey)))
	fmt.Printf("%x", w.PKH)
}

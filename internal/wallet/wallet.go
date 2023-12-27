package wallet

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"log/slog"
	"math/big"
	"math/rand"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/btcsuite/btcutil/base58"
)

// Wallet structure represents a wallet in network
type Wallet struct {
	key       *ecdsa.PrivateKey
	publicKey []byte
	pkh       [20]byte
	address   []byte
}

// Wallet.GetKey method returns a private key of the wallet
func (w Wallet) GetKey() *ecdsa.PrivateKey {
	return w.key
}

// Wallet.GetPKH method returns a public key hash of the wallet
func (w Wallet) GetPKH() [20]byte {
	return w.pkh
}

// wallet is a global variable containing wallet of the node
var wallet Wallet

// GetWallet function returns a global struct variable
func GetWallet() struct {
	Key       *ecdsa.PrivateKey
	PublicKey []byte
	PKH       [20]byte
	Address   []byte
} {
	return struct {
		Key       *ecdsa.PrivateKey
		PublicKey []byte
		PKH       [20]byte
		Address   []byte
	}{
		Key:       wallet.key,
		PublicKey: wallet.publicKey,
		PKH:       wallet.pkh,
		Address:   wallet.address,
	}
}

// InitializeWallet function fills wallet global variable with correct data, generated by a SEED_PHRASE from .env file
func InitializeWallet() {
	seed := GenerateSecretNumberBySeedPhrase(os.Getenv("SEED_PHRASE"))
	h := hmac.New(sha512.New, []byte(seed))
	hash := h.Sum(nil)
	pri := hash[:32]
	// TODO: learn, how to fix that
	// GenerateKey function is changed!!! First line, which changes private key, is delited
	private, err := ecdsa.GenerateKey(elliptic.P256(), bytes.NewReader(pri))
	if err != nil {
		panic(err)
	}
	wallet.key = private
	public := fmt.Sprintf("%x", private.X)
	if private.Y.Mod(private.Y, big.NewInt(2)) == big.NewInt(0) {
		public = "0" + public
	} else {
		public = "1" + public
	}
	wallet.publicKey = []byte(public)
	publicHash := sha256.Sum256(wallet.publicKey)
	copy(wallet.pkh[:], publicHash[:])
	wallet.address = GenerateWalletAddress(wallet.pkh)
}

// GenerateSecretNumberBySeedPhrase generates a secret number from the seed phrase
func GenerateSecretNumberBySeedPhrase(phrase string) string {
	seed := ""
	recPhrase := strings.Split(phrase, " ")
	f, err := os.ReadFile("../configs/wordlist.txt")
	if err != nil {
		slog.Error(fmt.Sprintf("error while reading wordlist: %s", err.Error()))
		panic(err)
	}
	words := strings.Split(string(f), "\n")
	for _, v := range recPhrase {
		i := slices.Index(words, v)
		if i == -1 {
			slog.Error("wrong recovery phrase")
			panic("wrong recovery phrase")
		}
		temp := strconv.FormatInt(int64(i), 2)
		for len(temp) < 11 {
			temp = "0" + temp
		}
		seed += temp
	}
	return seed[4:]
}

// GenerateSeedPhrase generates a seed phrase from the seed number
func GenerateSeedPhrase(seed string) string {
	hash := sha256.Sum256([]byte(seed))
	firstByte := strings.Replace(fmt.Sprintf("%b", hash[:4]), " ", "", -1)[1:5]
	seed = firstByte + seed
	f, err := os.ReadFile("../configs/wordlist.txt")
	if err != nil {
		slog.Error(fmt.Sprintf("error while reading wordlist: %s", err.Error()))
		panic(err)
	}
	words := strings.Split(string(f), "\n")
	seedPhrase := ""
	for i := 0; i < len(seed); i += 11 {
		num, _ := strconv.ParseInt(seed[i:i+11], 2, 0)
		seedPhrase += words[num] + " "
	}
	return seedPhrase[:len(seedPhrase)-1]
}

// GenerateSeedPhraseFromSkratch generates a randem seed number and uses it to generate a seed phrase
func GenerateSeedPhraseFromSkratch() string {
	seed := ""
	for i := 0; i < 128; i++ {
		seed += strconv.Itoa(rand.Int() % 2)
	}
	return GenerateSeedPhrase(seed)
}

// TODO: change this function to make generated address smaller
func GenerateWalletAddress(pkh [20]byte) []byte {
	pkhHash := sha256.Sum256(pkh[:])
	pkhWithCheckSum := append(pkh[:], pkhHash[:4]...)
	return []byte(base58.Encode(pkhWithCheckSum))
}

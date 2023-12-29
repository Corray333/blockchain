package wallet

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/btcsuite/btcutil/base58"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
)

// Wallet structure represents a wallet in network
type Wallet struct {
	privateKey [32]byte
	publicKey  [65]byte
	pkh        [20]byte
	address    string
}

// Wallet.GetKey method returns a private key of the wallet
func GetKey() []byte {
	return wallet.privateKey[:]
}

func GetPublicKey() []byte {
	r := make([]byte, len(wallet.publicKey))
	copy(r, wallet.publicKey[:])
	return r
}

// Wallet.GetPKH method returns a public key hash of the wallet
func GetPKH() [20]byte {
	return wallet.pkh
}

func GetAddress() string {
	return wallet.address
}

func (w Wallet) String() string {
	return fmt.Sprintf("Private key:\t%x\nPublic key:\t%x\nPKH:    \t%x\nAddress:\t%s", w.privateKey, w.publicKey, w.pkh, w.address)
}

// wallet is a global variable containing wallet of the node
var wallet Wallet

func GetWallet() Wallet {
	return wallet
}

// InitializeWallet function fills wallet global variable with correct data, generated by a SEED_PHRASE from .env file
func InitializeWallet() {
	seed := GenerateSecretNumberBySeedPhrase(os.Getenv("SEED_PHRASE"))
	h := hmac.New(sha512.New, []byte(seed))
	hash := h.Sum(nil)
	copy(wallet.privateKey[:], hash)
	msg := sha256.Sum256(nil)
	sig, err := secp256k1.Sign(msg[:], hash[:32])
	if err != nil {
		slog.Error("error while generating private and public keys: " + err.Error())
		panic("error while generating private and public keys: " + err.Error())
	}
	pub, err := secp256k1.RecoverPubkey(msg[:], sig)
	if err != nil {
		slog.Error("error while generating private and public keys: " + err.Error())
		panic("error while generating private and public keys: " + err.Error())
	}
	ok := secp256k1.VerifySignature(pub, msg[:], sig[:len(sig)-1])
	if !ok {
		slog.Error("error while generating private and public keys")
		panic("error while generating private and public keys")
	}
	copy(wallet.publicKey[:], pub)

	publicHash := sha256.Sum256(wallet.publicKey[:])
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
func GenerateWalletAddress(pkh [20]byte) string {
	pkhHash := sha256.Sum256(pkh[:])
	pkhWithCheckSum := append(pkh[:], pkhHash[:4]...)
	return base58.Encode(pkhWithCheckSum)
}

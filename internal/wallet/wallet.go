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
	"math/rand"
	"os"
	"slices"
	"strconv"
	"strings"
)

type Wallet struct {
	key *ecdsa.PrivateKey
	PHK [20]byte
}

var wallet Wallet

func GetWallet() Wallet {
	return wallet
}

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
}

func GenerateSecretNumberBySeedPhrase(phrase string) string {
	seed := ""
	// recPhrase := strings.Split(os.Getenv("RECOVERY_PHRASE"), " ")
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
	// hash := sha256.Sum256([]byte(seed[4:]))
	// fmt.Println(strings.Replace(fmt.Sprintf("%b", hash[:4]), " ", "", -1)[1:5])
	return seed[4:]
}

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

func GenerateSeedPhraseFromSkratch() string {
	seed := ""
	for i := 0; i < 128; i++ {
		seed += strconv.Itoa(rand.Int() % 2)
	}
	return GenerateSeedPhrase(seed)
}

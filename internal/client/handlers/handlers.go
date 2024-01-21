package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/Corray333/blockchain/internal/app"
	"github.com/Corray333/blockchain/internal/blockchain"
	"github.com/go-chi/chi/v5"
)

func GetLastTransaction(w http.ResponseWriter, r *http.Request) {
	marshalled, err := json.Marshal(app.Application.Blockchain.GetLastTransaction())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(marshalled)
}

func GetTransactionByBlockAndHash(w http.ResponseWriter, r *http.Request) {
	if chi.URLParam(r, "block") == "none" {
		tx := app.Application.Blockchain.GetTransactionByHash(chi.URLParam(r, "transaction"))
		if tx == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		marshalled, err := json.Marshal(tx)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(marshalled)
		return
	}

	blockData, err := os.ReadFile(chi.URLParam(r, "block"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	block, err := blockchain.LoadBlock(blockData)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	for _, tx := range block.Transactions {
		if fmt.Sprintf("%x", tx.Hash()) == chi.URLParam(r, "transaction") {
			marshalled, err := json.Marshal(tx)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Write(marshalled)
			return
		}
	}
	w.WriteHeader(http.StatusNotFound)
}

func GetLastBlock(w http.ResponseWriter, r *http.Request) {
	block, err := os.ReadFile(fmt.Sprintf("%d-%x.blk", len(app.Application.Blockchain.BlockList)-1, app.Application.Blockchain.BlockList[len(app.Application.Blockchain.BlockList)-1]))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(block)
}
func GetCurrentBlock(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(*app.Application.Blockchain.CreateBlock())
}
func LogIn(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	slog.Info("Log in")
	req := struct {
		RecoveryPhrase string `json:"recoveryPhrase"`
	}{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	app := app.CreateApp(req.RecoveryPhrase)
	if app == nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	app.Run()
	w.WriteHeader(http.StatusOK)
}

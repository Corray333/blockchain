package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Corray333/blockchain/internal/app"
)

func GetLastTransaction(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(app.Application.Blockchain.GetLastTransaction())
}

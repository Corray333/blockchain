package client

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/Corray333/blockchain/internal/client/handlers"
	"github.com/go-chi/chi/v5"
)

type Server struct {
	portServer int
	portClient int
}

func NewServer(server, client int) Server {
	return Server{
		portServer: server,
		portClient: client,
	}
}

func (s Server) Run() {
	go s.runServer()
	go s.runClient()
}
func (s Server) runServer() {
	r := chi.NewRouter()
	r.Get("/transactions/last", handlers.GetLastTransaction)
	panic(http.ListenAndServe("127.0.0.1:"+strconv.Itoa(s.portServer), r))
}
func (s Server) runClient() {
	http.Handle("/", http.FileServer(http.Dir("../frontend/dist")))
	slog.Info("Client is running on port " + strconv.Itoa(s.portClient))
	panic(http.ListenAndServe("127.0.0.1:"+strconv.Itoa(s.portClient), nil))
}

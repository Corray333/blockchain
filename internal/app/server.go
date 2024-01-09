package app

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/Corray333/blockchain/internal/blockchain"
	"github.com/Corray333/blockchain/internal/wallet"
)

const (
	Follower = iota
	Candidate
	Master
)

type Node struct {
	wallet string
}

type ServerP2P struct {
	port        int                 // port of the node for P2P connection
	connections map[string]Node     // map of nodes
	walletsBL   map[string]struct{} // black list of wallets
}

type ServerHTTP struct {
	port int
}

func GetOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}

func (a *App) Run() {
	go a.RunClient()
	listener, err := net.Listen("tcp", GetOutboundIP()+":"+strconv.Itoa(a.ServerP2P.port))
	fmt.Printf("Starting server %s\n", listener.Addr().String())
	if err != nil {
		slog.Error("error while starting server:" + err.Error())
		panic(err)
	}
	// go a.ConnectWithBootnodes()
	for {
		conn, err := listener.Accept()
		if err != nil {
			slog.Error("error while accepting connection:" + err.Error())
			continue
		}
		go a.handleConnection(conn)
	}
}

func (a *App) handleConnection(conn net.Conn) {
	defer conn.Close()
	slog.Info("new connection from " + conn.RemoteAddr().String())
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		slog.Error("error while reading from connection:" + err.Error())
		return
	}
	slog.Info("received request from " + conn.RemoteAddr().String() + ": " + string(buffer[:n]))
	err = a.HandleRequest(conn, buffer[:n])
	if err != nil {
		slog.Error("error while handling request:" + err.Error())
		return
	}
}

func (a *App) ConnectWithBootnodes() error {
	if len(a.Config.BootNodes) == 0 {
		return errors.New("error while connecting to network: no boot nodes")
	}
	for _, v := range a.Config.BootNodes {
		err := a.ConnectDirectly(v)
		if err != nil {
			slog.Error(err.Error())
			continue
		}
		return nil
	}
	return errors.New("error while connecting to network: all the boot nodes are unavalible, try to connect directly")
}

func (a *App) ConnectDirectly(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return errors.New("error while connecting to network:" + err.Error())
	}
	defer conn.Close()
	query := map[string]interface{}{
		"query": "01",
	}
	bytesQuery, err := json.Marshal(query)
	if err != nil {
		return errors.New("error while marshalling query to connect to network:" + err.Error())
	}
	if _, err := conn.Write(bytesQuery); err != nil {
		return errors.New("error while writing message to connect to network:" + err.Error())
	}
	buf := make([]byte, 524288)
	n, err := conn.Read(buf)
	if err != nil {
		return errors.New("error while reading response to connect to network:" + err.Error())
	}
	splited := strings.Split(string(buf[:n]), "|")
	for _, v := range splited {
		if v == "" {
			continue
		}
		a.ServerP2P.connections[v] = Node{}
	}

	conn, err = net.Dial("tcp", addr)
	if err != nil {
		return errors.New("error while connecting to network:" + err.Error())
	}
	defer conn.Close()
	query = map[string]interface{}{
		"query":  "02",
		"wallet": wallet.GetAddress(),
		"from":   GetOutboundIP() + ":" + strconv.Itoa(a.ServerP2P.port),
	}
	bytesQuery, err = json.Marshal(query)
	if err != nil {
		return errors.New("error while marshalling query to connect to network:" + err.Error())
	}
	if _, err := conn.Write(bytesQuery); err != nil {
		return errors.New("error while writing message to connect to network:" + err.Error())
	}
	buf = make([]byte, 128)
	n, err = conn.Read(buf)
	if err != nil {
		return errors.New("error while reading response to connect to network:" + err.Error())
	}
	if string(buf[:n]) != "ok" {
		return errors.New("error while connecting to network")
	}

	return nil
}

func (a *App) RunClient() {
	for {
		query := InputString()
		splitted := strings.Split(query, " ")
		switch splitted[0] {
		case "/create-transaction":
			pkh := [20]byte{}
			copy(pkh[:], splitted[2])
			tx := blockchain.NewTransaction(pkh, []byte(splitted[1]), wallet.GetPublicKey(), time.Now())
			if err := tx.Sign(wallet.GetPrivateKey()); err != nil {
				slog.Error(err.Error(), "type", "blockchain", "process", "create transaction")
			}
			if err := a.Blockchain.NewTransaction(tx); err != nil {
				slog.Error(err.Error(), "type", "blockchain", "process", "create transaction")
			}
			SendTransactionToNetwork(a, tx)
		case "/show-transactions":
			a.Blockchain.PrintTransactions()
		case "/show-nodes":
			fmt.Println("====================\tNodes\t====================")
			for k := range a.ServerP2P.connections {
				fmt.Println(k)
			}
		case "/show-wallet":
			fmt.Println("====================\tWallet\t====================")
			wallet.Print()
		case "/connect-through":
			if err := a.ConnectDirectly(splitted[1]); err != nil {
				slog.Error(err.Error(), "type", "network", "process", "connect through")
			}
		case "/clear":
			cmd := exec.Command("clear") //Linux example, its tested
			cmd.Stdout = os.Stdout
			cmd.Run()
		case "exit":
			os.Exit(0)
		}
	}
}

func InputString() string {
	msg, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	return strings.Replace(msg, "\n", "", -1)
}

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	handlers "github.com/kube-carbonara/api-server/handlers"
	"github.com/rancher/remotedialer"
	"github.com/sirupsen/logrus"
)

func init() {
}

func authorizer(req *http.Request) (string, bool, error) {
	id := req.Header.Get("x-tunnel-id")
	return id, id != "", nil
}

func handleServiceEvents(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{}
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
	}
}

func main() {

	var (
		addr      string
		peerID    string
		peerToken string
		peers     string
		debug     bool
	)
	flag.StringVar(&addr, "listen", ":8099", "Listen address")
	flag.StringVar(&peerID, "id", "", "Peer ID")
	flag.StringVar(&peerToken, "token", "", "Peer Token")
	flag.StringVar(&peers, "peers", "", "Peers format id:token:url,id:token:url")
	flag.BoolVar(&debug, "debug", false, "Enable debug logging")
	flag.Parse()
	if debug {
		logrus.SetLevel(logrus.DebugLevel)
		remotedialer.PrintTunnelData = true
	}
	handler := remotedialer.New(authorizer, remotedialer.DefaultErrorWriter)
	handler.PeerToken = peerToken
	handler.PeerID = peerID
	for _, peer := range strings.Split(peers, ",") {
		parts := strings.SplitN(strings.TrimSpace(peer), ":", 3)
		if len(parts) != 3 {
			continue
		}
		handler.AddPeer(parts[2], parts[0], parts[1])
	}

	router := mux.NewRouter()
	router.Handle("/connect", handler)
	router.HandleFunc("/services", handleServiceEvents)
	router.HandleFunc("/clusters/{id}/{path:.*}", func(rw http.ResponseWriter, req *http.Request) {
		handlers.ClientHandler{}.Handle(handler, rw, req)
	})

	fmt.Println("Listening on ", addr)
	http.ListenAndServe(addr, router)
}

package connections

import (
	"flag"
	"net/http"
	"strings"

	"github.com/rancher/remotedialer"
	"github.com/sirupsen/logrus"
)

var (
	peerID    string
	peerToken string
	peers     string
	debug     bool
)

type RemoteDialerClient struct {
	Server *remotedialer.Server
}

func (c RemoteDialerClient) NewServer() *RemoteDialerClient {
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
	c.Server = handler
	return &c
}

func authorizer(req *http.Request) (string, bool, error) {
	id := req.Header.Get("x-tunnel-id")
	return id, id != "", nil
}

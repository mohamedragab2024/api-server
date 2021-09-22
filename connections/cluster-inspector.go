package connections

import (
	ctx "context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/kube-carbonara/api-server/controllers"
	"github.com/kube-carbonara/api-server/handlers"
	"github.com/rancher/remotedialer"
	"github.com/rancher/wrangler/pkg/ticker"
	"github.com/sirupsen/logrus"
)

type ClusterInsector struct {
	Dialer *remotedialer.Server
	mu     *sync.Mutex
}

func (c ClusterInsector) Register(clientId string, syncDuration time.Duration) {
	go c.inspect(clientId, syncDuration)
}

func (c ClusterInsector) inspect(clientId string, syncDuration time.Duration) {
	// skip in case of registeration in progress

	for range ticker.Context(ctx.Background(), syncDuration) {
		controller := controllers.ClusterController{}
		cluster, err := controller.GetById(clientId)
		// stop goroutine if the cluster has been removed
		if err != nil && err.Error() == "cluster not found" {
			break
		}
		if cluster.MetricsCache.Provider == "" {
			logrus.Info(fmt.Sprintf("skip health check until cluster %s fully registered", clientId))
		} else {
			err := c.healthCheck(clientId)
			if err != nil {
				logrus.Error(err)
				c.cacheLastUpdate(clientId, false, err.Error())

			} else {
				logrus.Info(fmt.Sprintf("Cluster %s is Healthy", clientId))
				c.cacheLastUpdate(clientId, true, "Cluster is Healthy")

			}
		}

	}

}

func (c ClusterInsector) healthCheck(clientId string) error {
	ClientHandler := handlers.ClientHandler{}
	timeout := "50"
	scheme := "http"
	host := "127.0.0.1:1323"
	url := fmt.Sprintf("%s://%s", scheme, host)
	client := ClientHandler.GetClient(c.Dialer, clientId, timeout)
	if client == nil {
		err := errors.New("failed to get client")
		log.Print("upgrade:", err)
		return err
	}
	resp, err := client.Get(url)
	if err != nil {
		logrus.Errorf("[%03d] REQ ERR t=%s %s: %v", clientId, timeout, url, err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		logrus.Errorf("[%03d] REQ ERR t=%s %s: %v", clientId, timeout, url, err)
		return fmt.Errorf("failed to call cluster status code : %d", resp.StatusCode)
	}
	return nil
}

func (c ClusterInsector) Aknowlegement(w http.ResponseWriter, r *http.Request, syncDuration time.Duration) {
	var upgrader = websocket.Upgrader{}
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer conn.Close()
	for range ticker.Context(ctx.Background(), syncDuration) {
		clusters := controllers.ClusterController{}.GetResultList()
		conn.WriteJSON(clusters)
	}

}

func (c ClusterInsector) cacheLastUpdate(clientId string, isHealthy bool, message string) error {
	controller := controllers.ClusterController{}
	model, err := controller.GetById(clientId)
	if err != nil {
		logrus.Errorf("error updaing status of %s %v", clientId, err)
		return fmt.Errorf("error updaing status of %s %v", clientId, err)
	}
	model.IsConnected = isHealthy
	model.LastSyncMessage = message
	controller.SaveChanges(model)
	return nil
}

func (c ClusterInsector) OnStartUp() {
	controller := controllers.ClusterController{}
	result := controller.GetList()

	for _, v := range result {
		c.Register(v.Name, 30*time.Second)
	}
}

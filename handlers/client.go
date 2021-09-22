package handlers

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/mux"
	"github.com/kube-carbonara/api-server/utils"
	"github.com/rancher/remotedialer"
	"github.com/sirupsen/logrus"
)

var (
	clients = map[string]*http.Client{}
	l       sync.Mutex
	counter int64
)

type ClientHandler struct {
}

func (h ClientHandler) Handle(server *remotedialer.Server, rw http.ResponseWriter, req *http.Request) {
	if !(AuthorizationHandler{}).IsAuthorized(rw, req) {
		return
	}
	timeout := req.URL.Query().Get("timeout")
	if timeout == "" {
		timeout = "15"
	}
	queryParams := req.URL.Query().Encode()
	scheme := "http"
	host := "127.0.0.1:1323"
	vars := mux.Vars(req)
	clientKey := vars["id"]
	url := fmt.Sprintf("%s://%s/%s?%s", scheme, host, vars["path"], queryParams)
	client := h.GetClient(server, clientKey, timeout)
	switch req.Method {
	case http.MethodGet:
		get(server, rw, req, client, clientKey, timeout, url)
	case http.MethodPost:
		post(server, rw, req, client, clientKey, timeout, url)
	case http.MethodDelete:
		delete(server, rw, req, client, clientKey, timeout, url)
	case http.MethodPut:
		update(server, rw, req, client, clientKey, timeout, url)
	default:
		remotedialer.DefaultErrorWriter(rw, req, 405, errors.New("method not allowed"))
	}
	id := atomic.AddInt64(&counter, 1)
	logrus.Infof("[%03d] REQ t=%s %s", id, timeout, url)
}

func (h ClientHandler) GetClient(server *remotedialer.Server, clientKey, timeout string) *http.Client {
	l.Lock()
	defer l.Unlock()

	key := fmt.Sprintf("%s/%s", clientKey, timeout)
	client := clients[key]
	if client != nil {
		return client
	}

	dialer := server.Dialer(clientKey, 15*time.Second)
	client = &http.Client{
		Transport: &http.Transport{
			Dial: dialer,
		},
	}
	if timeout != "" {
		t, err := strconv.Atoi(timeout)
		if err == nil {
			client.Timeout = time.Duration(t) * time.Second
		}
	}

	clients[key] = client
	return client
}

func get(server *remotedialer.Server, rw http.ResponseWriter, req *http.Request, client *http.Client, id string, timeout string, url string) {
	resp, err := client.Get(url)
	if err != nil {
		logrus.Errorf("[%03d] REQ ERR t=%s %s: %v", id, timeout, url, err)
		remotedialer.DefaultErrorWriter(rw, req, 500, err)
		return
	}
	defer resp.Body.Close()

	logrus.Infof("[%03d] REQ OK t=%s %s", id, timeout, url)
	rw.WriteHeader(resp.StatusCode)
	rw.Header().Set("Content-Type", "application/json")

	io.Copy(rw, resp.Body)
	logrus.Infof("[%03d] REQ DONE t=%s %s", id, timeout, url)
}

func post(server *remotedialer.Server, rw http.ResponseWriter, req *http.Request, client *http.Client, id string, timeout string, url string) {
	resp, err := client.Post(url, utils.APPLICATION_JSON, req.Body)
	if err != nil {
		logrus.Errorf("[%03d] REQ ERR t=%s %s: %v", id, timeout, url, err)
		remotedialer.DefaultErrorWriter(rw, req, 500, err)
		return
	}
	defer resp.Body.Close()

	logrus.Infof("[%03d] REQ OK t=%s %s", id, timeout, url)
	rw.WriteHeader(resp.StatusCode)
	rw.Header().Set("Content-Type", "application/json")
	io.Copy(rw, resp.Body)
	logrus.Infof("[%03d] REQ DONE t=%s %s", id, timeout, url)
}

func delete(server *remotedialer.Server, rw http.ResponseWriter, req *http.Request, client *http.Client, id string, timeout string, url string) {
	newReq, err := http.NewRequest(http.MethodDelete, url, req.Body)

	if err != nil {
		logrus.Errorf("[%03d] REQ ERR t=%s %s: %v", id, timeout, url, err)
		remotedialer.DefaultErrorWriter(rw, req, 500, err)
		return
	}
	resp, err := client.Do(newReq)
	if err != nil {
		logrus.Errorf("[%03d] REQ ERR t=%s %s: %v", id, timeout, url, err)
		remotedialer.DefaultErrorWriter(rw, req, 500, err)
		return
	}
	defer resp.Body.Close()
	logrus.Infof("[%03d] REQ OK t=%s %s", id, timeout, url)
	rw.WriteHeader(resp.StatusCode)
	rw.Header().Set("Content-Type", "application/json")
	io.Copy(rw, resp.Body)
	logrus.Infof("[%03d] REQ DONE t=%s %s", id, timeout, url)
}

func update(server *remotedialer.Server, rw http.ResponseWriter, req *http.Request, client *http.Client, id string, timeout string, url string) {
	newReq, err := http.NewRequest(http.MethodPut, url, req.Body)

	if err != nil {
		logrus.Errorf("[%03d] REQ ERR t=%s %s: %v", id, timeout, url, err)
		remotedialer.DefaultErrorWriter(rw, req, 500, err)
		return
	}
	resp, err := client.Do(newReq)
	if err != nil {
		logrus.Errorf("[%03d] REQ ERR t=%s %s: %v", id, timeout, url, err)
		remotedialer.DefaultErrorWriter(rw, req, 500, err)
		return
	}
	defer resp.Body.Close()
	logrus.Infof("[%03d] REQ OK t=%s %s", id, timeout, url)
	rw.WriteHeader(resp.StatusCode)
	rw.Header().Set("Content-Type", "application/json")
	io.Copy(rw, resp.Body)
	logrus.Infof("[%03d] REQ DONE t=%s %s", id, timeout, url)
}

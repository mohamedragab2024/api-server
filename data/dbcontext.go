package data

import (
	context "context"
	"log"
	"strings"
	"time"

	"github.com/kube-carbonara/api-server/utils"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type DBContext struct {
	client clientv3.Client
}

func (c DBContext) new() *DBContext {
	config := utils.NewConfig()
	nodes := strings.Split(config.EtcdNodes, ",")
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   nodes,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatal(err)
		panic(err)
	}

	c.client = *cli

	return &c
}

func (db DBContext) Set(key string, value interface{}) error {
	c := db.new()
	defer c.client.Close()
	v := utils.StructToJson(value)
	_, err := c.client.Put(context.TODO(), key, string(v))
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
	return err
}

func (db DBContext) Get(key string) []byte {
	c := db.new()
	defer c.client.Close()
	resp, err := c.client.Get(context.TODO(), key)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}

	if len(resp.Kvs) == 0 {
		return nil
	}

	return resp.Kvs[0].Value
}

func (db DBContext) GetRangePrefixedOfType(prefix string) [][]byte {
	c := db.new()
	defer c.client.Close()
	resp, err := c.client.Get(context.TODO(), prefix, clientv3.WithPrefix())
	if err != nil {
		log.Fatal(err)
		panic(err)
	}

	var result [][]byte
	for _, v := range resp.Kvs {
		result = append(result, v.Value)
	}
	return result

}

func (db DBContext) Delete(key string) {
	c := db.new()
	defer c.client.Close()
	_, err := c.client.Delete(context.TODO(), key)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
}

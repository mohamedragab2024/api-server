package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/kube-carbonara/api-server/models"
	"github.com/labstack/echo/v4"
)

func init() {
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")
}

var connectionLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connection Lost: %s\n", err.Error())
}

func main() {
	options := mqtt.NewClientOptions()
	options.AddBroker("tcp://localhost:1883")
	options.SetClientID("go_server")
	options.OnConnect = connectHandler
	options.OnConnectionLost = connectionLostHandler
	client := mqtt.NewClient(options)
	token := client.Connect()
	if token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	e := echo.New()
	e.GET("/", func(context echo.Context) error {
		query := context.Request().URL.Query()
		host := query.Get("client")
		path := query.Get("path")
		if host == "" {
			return context.String(http.StatusBadRequest, "missing parameter Path")
		}

		req := models.ServerRequest{
			Path:         path,
			Verb:         "GET",
			ResourceType: "namespace",
		}

		toSend, err := json.Marshal(req)
		if err != nil {
			fmt.Printf("Error encoding the request as JSON: %s\n", err.Error())
			return context.String(http.StatusInternalServerError, err.Error())
		}
		var response models.Response

		token := client.Publish("clients/"+host, 0, false, string(toSend))
		token.Wait()
		subToken := client.Subscribe("clients/"+host, 0, func(client mqtt.Client, msg mqtt.Message) {
			err := json.Unmarshal(msg.Payload(), &response)
			if err != nil {
				fmt.Print(err.Error())
			}
			if response.Prefix == "" {
				return
			}

		})

		subToken.Wait()
		if subToken.Error() != nil {
			fmt.Printf("Error subscribing to clients/%s - %s\n", client, subToken.Error())
			return context.String(http.StatusInternalServerError, err.Error())
		}

		count := 0
		for response.Status == 0 && count < 80 {
			time.Sleep(250 * time.Millisecond)
			count++

		}

		if response.Status <= 0 {
			return context.HTML(http.StatusInternalServerError, "<p>Failed to call remote server</p>")
		}

		if response.Status > 200 {
			return context.JSON(response.Status, response)
		}

		return context.JSON(http.StatusOK, response)
	})
	e.Logger.Fatal(e.Start(":8099"))
}

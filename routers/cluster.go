package routers

import (
	"github.com/gorilla/mux"
	"github.com/kube-carbonara/api-server/controllers"
	httpverbs "github.com/kube-carbonara/api-server/httpverbs"
)

type ClusterRouter struct{}

func (r ClusterRouter) Handle(router *mux.Router) {
	controller := controllers.ClusterController{}
	router.HandleFunc("/clusters", controller.GetAll).Methods(httpverbs.MethodGet)
	router.HandleFunc("/clusters/{id}", controller.GetOne).Methods(httpverbs.MethodGet)
	router.HandleFunc("/clusters/{id}", controller.Delete).Methods(httpverbs.MethodDelete)
	router.HandleFunc("/clusters/{id}", controller.Update).Methods(httpverbs.MethodPut)
	router.HandleFunc("/clusters", controller.Create).Methods(httpverbs.MethodPost)
	//handle as well / at the end
	router.HandleFunc("/clusters/", controller.GetAll).Methods(httpverbs.MethodGet)
	router.HandleFunc("/clusters/{id}/", controller.GetOne).Methods(httpverbs.MethodGet)
	router.HandleFunc("/clusters/{id}/", controller.Delete).Methods(httpverbs.MethodDelete)
	router.HandleFunc("/clusters/{id}/", controller.Update).Methods(httpverbs.MethodPut)
	router.HandleFunc("/clusters/", controller.Create).Methods(httpverbs.MethodPost)
	router.HandleFunc("/clusters/{id}/Config", controller.ConfigFile).Methods(httpverbs.MethodGet)
}

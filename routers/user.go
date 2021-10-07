package routers

import (
	"github.com/gorilla/mux"
	"github.com/kube-carbonara/api-server/controllers"
	httpverbs "github.com/kube-carbonara/api-server/httpverbs"
)

type UserRouter struct{}

func (r UserRouter) Handle(router *mux.Router) {
	controller := controllers.UsersController{}
	router.HandleFunc("/users", controller.GetAll).Methods(httpverbs.MethodGet)
	router.HandleFunc("/users/{id}", controller.GetOne).Methods(httpverbs.MethodGet)
	router.HandleFunc("/users/{id}", controller.Delete).Methods(httpverbs.MethodDelete)
	router.HandleFunc("/users", controller.Create).Methods(httpverbs.MethodPost)
	router.HandleFunc("/users/changePassword/", controller.ChangePassword).Methods(httpverbs.MethodPut)
	//handle as well / at the end
	router.HandleFunc("/users/", controller.GetAll).Methods(httpverbs.MethodGet)
	router.HandleFunc("/users/{id}/", controller.GetOne).Methods(httpverbs.MethodGet)
	router.HandleFunc("/users/{id}/", controller.Delete).Methods(httpverbs.MethodDelete)
	router.HandleFunc("/users/", controller.Create).Methods(httpverbs.MethodPost)
	router.HandleFunc("/users/", controller.Update).Methods(httpverbs.MethodPut)

	router.HandleFunc("/users/changePassword/", controller.ChangePassword).Methods(httpverbs.MethodPut)
}

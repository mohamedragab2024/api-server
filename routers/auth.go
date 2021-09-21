package routers

import (
	"github.com/gorilla/mux"
	"github.com/kube-carbonara/api-server/controllers"
	httpverbs "github.com/kube-carbonara/api-server/httpverbs"
)

type AuthRouter struct{}

func (r AuthRouter) Handle(router *mux.Router) {
	controller := controllers.AuthController{}
	router.HandleFunc("/auth", controller.Login).Methods(httpverbs.MethodPost)
}

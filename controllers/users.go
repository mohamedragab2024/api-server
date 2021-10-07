package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	crypto "github.com/kube-carbonara/api-server/crypto"
	data "github.com/kube-carbonara/api-server/data"
	"github.com/kube-carbonara/api-server/handlers"
	"github.com/kube-carbonara/api-server/models"
	uuid "github.com/satori/go.uuid"
)

const (
	UserPrefix = "User-"
)

type UsersController struct{}

func (c UsersController) GetAll(rw http.ResponseWriter, r *http.Request) {
	if !(handlers.AuthorizationHandler{}).IsAuthorized(rw, r) {
		return
	}
	result := []models.Users{}
	db := data.DBContext{}.GetRangePrefixedOfType(UserPrefix)
	for _, v := range db {
		var model models.Users
		json.Unmarshal(v, &model)
		result = append(result, model)
	}

	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(result)

}

func (c UsersController) GetOne(rw http.ResponseWriter, r *http.Request) {
	if !(handlers.AuthorizationHandler{}).IsAuthorized(rw, r) {
		return
	}
	id := mux.Vars(r)["id"]
	var model = models.Users{}
	db := data.DBContext{}.GetRangePrefixedOfType(fmt.Sprintf("%s%s", UserPrefix, id))
	if len(db) == 0 {
		rw.WriteHeader(http.StatusNotFound)
		return
	}
	json.Unmarshal(db[0], &model)
	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(model)
}

func (c UsersController) Delete(rw http.ResponseWriter, r *http.Request) {
	if !(handlers.AuthorizationHandler{}).IsAuthorized(rw, r) {
		return
	}
	id := mux.Vars(r)["id"]
	db := data.DBContext{}.GetRangePrefixedOfType(fmt.Sprintf("%s%s", UserPrefix, id))
	if len(db) == 0 {
		rw.WriteHeader(http.StatusNotFound)
		return
	}
	var model = models.Users{}
	json.Unmarshal(db[0], &model)
	data.DBContext{}.Delete(fmt.Sprintf("%s%s-%s", UserPrefix, id, model.UserId))
	rw.WriteHeader(http.StatusNoContent)
}

func (c UsersController) Create(rw http.ResponseWriter, r *http.Request) {
	if !(handlers.AuthorizationHandler{}).IsAuthorized(rw, r) {
		return
	}
	uuid := uuid.NewV4()
	var model models.Users
	err := json.NewDecoder(r.Body).Decode(&model)
	if err != nil {
		fmt.Print("error in decoding")
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	if check, len := validation(model); len > 0 {
		fmt.Print("error in validation")
		http.Error(rw, check, http.StatusBadRequest)
		return
	}

	model.Password, err = crypto.HashPassword(model.Password)

	if err != nil {
		fmt.Print("error in crypto")
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	model.UserId = uuid.String()
	if c.userIsExists(model.UserName) {
		http.Error(rw, fmt.Sprintf("%s already exists", model.UserName), http.StatusConflict)
		return
	}

	data.DBContext{}.Set(fmt.Sprintf("%s%s-%s", UserPrefix, model.UserName, model.UserId), model)
	rw.WriteHeader(http.StatusCreated)
}

func (c UsersController) Update(rw http.ResponseWriter, r *http.Request) {
	if !(handlers.AuthorizationHandler{}).IsAuthorized(rw, r) {
		return
	}
	var model models.Users
	err := json.NewDecoder(r.Body).Decode(&model)
	if err != nil {
		fmt.Print("error in decoding")
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	db := data.DBContext{}.Get(fmt.Sprintf("%s%s-%s", UserPrefix, model.UserName, model.UserId))
	if db == nil {
		http.Error(rw, "User Not found", http.StatusNotFound)
		return
	}

	data.DBContext{}.Set(fmt.Sprintf("%s%s-%s", UserPrefix, model.UserName, model.UserId), model)

	rw.WriteHeader(http.StatusOK)
}

func (c UsersController) ChangePassword(rw http.ResponseWriter, r *http.Request) {
	if !(handlers.AuthorizationHandler{}).IsAuthenticated(rw, r) {
		return
	}
	var model models.Users
	err := json.NewDecoder(r.Body).Decode(&model)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	if check, len := validation(model); len > 0 {
		http.Error(rw, check, http.StatusBadRequest)
		return
	}

	db := data.DBContext{}.Get(fmt.Sprintf("%s%s-%s", UserPrefix, model.UserName, model.UserId))
	if db == nil {
		http.Error(rw, "User Not found", http.StatusNotFound)
		return
	}
	model.Password, err = crypto.HashPassword(model.Password)

	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	data.DBContext{}.Set(fmt.Sprintf("%s%s-%s", UserPrefix, model.UserName, model.UserId), model)
	rw.WriteHeader(http.StatusOK)
}

func validation(model models.Users) (string, int) {
	var validations []string
	if model.UserName == "" {
		validations = append(validations, "User Name is required .")
	}

	if model.Password == "" {
		validations = append(validations, "Password is required .")
	}

	res, _ := json.Marshal(validations)
	return string(res), len(validations)
}

func (c UsersController) userIsExists(userName string) bool {
	db := data.DBContext{}.GetRangePrefixedOfType(fmt.Sprintf("%s%s-", UserPrefix, userName))
	return len(db) > 0
}

func (c UsersController) CreateDefaultUser(model models.Users) error {
	uuid := uuid.NewV4()
	if check, len := validation(model); len > 0 {
		return fmt.Errorf("error in validation %s", check)
	}
	hash, err := crypto.HashPassword(model.Password)
	if err != nil {
		return err
	}
	model.Password = hash
	model.UserId = uuid.String()
	if c.userIsExists(model.UserName) {
		return fmt.Errorf("%s already exists", model.UserName)
	}
	data.DBContext{}.Set(fmt.Sprintf("%s%s-%s", UserPrefix, model.UserName, model.UserId), model)
	return nil
}

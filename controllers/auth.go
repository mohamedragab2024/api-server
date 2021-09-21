package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/kube-carbonara/api-server/crypto"
	data "github.com/kube-carbonara/api-server/data"
	"github.com/kube-carbonara/api-server/models"
)

type AuthController struct{}

func (c AuthController) Login(rw http.ResponseWriter, r *http.Request) {
	var model models.AuthModel
	var user = models.Users{}
	err := json.NewDecoder(r.Body).Decode(&model)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	dbResult := data.DBContext{}.GetRangePrefixedOfType(fmt.Sprintf("%s%s-", UserPrefix, model.UserName))
	if len(dbResult) == 0 {
		http.Error(rw, "Invalid user name or password ", http.StatusForbidden)
		return
	}

	json.Unmarshal(dbResult[0], &user)

	if user.UserId == "" || !crypto.CheckPasswordHash(model.Password, user.Password) {
		http.Error(rw, "Invalid user name or password ", http.StatusForbidden)
		return
	}

	token, err := c.createToken(user)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(token)
}

func (c AuthController) createToken(user models.Users) (models.Token, error) {
	var err error
	atClaims := jwt.MapClaims{}
	atClaims["authorized"] = true
	atClaims["userId"] = user.UserId
	atClaims["userName"] = user.UserName
	atClaims["isAdmin"] = user.IsAdmin
	atClaims["exp"] = time.Now().Add(time.Hour * 24).Unix()
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	token, err := at.SignedString([]byte(os.Getenv("ACCESS_SECRET")))
	if err != nil {
		return models.Token{}, err
	}
	return models.Token{
		JwtToken: token,
	}, nil
}

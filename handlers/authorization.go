package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/kube-carbonara/api-server/data"
	"github.com/kube-carbonara/api-server/httpverbs"
	"github.com/kube-carbonara/api-server/models"
	"github.com/kube-carbonara/api-server/utils"
)

type AuthorizationHandler struct{}

func (h AuthorizationHandler) IsAuthorized(rw http.ResponseWriter, req *http.Request) bool {
	if agent := req.Header.Get("x-agent"); agent != "" {
		appKey := req.Header.Get("x-agent-app-key")
		db := data.DBContext{}.GetRangePrefixedOfType(fmt.Sprintf("Cluster-%s-", agent))
		if len(db) > 0 {
			var model models.Clusters
			json.Unmarshal(db[0], &model)
			if model.AppKey == appKey {
				return true
			}
		}
	}
	var IsAuthorized bool
	user, err := h.ExtractTokenMetadata(req)
	if err != nil {
		IsAuthorized = false
	}

	if h.isRequestToClient(req) {
		IsAuthorized = h.handleClusterAuthorization(req, user)
	}

	if !IsAuthorized && !user.IsAdmin {
		rw.WriteHeader(http.StatusForbidden)
		rw.Write([]byte("Unauthorized access ."))
	}

	return IsAuthorized || user.IsAdmin
}

func (h AuthorizationHandler) IsAuthenticated(rw http.ResponseWriter, req *http.Request) bool {
	_, err := h.ExtractTokenMetadata(req)
	if err != nil {
		rw.WriteHeader(http.StatusForbidden)
		rw.Write([]byte("Unauthorized access ."))
	}
	return err == nil
}

func (h AuthorizationHandler) isRequestToClient(req *http.Request) bool {

	return strings.Contains(req.URL.Path, "/connections")
}

func (h AuthorizationHandler) handleClusterAuthorization(req *http.Request, user models.Users) bool {
	vars := mux.Vars(req)
	clientKey := vars["id"]
	action := req.Method
	var model = models.Clusters{}
	db := data.DBContext{}.GetRangePrefixedOfType(fmt.Sprintf("%s%s-", "Cluster-", clientKey))

	if len(db) > 0 {
		err := json.Unmarshal(db[0], &model)
		if err == nil {
			permissionRecord, err := h.getUserPermission(user, model)
			return err == nil && h.hasAcess(user, permissionRecord, action)
		}
	}
	return user.IsAdmin

}

func (h AuthorizationHandler) getUserPermission(user models.Users, cluster models.Clusters) (models.ClusterPermissons, error) {
	var permissionRecord models.ClusterPermissons
	var exsisted bool
	for _, v := range cluster.UsersAcl {
		if v.UserOrGroupIdenetity == user.UserId {
			exsisted = true
			permissionRecord = v
			break
		}
	}
	if !exsisted {
		return models.ClusterPermissons{}, errors.New("user not authorized")
	}

	return permissionRecord, nil
}

func (h AuthorizationHandler) hasAcess(user models.Users, permissionRecord models.ClusterPermissons, action string) bool {

	return action == httpverbs.MethodGet ||
		utils.Contains(permissionRecord.Permissons, models.ClusterManager) ||
		utils.Contains(permissionRecord.Permissons, models.ReadWrite) || user.IsAdmin

}

func (h AuthorizationHandler) verifyToken(r *http.Request) (*jwt.Token, error) {
	tokenString := h.extractToken(r)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			fmt.Printf("unexpected signing method: %v", token.Header["alg"])
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("ACCESS_SECRET")), nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (h AuthorizationHandler) extractToken(r *http.Request) string {
	bearToken := r.Header.Get("Authorization")
	strArr := strings.Split(bearToken, " ")
	if len(strArr) == 2 {
		return strArr[1]
	}
	return ""
}

func (h AuthorizationHandler) TokenValid(r *http.Request) (*jwt.Token, error) {
	token, err := h.verifyToken(r)
	if err != nil {
		return nil, err
	}
	if _, ok := token.Claims.(jwt.Claims); !ok && !token.Valid {
		return nil, err
	}
	return token, nil
}

func (h AuthorizationHandler) ExtractTokenMetadata(r *http.Request) (models.Users, error) {
	token, err := h.TokenValid(r)
	if err != nil {
		return models.Users{}, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return models.Users{}, err
	}
	var user models.Users
	payload, err := json.Marshal(claims)
	if err == nil {
		json.Unmarshal(payload, &user)
		return user, nil
	}
	return models.Users{}, err
}

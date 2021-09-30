package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/kube-carbonara/api-server/crypto"
	data "github.com/kube-carbonara/api-server/data"
	"github.com/kube-carbonara/api-server/handlers"
	"github.com/kube-carbonara/api-server/models"
	"github.com/kube-carbonara/api-server/utils"
	uuid "github.com/satori/go.uuid"
)

const (
	ClusterPrefix = "Cluster-"
)

type ClusterController struct{}

func (c ClusterController) GetAll(rw http.ResponseWriter, r *http.Request) {
	authHandler := handlers.AuthorizationHandler{}
	if !authHandler.IsAuthenticated(rw, r) {
		return
	}

	user, _ := authHandler.CurrentUser(rw, r)
	result := c.GetListByUser(user)

	aggregation := c.calculateAggregation(result)
	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(models.ClustersResult{
		Data:        result,
		Aggregation: aggregation,
	})

}

func (c ClusterController) GetOne(rw http.ResponseWriter, r *http.Request) {
	authHandler := handlers.AuthorizationHandler{}
	config := utils.NewConfig()
	if !authHandler.IsAuthenticated(rw, r) {
		return
	}
	id := mux.Vars(r)["id"]
	var result = []models.Clusters{}
	db := data.DBContext{}.GetRangePrefixedOfType(fmt.Sprintf("%s%s-", ClusterPrefix, id))
	if db == nil {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	currentUser, _ := authHandler.CurrentUser(rw, r)

	for _, v := range db {
		var model models.Clusters
		json.Unmarshal(v, &model)
		if c.canRead(currentUser, model) {
			model.RegisterScript = fmt.Sprintf("kubectl  apply -f %s/clusters/%s/Config", config.ServerUrl, model.Name)
			result = append(result, model)
		}

	}
	aggregation := c.calculateAggregation(result)
	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(models.ClustersResult{
		Data:        result,
		Aggregation: aggregation,
	})
}

func (c ClusterController) Delete(rw http.ResponseWriter, r *http.Request) {
	if !(handlers.AuthorizationHandler{}).IsAuthorized(rw, r) {
		return
	}
	id := mux.Vars(r)["id"]
	var model = models.Clusters{}
	db := data.DBContext{}.GetRangePrefixedOfType(fmt.Sprintf("%s%s-", ClusterPrefix, id))
	if len(db) == 0 {
		rw.WriteHeader(http.StatusNotFound)
		return
	}
	json.Unmarshal(db[0], &model)
	data.DBContext{}.Delete(fmt.Sprintf("%s%s-%s", ClusterPrefix, id, model.Id))
	rw.WriteHeader(http.StatusNoContent)
}

func (c ClusterController) UpdateMetrics(rw http.ResponseWriter, r *http.Request) {
	if !(handlers.AuthorizationHandler{}).IsAuthorized(rw, r) {
		return
	}
	id := mux.Vars(r)["id"]
	var model = models.Clusters{}
	db := data.DBContext{}.GetRangePrefixedOfType(fmt.Sprintf("%s%s-", ClusterPrefix, id))
	if len(db) == 0 {
		rw.WriteHeader(http.StatusNotFound)
		return
	}
	json.Unmarshal(db[0], &model)
	var metrics models.ClusterMetricsCache
	err := json.NewDecoder(r.Body).Decode(&metrics)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	model.MetricsCache = metrics
	data.DBContext{}.Set(fmt.Sprintf("%s%s-%s", ClusterPrefix, id, model.Id), model)
	rw.WriteHeader(http.StatusNoContent)
}

func (c ClusterController) Update(rw http.ResponseWriter, r *http.Request) {
	if !(handlers.AuthorizationHandler{}).IsAuthorized(rw, r) {
		return
	}
	id := mux.Vars(r)["id"]
	var cluster = models.Clusters{}
	db := data.DBContext{}.GetRangePrefixedOfType(fmt.Sprintf("%s%s-", ClusterPrefix, id))
	if len(db) == 0 {
		rw.WriteHeader(http.StatusNotFound)
		return
	}
	json.Unmarshal(db[0], &cluster)
	var model models.Clusters
	err := json.NewDecoder(r.Body).Decode(&model)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	cluster.UsersAcl = []models.ClusterPermissons{}
	cluster.UsersAcl = append(cluster.UsersAcl, model.UsersAcl...)
	data.DBContext{}.Set(fmt.Sprintf("%s%s-%s", ClusterPrefix, id, cluster.Id), cluster)
	rw.WriteHeader(http.StatusNoContent)
}

func (c ClusterController) Create(rw http.ResponseWriter, r *http.Request) {
	if !(handlers.AuthorizationHandler{}).IsAuthorized(rw, r) {
		return
	}
	uuid := uuid.NewV4()
	appkey, _ := crypto.GenerateBase64ID(10)
	var model models.Clusters
	err := json.NewDecoder(r.Body).Decode(&model)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	if model.Name == "" {
		http.Error(rw, "Name is required .", http.StatusBadRequest)
		return
	}

	if c.clustertIsExists(model.Name) {
		http.Error(rw, fmt.Sprintf("%s already exists", model.Name), http.StatusConflict)
		return
	}
	model.Id = uuid.String()
	model.AppKey = fmt.Sprintf("%s-%s", model.Name, appkey)
	data.DBContext{}.Set(fmt.Sprintf("%s%s-%s", ClusterPrefix, model.Name, model.Id), model)
	rw.WriteHeader(http.StatusCreated)
}

func (c ClusterController) ConfigFile(rw http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var model = models.Clusters{}
	obj := data.DBContext{}.GetRangePrefixedOfType(fmt.Sprintf("%s%s-", ClusterPrefix, id))[0]
	if obj == nil {
		rw.WriteHeader(http.StatusNotFound)
		return
	}
	json.Unmarshal(obj, &model)
	db := data.DBContext{}.Get(fmt.Sprintf("%s%s-%s", ClusterPrefix, model.Name, model.Id))
	if db == nil {
		rw.WriteHeader(http.StatusNotFound)
		return
	}
	reader, err := os.ReadFile("asstes/agent-deployment.yaml")
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	content := string(reader)
	content = strings.ReplaceAll(content, "{SERVER_ADDRESS}", r.Host)
	content = strings.ReplaceAll(content, "{CLIENT_ID}", model.Name)
	content = strings.ReplaceAll(content, "{APP_KEY}", model.AppKey)
	content = strings.ReplaceAll(content, "{REMOTE_SCHEMA}", "http")
	b := bytes.NewBuffer([]byte(content))
	rw.Header().Set("Content-Type", r.Header.Get("application/x-yaml"))
	n, _ := b.WriteTo(rw)
	fmt.Print(n)
}

func (c ClusterController) clustertIsExists(name string) bool {

	return len(data.DBContext{}.GetRangePrefixedOfType(fmt.Sprintf("%s%s-", ClusterPrefix, name))) > 0

}

func (c ClusterController) calculateAggregation(clusters []models.Clusters) models.ClusterAggregation {
	aggregation := models.ClusterAggregation{}
	for _, v := range clusters {
		aggregation.TotalCount++
		aggregation.TotalCpu += v.MetricsCache.TotalCpuCores
		aggregation.TotalCpuUsage += v.MetricsCache.TotalCpuUsage
		aggregation.TotalMemory += v.MetricsCache.TotalMemory
		aggregation.TotalMemoryUsage += v.MetricsCache.TotalMemoryUsage
		aggregation.TotalNodes += v.MetricsCache.NodesCount
	}

	if aggregation.TotalCpuUsage > 0 && aggregation.TotalCpu > 0 {
		aggregation.CpuPercentage = fmt.Sprintf("%v%%", aggregation.TotalCpuUsage*100/aggregation.TotalCpu)
	}

	if aggregation.TotalMemoryUsage > 0 && aggregation.TotalMemory > 0 {
		aggregation.MemoryPercentage = fmt.Sprintf("%v%%", aggregation.TotalMemoryUsage*100/aggregation.TotalMemory)
	}

	return aggregation
}

func (c ClusterController) GetList() []models.Clusters {
	config := utils.NewConfig()
	result := []models.Clusters{}
	db := data.DBContext{}.GetRangePrefixedOfType(ClusterPrefix)
	for _, v := range db {
		var model models.Clusters
		json.Unmarshal(v, &model)
		model.RegisterScript = fmt.Sprintf("kubectl  apply -f %s/clusters/%s/Config", config.ServerUrl, model.Name)
		result = append(result, model)
	}

	return result
}

func (c ClusterController) GetListByUser(user models.Users) []models.Clusters {
	config := utils.NewConfig()
	result := []models.Clusters{}
	db := data.DBContext{}.GetRangePrefixedOfType(ClusterPrefix)
	for _, v := range db {
		var model models.Clusters
		json.Unmarshal(v, &model)
		if c.canRead(user, model) {
			model.RegisterScript = fmt.Sprintf("kubectl  apply -f %s/clusters/%s/Config", config.ServerUrl, model.Name)
			result = append(result, model)
		}

	}

	return result
}

func (c ClusterController) GetResultList() models.ClustersResult {
	data := c.GetList()
	return models.ClustersResult{
		Data:        data,
		Aggregation: c.calculateAggregation(data),
	}
}

func (c ClusterController) GetById(id string) (models.Clusters, error) {
	config := utils.NewConfig()
	var model = models.Clusters{}
	db := data.DBContext{}.GetRangePrefixedOfType(fmt.Sprintf("%s%s-", ClusterPrefix, id))
	if len(db) == 0 {
		return model, fmt.Errorf("cluster not found")
	}
	json.Unmarshal(db[0], &model)
	model.RegisterScript = fmt.Sprintf("kubectl  apply -f %s/clusters/%s/Config", config.ServerUrl, model.Name)
	return model, nil
}

func (c ClusterController) SaveChanges(model models.Clusters) error {
	err := data.DBContext{}.Set(fmt.Sprintf("%s%s-%s", ClusterPrefix, model.Name, model.Id), model)
	return err
}

func (c ClusterController) canRead(user models.Users, cluster models.Clusters) bool {
	canRead := false
	for _, v := range cluster.UsersAcl {
		if v.UserOrGroupIdenetity == user.UserId {
			canRead = true
		}
	}
	return canRead
}

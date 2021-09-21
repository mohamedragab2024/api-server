package models

type Clusters struct {
	Id              string              `json:"id"`
	Name            string              `json:"name"`
	UsersAcl        []ClusterPermissons `json:"users"`
	MetricsCache    ClusterMetricsCache `json:"metrics"`
	AppKey          string              `json:"appKey"`
	IsConnected     bool                `json:"isConnected"`
	LastSyncMessage string              `json:"lastSyncMessage"`
}

type ClusterPermissons struct {
	UserOrGroupIdenetity string   `json:"userId"`
	Permissons           []string `json:"permissons"`
}

type ClusterMetricsCache struct {
	TotalCpuCores    int64  `json:"totalCpuCores"`
	TotalCpuUsage    int64  `json:"totalCpuUsage"`
	TotalMemory      int64  `json:"totalMemory"`
	TotalMemoryUsage int64  `json:"totalMemoryUsage"`
	CpuPercentage    string `json:"cpuPercentage"`
	MemoryPercentage string `json:"memoryPercentage"`
	NodesCount       int64  `json:"nodesCount"`
	Provider         string `json:"provider"`
}

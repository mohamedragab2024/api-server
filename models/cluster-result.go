package models

type ClustersResult struct {
	Data        []Clusters         `json:"data"`
	Aggregation ClusterAggregation `json:"aggregation"`
}

type ClusterAggregation struct {
	TotalCount       int64  `json:"totalCount"`
	TotalNodes       int64  `json:"totalNodes"`
	TotalCpu         int64  `json:"totalCpu"`
	TotalCpuUsage    int64  `json:"totalCpuUsage"`
	TotalMemory      int64  `json:"totalMemory"`
	TotalMemoryUsage int64  `json:"totalMemoryUsage"`
	CpuPercentage    string `json:"cpuPercentage"`
	MemoryPercentage string `json:"memoryPercentage"`
}

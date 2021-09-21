package models

const (
	ReadOnly       = "ReadOnly"       // read only the resources on the cluster
	ReadWrite      = "ReadWrite"      // can read and create any resources under namespaces
	ClusterManager = "ClusterManager" // can create new namespace and maange all resources
)

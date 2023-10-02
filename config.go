package main

const (
	local = "local"
	cloud = "server"
)

var envMap = map[string]struct{}{}

type config struct {
	IsMainNode bool `envconfig:"main_node"`
	// Replicas is the list of child nodes for data replication
	Replicas []string `envconfig:"replicas"`
}

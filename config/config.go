package config

type Config struct {
	IsMainNode bool `envconfig:"main_node"`
	// Replicas is the list of child nodes for data replication
	Replicas []string `envconfig:"replicas"`
}

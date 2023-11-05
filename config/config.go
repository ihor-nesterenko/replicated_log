package config

type Config struct {
	IsMainNode bool `envconfig:"main_node"`
	// Replicas is the list of child nodes for data replication
	Replicas []string `envconfig:"replicas"`
	// Delay is used to artificially slow down replication process (in seconds)
	Delay int `envconfig:"delay"`
}

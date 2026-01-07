package config

type Config struct {
	Clusters       []ClusterConfig `yaml:"clusters"`
	CurrentContext string          `yaml:"currentContext,omitempty"` // Ensure this is included
}

type ClusterConfig struct {
	Name           string  `yaml:"name"`
	Cluster        Cluster `yaml:"cluster"`
	CurrentContext string  `yaml:"currentContext,omitempty"` // Ensure this is included
}

type Cluster struct {
	Server string `yaml:"server"`
	Auth   Auth   `yaml:"auth"`
}

type Auth struct {
	BaseAuth    string `yaml:"baseAuth,omitempty"`
	BearerToken string `yaml:"bearerToken,omitempty"`
	SSOToken    string `yaml:"ssoToken,omitempty"`
}

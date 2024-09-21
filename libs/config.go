package libs

import (
	"log"
	"os"
	"gopkg.in/yaml.v2"
)

// Define nested structures for the config
type ImageDetails struct {
	ImageProject string `yaml:"image_project"`
	ImageFamily  string `yaml:"image_family"`
}

type IpDetails struct {
	IpV6 bool `yaml:"ipv6"`
	Name string `yaml:"name"`
}

type SSHDetails struct {
	Username  string `yaml:"username"`
	PublicKey string `yaml:"publickey"`
}

type Config struct {
	ProjectID    string       `yaml:"project_id"`
	Zone         string       `yaml:"zone"`
	InstanceName string       `yaml:"instance_name"`
	MachineType  string       `yaml:"machine_type"`
	ImageDetails ImageDetails `yaml:"image_deatails"`
	Credentials  string       `yaml:"credentials_file"`
	SSHDetails   SSHDetails   `yaml:"ssh_details"`
	IpDetails   IpDetails   `yaml:"ip_details"`
}


// LoadConfig loads the configuration from the specified YAML file
func LoadConfig(filePath string) (*Config, error) {
	configData, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(configData, &config); err != nil {
		log.Fatalf("Failed to parse config file: %v", err)
		return nil, err
	}

	return &config, nil
}
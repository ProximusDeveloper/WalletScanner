package setup

import (
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Version     string      `yaml:"version"`
	RpcProvider RpcProvider `yaml:"rpc_provider"`
	Chains      []Chain     `yaml:"chains"`
	Logging     Logging     `yaml:"logging"`
	Entropy     int         `yaml:"entropy"`
}

type RpcProvider struct {
	Name        string `yaml:"name"`
	URL         string `yaml:"url"`
	ApiKeysPath string `yaml:"api_keys"`
	RateLimit   int    `yaml:"rate_limit"`
	UsagePeriod int    `yaml:"usage_period"`
	BatchSize   int    `yaml:"batch_size"`
}

type Chain struct {
	Name     string `yaml:"name"`
	ChainID  int    `yaml:"chain_id"`
	Endpoint string `yaml:"endpoint"`
}

type Logging struct {
	SaveEmpty bool   `yaml:"save_empty"`
	Empty     string `yaml:"empty"`
	Success   string `yaml:"success"`
}

func MustLoad(path string) *Config {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Fatalf("file %s does not exist", path)
	}

	bytes, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	var config Config
	err = yaml.Unmarshal(bytes, &config)
	if err != nil {
		log.Fatalf("cannot unmarshal config: %s", err)
	}

	return &config
}

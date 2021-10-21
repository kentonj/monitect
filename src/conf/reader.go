package conf

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

func NewConfig(configPath string) *Config {
	// read in config
	config := new(Config)
	file, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Fatalf("unable to read config file %s", configPath)
	}
	err = yaml.Unmarshal(file, config)
	if err != nil {
		log.Fatalf("unable to unmarshal config file %s with error %v", configPath, err)
	}
	return config
}

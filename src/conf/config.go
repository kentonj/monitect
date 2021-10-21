package conf

type Config struct {
	Mongo struct {
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		DB       string `yaml:"db"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"mongo"`
}

package conf

type Config struct {
	Database struct {
		File    string `yaml:"file"`
		InitSQL string `yaml:"init_sql"`
		Debug   bool   `yaml:"debug"`
	} `yaml:"database"`
	Server struct {
		Port string `yaml:"port"`
		Host string `yaml:"host"`
	} `yaml:"server"`
}

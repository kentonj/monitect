package conf

type Config struct {
	Database struct {
		File    string `yaml:"file"`
		InitSQL string `yaml:"init_sql"`
		Debug   bool   `yaml:"debug"`
	} `yaml:"database"`
}

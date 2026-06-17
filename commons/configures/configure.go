package configures

import (
	"os"

	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	Port         int `yaml:"port"`
	CallbackPort int `yaml:"callbackPort"`

	Log struct {
		LogPath string `yaml:"logPath"`
		LogName string `yaml:"logName"`
	} `yaml:"log"`

	Mysql struct {
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Address  string `yaml:"address"`
		DbName   string `yaml:"name"`
		Debug    bool   `yaml:"debug"`
	} `yaml:"mysql"`

	ImApiDomain string `yaml:"imApiDomain"`

	CallbackBaseUrl   string `yaml:"callbackBaseUrl"`
	AiBotCallbackUrl  string `yaml:"aiBotCallbackUrl"`
	AssistantAgentUrl string `yaml:"assistantAgentUrl"`

	BotConnector struct {
		Domain string `yaml:"domain"`
	} `yaml:"botConnector"`
}

var Config AppConfig
var Env string

const (
	EnvDev  = "dev"
	EnvProd = "prod"
)

func InitConfigures() error {
	cfBytes, err := os.ReadFile("conf/config.yml")
	if err == nil {
		var conf AppConfig
		yaml.Unmarshal(cfBytes, &conf)
		Config = conf
		if Config.Port <= 0 {
			Config.Port = 8070
		}
		return nil
	} else {
		return err
	}
}

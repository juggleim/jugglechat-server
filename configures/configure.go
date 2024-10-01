package configures

import (
	"os"

	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	Port int `yaml:"port"`
	Log  struct {
		LogPath string `yaml:"logPath"`
		LogName string `yaml:"logName"`
	} `ymal:"log"`
	Mysql struct {
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Address  string `yaml:"address"`
		DbName   string `yaml:"name"`
	} `yaml:"mysql"`
	Qiniu struct {
		AccessKey string `yaml:"accessKey"`
		SecretKey string `yaml:"secretKey"`
		Bucket    string `yaml:"bucket"`
		Domain    string `yaml:"domain"`
	} `yaml:"qiniu"`
	BaiduSms struct {
		ApiKey    string `yaml:"apiKey"`
		SecretKey string `yaml:"secretKey"`
	} `yaml:"baidusms"`
	Im struct {
		AppKey    string `yaml:"appKey"`
		AppSecret string `yaml:"appSecret"`
		ApiUrl    string `yaml:"apiUrl"`
	} `yaml:"im"`
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
		return nil
	} else {
		return err
	}
}

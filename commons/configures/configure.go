package configures

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	PasswordSMSPepper string `yaml:"password_sms_pepper"`
	// New explicit setting takes precedence; nil uses the deprecated alias.
	DeviceIMKickEnabled         *bool            `yaml:"device_im_kick_enabled"`
	DeviceIMBanEnabled          bool             `yaml:"device_im_ban_enabled"`
	AssistantServiceCredentials bool             `yaml:"assistant_service_credentials"`
	DeviceAuthRequiredAt        map[string]int64 `yaml:"device_auth_required_at"`
	Port                        int              `yaml:"port"`
	CallbackPort                int              `yaml:"callbackPort"`

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

func (c AppConfig) DeviceKickEnabled() bool {
	if c.DeviceIMKickEnabled != nil {
		return *c.DeviceIMKickEnabled
	}
	return c.DeviceIMBanEnabled
}

const (
	EnvDev  = "dev"
	EnvProd = "prod"
)

func InitConfigures() error {
	cfBytes, err := os.ReadFile("conf/config.yml")
	if err == nil {
		var conf AppConfig
		if err := yaml.Unmarshal(cfBytes, &conf); err != nil {
			return err
		}
		for _, at := range conf.DeviceAuthRequiredAt {
			if at < 0 {
				return fmt.Errorf("device_auth_required_at must be Unix milliseconds or zero")
			}
		}
		Config = conf
		if Config.Port <= 0 {
			Config.Port = 8070
		}
		return nil
	} else {
		return err
	}
}

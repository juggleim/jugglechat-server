package models

import (
	"github.com/juggleim/jugglechat-server/services"
)

type EmailConf struct {
	AppKey string                   `json:"app_key"`
	Conf   *services.MailEngineConf `json:"conf"`
}

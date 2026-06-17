package events

import "github.com/juggleim/jugglechat-server/storages/models"

type UserRegisteEvent func(appkey string, user models.User)

var userRegisteEvents []UserRegisteEvent

func init() {
	userRegisteEvents = []UserRegisteEvent{}
}

func RegisteUserRegisteEvent(event UserRegisteEvent) {
	userRegisteEvents = append(userRegisteEvents, event)
}

func TriggerUserRegiste(appkey string, user models.User) {
	for _, event := range userRegisteEvents {
		event(appkey, user)
	}
}

package deviceauth

import "github.com/juggleim/jugglechat-server/commons/configures"

// Zero means compatibility remains enabled. Equality starts strict mode.
func Required(app string, now int64) bool {
	at := configures.Config.DeviceAuthRequiredAt[app]
	return at > 0 && now >= at
}

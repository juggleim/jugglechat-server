package configures

import (
	"gopkg.in/yaml.v3"
	"testing"
)

func TestDeviceKickConfigurationMigration(t *testing.T) {
	for _, tc := range []struct {
		yaml string
		want bool
	}{
		{"{}", false},
		{"device_im_ban_enabled: true", true},
		{"device_im_ban_enabled: true\ndevice_im_kick_enabled: false", false},
		{"device_im_ban_enabled: false\ndevice_im_kick_enabled: true", true},
	} {
		var c AppConfig
		if e := yaml.Unmarshal([]byte(tc.yaml), &c); e != nil {
			t.Fatal(e)
		}
		if got := c.DeviceKickEnabled(); got != tc.want {
			t.Errorf("%s got=%v", tc.yaml, got)
		}
	}
}

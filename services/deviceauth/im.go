package deviceauth

import (
	"context"
	"sort"
)

// IMAdapter reconciles current device offline state before retrying the same
// target session under Store's device lock. It never bans or unbans credentials.
type IMAdapter interface {
	Online(context.Context, string, string) (map[string]bool, error)
	Reconcile(context.Context, RevokeJob) (bool, error)
	Disconnect(context.Context, RevokeJob) error
}

// UnsupportedIM is the explicit integration placeholder. Never replace it with
// user-level online status, an empty online map, or a successful no-op kick.
type UnsupportedIM struct{}

func (UnsupportedIM) Online(context.Context, string, string) (map[string]bool, error) {
	return nil, ErrIM
}
func (UnsupportedIM) Reconcile(context.Context, RevokeJob) (bool, error) { return false, ErrIM }
func (UnsupportedIM) Disconnect(context.Context, RevokeJob) error        { return ErrIM }

type DeviceView struct {
	Device
	IsCurrent bool `json:"is_current_device"`
}
type Groups struct {
	Online  []DeviceView `json:"online_devices"`
	Offline []DeviceView `json:"offline_devices"`
}

func Group(ds []Device, online map[string]bool, app, user, current string) Groups {
	g := Groups{Online: []DeviceView{}, Offline: []DeviceView{}}
	sort.Slice(ds, func(i, j int) bool {
		if ds[i].LastLoginTime == ds[j].LastLoginTime {
			return ds[i].DeviceID < ds[j].DeviceID
		}
		return ds[i].LastLoginTime > ds[j].LastLoginTime
	})
	for _, d := range ds {
		if d.AppKey != app || d.UserID != user || d.State != "active" || d.DeletedAt != nil {
			continue
		}
		v := DeviceView{d, d.DeviceID == current}
		if online[d.DeviceID] {
			g.Online = append(g.Online, v)
		} else {
			g.Offline = append(g.Offline, v)
		}
	}
	return g
}

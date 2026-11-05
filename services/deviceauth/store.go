package deviceauth

import (
	"context"
	"errors"
	"time"

	mysqlerr "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type RevokeJob struct {
	OperationID     string `gorm:"column:operation_id;primaryKey"`
	AppKey          string
	UserID          string `gorm:"column:user_id"`
	DeviceID        string `gorm:"column:device_id"`
	TargetSessionID string `gorm:"column:target_session_id"`
	BannedUntil     int64  // Deprecated history; zero on new jobs, never a deadline.
	Done            bool
	Attempts        int
	NextAttempt     int64
	LastError       string
}

func (RevokeJob) TableName() string { return "device_revoke_jobs" }

type Store struct {
	DB *gorm.DB
	IM IMAdapter
}

func (s Store) adapter() IMAdapter {
	if s.IM == nil {
		return UnsupportedIM{}
	}
	return s.IM
}
func (s Store) tx(ctx context.Context, fn func(*gorm.DB) error) error {
	var err error
	for i := 0; i < 3; i++ {
		err = s.DB.WithContext(ctx).Transaction(fn)
		var me *mysqlerr.MySQLError
		if !errors.As(err, &me) || (me.Number != 1062 && me.Number != 1213 && me.Number != 1205) {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(i+1) * 10 * time.Millisecond):
		}
	}
	return err
}
func deviceQuery(tx *gorm.DB, app, user, id string) *gorm.DB {
	return tx.Where("app_key=? AND user_id=? AND device_id=?", app, user, id)
}
func issueTx(tx *gorm.DB, app, user string, p DeviceParams, now int64) (Device, error) {
	id, platform := p.Identity()
	var d Device
	err := deviceQuery(tx.Clauses(clause.Locking{Strength: "UPDATE"}), app, user, id).Take(&d).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return d, err
	}
	if err == nil {
		if err = d.CanLogin(now); err != nil {
			return d, err
		}
	} else {
		d = Device{AppKey: app, UserID: user, DeviceID: id, CreatedTime: now}
	}
	session, err := RandomID()
	if err != nil {
		return d, err
	}
	d.SessionID = session
	d.DeviceName = p.DeviceName
	d.Platform = platform
	d.OSVersion = p.OSVersion
	d.AppVersion = p.AppVersion
	d.LastLoginTime = now
	d.LastActiveTime = nil
	d.UpdatedTime = now
	d.State = "active"
	d.DeletedAt = nil
	d.RevokeOperationID = ""
	return d, tx.Save(&d).Error
}
func (s Store) Issue(ctx context.Context, app, user string, p DeviceParams, key []byte, now int64) (string, error) {
	if legacy, e := p.Validate(Required(app, now)); e != nil {
		return "", e
	} else if legacy {
		return "", ErrUpgrade
	}
	var token string
	err := s.tx(ctx, func(tx *gorm.DB) error {
		d, e := issueTx(tx, app, user, p, now)
		if e != nil {
			return e
		}
		token, e = Sign(Claims{app, user, d.DeviceID, d.SessionID, "user_session", now}, key)
		return e
	})
	return token, err
}
func (s Store) Authenticate(ctx context.Context, c Claims, now int64) error {
	return s.tx(ctx, func(tx *gorm.DB) error {
		var d Device
		// Share the device lock with deletion/session issuance. The validated
		// session cannot change between this read and the activity write/commit.
		if e := deviceQuery(tx.Clauses(clause.Locking{Strength: "UPDATE"}), c.AppKey, c.UserID, c.DeviceID).Take(&d).Error; e != nil {
			if errors.Is(e, gorm.ErrRecordNotFound) {
				return ErrAuth
			}
			return e
		}
		if e := d.Authenticate(c, now); e != nil {
			return e
		}
		// Throttling is not authentication: a valid session remains valid when
		// no activity write is due (including same-millisecond requests).
		if d.LastActiveTime != nil && *d.LastActiveTime >= now-60000 {
			return nil
		}
		return tx.Model(&d).Update("last_active_time", now).Error
	})
}
func (d *Device) BeginRevoke(op string, now int64) (RevokeJob, error) {
	if d.State != "active" && d.State != "revoking" && d.State != "revoked" {
		return RevokeJob{}, ErrAuth
	}
	if d.State == "active" {
		d.State = "revoking"
		d.DeletedAt = &now
		d.RevokeOperationID = op
		d.UpdatedTime = now
	}
	return RevokeJob{OperationID: d.RevokeOperationID, AppKey: d.AppKey, UserID: d.UserID, DeviceID: d.DeviceID, TargetSessionID: d.SessionID, Done: d.State == "revoked", NextAttempt: now}, nil
}
func (s Store) Delete(ctx context.Context, c Claims, id string, now int64) (RevokeJob, error) {
	if !validDeviceID(id) {
		return RevokeJob{}, ErrParams
	}
	if id == c.DeviceID {
		return RevokeJob{}, ErrCurrent
	}
	var j RevokeJob
	err := s.tx(ctx, func(tx *gorm.DB) error {
		var d Device
		if e := deviceQuery(tx.Clauses(clause.Locking{Strength: "UPDATE"}), c.AppKey, c.UserID, id).Take(&d).Error; e != nil {
			if errors.Is(e, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return e
		}
		// Compare the locked row too: database equality may have resolved a
		// different spelling to the current device (including pre-fix rows).
		if d.DeviceID == c.DeviceID {
			return ErrCurrent
		}
		if d.State != "active" {
			return tx.Where("operation_id=?", d.RevokeOperationID).Take(&j).Error
		}
		op, e := RandomID()
		if e != nil {
			return e
		}
		j, e = d.BeginRevoke(op, now)
		if e != nil {
			return e
		}
		if e = tx.Save(&d).Error; e != nil {
			return e
		}
		return tx.Create(&j).Error
	})
	if err != nil {
		return j, err
	}
	if j.Done {
		return j, nil
	}
	err = s.Process(ctx, j, now)
	if err == nil {
		err = s.DB.WithContext(ctx).Where("operation_id=?", j.OperationID).Take(&j).Error
	}
	return j, err
}
func RunIM(ctx context.Context, a IMAdapter, j RevokeJob) error {
	done, e := a.Reconcile(ctx, j)
	if e != nil {
		return e
	}
	if done {
		return nil
	}
	return a.Disconnect(ctx, j)
}
func (s Store) Process(ctx context.Context, j RevokeJob, now int64) error {
	// Read identity from durable storage, never trust a caller's target fields.
	if e := s.DB.WithContext(ctx).Where("operation_id=?", j.OperationID).Take(&j).Error; e != nil {
		return e
	}
	callCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var imErr error
	// No automatic transaction retries around external effects. Lock order is
	// device then job, shared with login/delete. The device lock fences business
	// session issuance across processes through send AND readback, not remote
	// asynchronous execution after timeout/acknowledgement.
	err := s.DB.WithContext(callCtx).Transaction(func(tx *gorm.DB) error {
		var d Device
		e := deviceQuery(tx.Clauses(clause.Locking{Strength: "UPDATE"}), j.AppKey, j.UserID, j.DeviceID).Take(&d).Error
		if e != nil && !errors.Is(e, gorm.ErrRecordNotFound) {
			return e
		}
		var current RevokeJob
		if e := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("operation_id=?", j.OperationID).Take(&current).Error; e != nil {
			return e
		}
		if current.Done {
			return nil
		}
		if errors.Is(e, gorm.ErrRecordNotFound) || d.SessionID != current.TargetSessionID || d.RevokeOperationID != current.OperationID || d.State != "revoking" {
			return tx.Model(&current).Updates(map[string]interface{}{"done": true, "last_error": "superseded"}).Error
		}
		if imErr = RunIM(callCtx, s.adapter(), current); imErr != nil {
			return tx.Model(&current).Updates(map[string]interface{}{"attempts": gorm.Expr("attempts+1"), "next_attempt": now + 60000, "last_error": "IM kick unconfirmed; retry pending"}).Error
		}
		r := deviceQuery(tx.Model(&Device{}), j.AppKey, j.UserID, j.DeviceID).Where("state='revoking' AND revoke_operation_id=? AND session_id=?", j.OperationID, j.TargetSessionID).Updates(map[string]interface{}{"state": "revoked", "updated_time": now})
		if r.Error != nil {
			return r.Error
		}
		if r.RowsAffected != 1 {
			return ErrAuth
		}
		return tx.Model(&RevokeJob{}).Where("operation_id=? AND done=false", j.OperationID).Updates(map[string]interface{}{"done": true, "last_error": ""}).Error
	})
	if err != nil {
		return err
	}
	return imErr
}
func (s Store) List(ctx context.Context, c Claims) (Groups, error) {
	var ds []Device
	err := s.DB.WithContext(ctx).Where("app_key=? AND user_id=? AND state='active' AND deleted_at IS NULL", c.AppKey, c.UserID).Find(&ds).Error
	if err != nil {
		return Groups{}, err
	}
	callCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	online, err := s.adapter().Online(callCtx, c.AppKey, c.UserID)
	if err != nil {
		return Groups{}, err
	}
	if online == nil {
		return Groups{}, ErrIM
	}
	return Group(ds, online, c.AppKey, c.UserID, c.DeviceID), nil
}

// Recover processes a bounded page. Jobs remain durable across crashes/retries.
func (s Store) Recover(ctx context.Context, now int64) error {
	var jobs []RevokeJob
	if e := s.DB.WithContext(ctx).Where("done=false AND next_attempt<=?", now).Order("next_attempt").Limit(20).Find(&jobs).Error; e != nil {
		return e
	}
	for _, j := range jobs {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		_ = s.Process(ctx, j, now)
	}
	return s.DB.WithContext(ctx).Model(&QRRecord{}).Where("created_time<=? AND result_cipher<>''", now-600000).Update("result_cipher", "").Error
}

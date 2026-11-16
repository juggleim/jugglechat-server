package dbs

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"sort"
	"strconv"
	"time"

	mysqlerr "github.com/go-sql-driver/mysql"

	"github.com/google/uuid"
	"github.com/juggleim/jugglechat-server/storages/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ConverScope struct {
	AppKey     string `json:"app_key"`
	ConverID   string `json:"conver_id"`
	ConverType int32  `json:"conver_type"`
	SubChannel string `json:"sub_channel"`
}

func (s ConverScope) ID() string {
	b, _ := json.Marshal(s)
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

type ConverConfigEvent struct {
	EventID     string                 `json:"event_id"`
	Version     int64                  `json:"version"`
	Scope       ConverScope            `json:"scope"`
	ActorID     string                 `json:"actor_id"`
	TargetID    string                 `json:"target_id"`
	Type        *int64                 `json:"type,omitempty"`
	Confs       map[string]bool        `json:"confs,omitempty"`
	RawConfig   map[string]interface{} `json:"raw_config"`
	ChangedKeys []string               `json:"changed_keys"`
}

type ConverOutbox struct {
	EventID       string `gorm:"column:event_id;primaryKey"`
	ScopeID       string `gorm:"column:scope_id"`
	Version       int64
	AppKey        string
	ActorID       string `gorm:"column:actor_id"`
	TargetID      string `gorm:"column:target_id"`
	ConverType    int32
	Payload       string
	Status        string
	Attempts      int64
	NextAttemptAt int64
	LeaseUntil    int64
	ClaimToken    string
	Diagnostic    string
}

func (ConverOutbox) TableName() string { return "converconf_outbox" }

type converScopeVersion struct {
	ScopeID    string `gorm:"column:scope_id;primaryKey"`
	Version    int64
	AppKey     string
	ConverID   string `gorm:"column:conver_id"`
	ConverType int32
	SubChannel string
}

func (converScopeVersion) TableName() string { return "converconf_scope_versions" }

type ConverOutboxStore struct{ DB *gorm.DB }

// ReadSnapshot returns configuration and its server-authenticated version from
// one repeatable-read snapshot, including the pre-outbox version-zero case.
func (s ConverOutboxStore) ReadSnapshot(ctx context.Context, scope ConverScope) ([]ConverConfDao, map[string]interface{}, error) {
	var items []ConverConfDao
	meta := map[string]interface{}{"version": int64(0), "event_id": ""}
	err := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var rows []converScopeVersion
		if e := tx.Where("app_key=? AND conver_id=? AND conver_type=? AND sub_channel=?", scope.AppKey, scope.ConverID, scope.ConverType, scope.SubChannel).Find(&rows).Error; e != nil {
			return e
		}
		if len(rows) > 0 {
			row := rows[0]
			meta["version"] = row.Version
			var jobs []ConverOutbox
			if e := tx.Where("scope_id=? AND version=?", row.ScopeID, row.Version).Find(&jobs).Error; e != nil {
				return e
			}
			if len(jobs) > 0 {
				meta["event_id"] = jobs[0].EventID
			}
		}
		return tx.Where("app_key=? AND conver_id=? AND conver_type=? AND sub_channel=?", scope.AppKey, scope.ConverID, scope.ConverType, scope.SubChannel).Find(&items).Error
	}, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	return items, meta, err
}

// Save serializes the entire snapshot, version increment and outbox insert with
// a SQL-collation-equivalent scope row lock. INSERT ON DUPLICATE KEY acquires the same lock even on first
// use, avoiding an absent-row read/creation race across processes.
func (s ConverOutboxStore) Save(ctx context.Context, actor, target string, items []models.ConverConf) error {
	if s.DB == nil || len(items) == 0 {
		return errors.New("configuration storage unavailable")
	}
	first := items[0]
	scope := ConverScope{first.AppKey, first.ConverId, first.ConverType, first.SubChannel}
	for _, item := range items {
		if (ConverScope{item.AppKey, item.ConverId, item.ConverType, item.SubChannel}) != scope {
			return errors.New("mixed configuration scopes")
		}
	}
	for attempt := 0; ; attempt++ {
		err := s.saveOnce(ctx, actor, target, items, scope)
		var deadlock *mysqlerr.MySQLError
		if !errors.As(err, &deadlock) || deadlock.Number != 1213 || attempt >= 9 {
			return err
		}
		timer := time.NewTimer(time.Duration(attempt+1) * 5 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}

func (s ConverOutboxStore) saveOnce(ctx context.Context, actor, target string, items []models.ConverConf, scope ConverScope) error {
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		id := scope.ID()
		if e := tx.Exec("INSERT INTO converconf_scope_versions(scope_id,app_key,conver_id,conver_type,sub_channel,version) VALUES (?,?,?,?,?,0) ON DUPLICATE KEY UPDATE version=version", id, scope.AppKey, scope.ConverID, scope.ConverType, scope.SubChannel).Error; e != nil {
			return e
		}
		var row converScopeVersion
		if e := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("app_key=? AND conver_id=? AND conver_type=? AND sub_channel=?", scope.AppKey, scope.ConverID, scope.ConverType, scope.SubChannel).Take(&row).Error; e != nil {
			return e
		}
		id = row.ScopeID // candidate hash is never the lock/delivery identity after a unique conflict
		var current []ConverConfDao
		if e := tx.Where("app_key=? AND conver_id=? AND conver_type=? AND sub_channel=?", scope.AppKey, scope.ConverID, scope.ConverType, scope.SubChannel).Find(&current).Error; e != nil {
			return e
		}
		values := map[string]string{"msg_life_time": "0", "forbid_forward": "false", "content_protection": "false", "watermark_enabled": "false"}
		for _, v := range current {
			values[v.ItemKey] = v.ItemValue
		}
		changed := []string{}
		for _, v := range items {
			if values[v.ItemKey] != v.ItemValue {
				changed = append(changed, v.ItemKey)
			}
			values[v.ItemKey] = v.ItemValue
		}
		if len(changed) == 0 {
			return nil
		}
		sort.Strings(changed)
		for _, v := range items {
			if e := tx.Exec("INSERT INTO converconfs(app_key,conver_id,conver_type,sub_channel,item_key,item_value,item_type) VALUES(?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE item_value=VALUES(item_value)", v.AppKey, v.ConverId, v.ConverType, v.SubChannel, v.ItemKey, v.ItemValue, v.ItemType).Error; e != nil {
				return e
			}
		}
		lifetime, e := strconv.ParseInt(values["msg_life_time"], 10, 64)
		if e != nil {
			return errors.New("invalid stored lifetime")
		}
		bools := map[string]bool{}
		raw := map[string]interface{}{"msg_life_time": lifetime}
		for _, key := range []string{"forbid_forward", "content_protection", "watermark_enabled"} {
			b, e := strconv.ParseBool(values[key])
			if e != nil {
				return errors.New("invalid stored config")
			}
			bools[key] = b
			raw[key] = b
		}
		event := ConverConfigEvent{EventID: uuid.NewString(), Version: row.Version + 1, Scope: scope, ActorID: actor, TargetID: target, RawConfig: raw, ChangedKeys: changed}
		// Old decoders default an omitted type to off. Preserve every raw value:
		// existing clients interpret known day menus, otherwise seconds.
		event.Type = &lifetime
		if len(changed) != 1 || changed[0] != "msg_life_time" {
			event.Confs = bools
		}
		payload, e := json.Marshal(event)
		if e != nil {
			return e
		}
		if e = tx.Model(&row).Update("version", event.Version).Error; e != nil {
			return e
		}
		return tx.Create(&ConverOutbox{EventID: event.EventID, ScopeID: id, Version: event.Version, AppKey: scope.AppKey, ActorID: actor, TargetID: target, ConverType: scope.ConverType, Payload: string(payload), Status: "pending"}).Error
	})
}

// Claim never passes an unfinished predecessor in the same scope. SKIP LOCKED
// permits independent scopes to progress concurrently on MySQL 8.0.
func (s ConverOutboxStore) Claim(ctx context.Context, now, lease int64) (*ConverOutbox, error) {
	if s.DB == nil || lease <= 0 {
		return nil, errors.New("outbox unavailable")
	}
	var job ConverOutbox
	e := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		e := tx.Raw(`SELECT o.* FROM converconf_outbox o
   WHERE o.status <> 'sent' AND o.next_attempt_at <= ? AND o.lease_until <= ?
   AND NOT EXISTS (SELECT 1 FROM converconf_outbox p WHERE p.scope_id=o.scope_id AND p.version<o.version AND p.status<>'sent')
   ORDER BY o.next_attempt_at,o.scope_id,o.version LIMIT 1 FOR UPDATE SKIP LOCKED`, now, now).Scan(&job).Error
		if e != nil || job.EventID == "" {
			return e
		}
		job.Attempts++
		job.Status = "claimed"
		job.ClaimToken = uuid.NewString()
		job.LeaseUntil = now + lease
		return tx.Model(&ConverOutbox{}).Where("event_id=?", job.EventID).Updates(map[string]interface{}{"attempts": job.Attempts, "status": job.Status, "claim_token": job.ClaimToken, "lease_until": job.LeaseUntil}).Error
	})
	if e != nil {
		return nil, e
	}
	if job.EventID == "" {
		return nil, nil
	}
	return &job, nil
}

func (s ConverOutboxStore) Finish(ctx context.Context, job ConverOutbox, success bool, now int64) error {
	status, diagnostic, next := "pending", "send_failed", now+retryDelay(job.Attempts)
	if success {
		status, diagnostic, next = "sent", "", 0
	}
	result := s.DB.WithContext(ctx).Model(&ConverOutbox{}).Where("event_id=? AND status='claimed' AND claim_token=?", job.EventID, job.ClaimToken).Updates(map[string]interface{}{"status": status, "diagnostic": diagnostic, "next_attempt_at": next, "lease_until": 0, "claim_token": ""})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return errors.New("outbox claim lost")
	}
	return nil
}
func retryDelay(attempt int64) int64 {
	if attempt > 10 {
		attempt = 10
	}
	if attempt < 1 {
		attempt = 1
	}
	return 1000 * (1 << uint(attempt-1))
}

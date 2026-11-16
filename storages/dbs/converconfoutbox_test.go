package dbs

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/juggleim/jugglechat-server/storages/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestConverOutboxMySQL(t *testing.T) {
	if os.Getenv("CONVERCONF_TEST_MYSQL") != "1" {
		t.Skip("isolated local MySQL required")
	}
	db, e := gorm.Open(mysql.Open("root@tcp(127.0.0.1:13316)/converconf_outbox_test?parseTime=true"), &gorm.Config{})
	if e != nil {
		t.Fatal(e)
	}
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	must := func(e error) {
		t.Helper()
		if e != nil {
			t.Fatal(e)
		}
	}
	for _, table := range []string{"converconf_outbox", "converconf_scope_versions", "converconfs"} {
		must(db.Exec("DROP TABLE IF EXISTS " + table).Error)
	}
	base, e := os.ReadFile("../../commons/dbcommons/sqls/20250201.sql")
	must(e)
	ddl := string(base)[strings.Index(string(base), "CREATE TABLE IF NOT EXISTS `converconfs`"):]
	must(db.Exec(ddl[:strings.Index(ddl, ";")+1]).Error)
	migration, e := os.ReadFile("../../commons/dbcommons/sqls/20261111.sql")
	must(e)
	for i := 0; i < 2; i++ {
		for _, s := range strings.Split(string(migration), ";") {
			if strings.TrimSpace(s) != "" {
				must(db.Exec(s).Error)
			}
		}
	}
	store := ConverOutboxStore{DB: db}
	item := func(k, v string) models.ConverConf {
		return models.ConverConf{AppKey: "a", ConverId: "alice_bob", ConverType: 1, ItemKey: k, ItemValue: v}
	}
	ctx := context.Background()
	t.Run("snapshot noop mixed false and raw", func(t *testing.T) {
		must(store.Save(ctx, "alice", "bob", []models.ConverConf{item("msg_life_time", "42"), item("forbid_forward", "true")}))
		must(store.Save(ctx, "alice", "bob", []models.ConverConf{item("msg_life_time", "42")}))
		must(store.Save(ctx, "bob", "alice", []models.ConverConf{item("forbid_forward", "false"), item("watermark_enabled", "true")}))
		var jobs []ConverOutbox
		must(db.Order("version").Find(&jobs).Error)
		if len(jobs) != 2 || jobs[0].Version != 1 || jobs[1].Version != 2 {
			t.Fatal(jobs)
		}
		var payload map[string]interface{}
		must(json.Unmarshal([]byte(jobs[1].Payload), &payload))
		confs := payload["confs"].(map[string]interface{})
		raw := payload["raw_config"].(map[string]interface{})
		if confs["forbid_forward"] != false || confs["content_protection"] != false || confs["watermark_enabled"] != true || raw["msg_life_time"] != float64(42) || payload["event_id"] != jobs[1].EventID {
			t.Fatal(payload)
		}
		if len(payload["changed_keys"].([]interface{})) != 2 {
			t.Fatal(payload)
		}
		for _, job := range jobs {
			if _, e := uuid.Parse(job.EventID); e != nil {
				t.Fatal(e)
			}
		}
		if payload["type"] != float64(42) {
			t.Fatal("legacy decoder defaults missing type to off", payload)
		}
	})
	t.Run("outbox failure rolls back config and version", func(t *testing.T) {
		must(db.Exec("CREATE TRIGGER reject_event BEFORE INSERT ON converconf_outbox FOR EACH ROW SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='fixture'").Error)
		if store.Save(ctx, "alice", "bob", []models.ConverConf{item("msg_life_time", "7")}) == nil {
			t.Fatal("false success")
		}
		must(db.Exec("DROP TRIGGER reject_event").Error)
		var got ConverConfDao
		must(db.Where("item_key='msg_life_time'").Take(&got).Error)
		if got.ItemValue != "42" {
			t.Fatal(got)
		}
		var version int64
		must(db.Table("converconf_scope_versions").Select("version").Scan(&version).Error)
		if version != 2 {
			t.Fatal(version)
		}
	})
	t.Run("concurrent same scope monotonic versions", func(t *testing.T) {
		var wg sync.WaitGroup
		errors := make(chan error, 12)
		for i := 0; i < 12; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				v := item("msg_life_time", time.Unix(int64(i+100), 0).Format("150405"))
				errors <- store.Save(ctx, "alice", "bob", []models.ConverConf{v})
			}(i)
		}
		wg.Wait()
		close(errors)
		for e := range errors {
			must(e)
		}
		var jobs []ConverOutbox
		must(db.Order("version").Find(&jobs).Error)
		if len(jobs) != 14 {
			t.Fatal(len(jobs))
		}
		for i, j := range jobs {
			if j.Version != int64(i+1) {
				t.Fatal(j)
			}
		}
	})
	t.Run("claim retry lease recovery fencing ordered delivery", func(t *testing.T) {
		first, e := store.Claim(ctx, 1000, 100)
		must(e)
		if first == nil || first.Version != 1 {
			t.Fatal(first)
		}
		other, e := store.Claim(ctx, 1050, 100)
		must(e)
		if other != nil {
			t.Fatal("concurrent same scope claim", other)
		}
		recovered, e := store.Claim(ctx, 1101, 100)
		must(e)
		if recovered == nil || recovered.EventID != first.EventID || recovered.ClaimToken == first.ClaimToken || recovered.Attempts != 2 {
			t.Fatal(recovered)
		}
		if e := store.Finish(ctx, *first, true, 1102); e == nil {
			t.Fatal("stale owner finished")
		}
		must(store.Finish(ctx, *recovered, false, 1102))
		other, e = store.Claim(ctx, 1103, 100)
		must(e)
		if other != nil {
			t.Fatal("retry backoff bypass", other)
		}
		retried, e := store.Claim(ctx, 999999999, 100)
		must(e)
		if retried == nil || retried.Version != 1 {
			t.Fatal(retried)
		}
		must(store.Finish(ctx, *retried, true, 999999999))
		next, e := store.Claim(ctx, 999999999, 100)
		must(e)
		if next == nil || next.Version != 2 {
			t.Fatal(next)
		}
	})
	t.Run("independent scope progresses while predecessor leased", func(t *testing.T) {
		v := item("msg_life_time", "7")
		v.SubChannel = "independent"
		must(store.Save(ctx, "alice", "bob", []models.ConverConf{v}))
		job, e := store.Claim(ctx, 999999999, 100)
		must(e)
		if job == nil || job.ScopeID == (ConverScope{AppKey: "a", ConverID: "alice_bob", ConverType: 1}).ID() {
			t.Fatal(job)
		}
		var event ConverConfigEvent
		must(json.Unmarshal([]byte(job.Payload), &event))
		if event.Type == nil || *event.Type != 7 || event.Confs != nil || len(event.ChangedKeys) != 1 || event.ChangedKeys[0] != "msg_life_time" {
			t.Fatal(event)
		}
		must(store.Finish(ctx, *job, true, 999999999))
	})
	t.Run("concurrent workers claim only one owner", func(t *testing.T) {
		must(db.Exec("UPDATE converconf_outbox SET status='sent',lease_until=0,claim_token=''").Error)
		v := item("msg_life_time", "30")
		v.SubChannel = "workers"
		must(store.Save(ctx, "alice", "bob", []models.ConverConf{v}))
		var wg sync.WaitGroup
		jobs := make(chan *ConverOutbox, 8)
		errs := make(chan error, 8)
		for i := 0; i < 8; i++ {
			wg.Add(1)
			go func() { defer wg.Done(); job, e := store.Claim(ctx, 1000, 100); jobs <- job; errs <- e }()
		}
		wg.Wait()
		close(jobs)
		close(errs)
		for e := range errs {
			must(e)
		}
		count := 0
		for job := range jobs {
			if job != nil {
				count++
			}
		}
		if count != 1 {
			t.Fatal("multiple owners", count)
		}
	})
	t.Run("SQL equivalent scopes share fixed lock identity", func(t *testing.T) {
		for _, pair := range [][2]string{{"thread", "THREAD"}, {"café", "CAFE"}, {"pad", "pad "}} {
			var equal bool
			must(db.Raw("SELECT CONVERT(? USING utf8mb4) COLLATE utf8mb4_0900_ai_ci = CONVERT(? USING utf8mb4) COLLATE utf8mb4_0900_ai_ci", pair[0], pair[1]).Scan(&equal).Error)
			var ids []string
			for i, sub := range pair {
				v := item("msg_life_time", string(rune('7'+i)))
				v.SubChannel = sub
				must(store.Save(ctx, "alice", "bob", []models.ConverConf{v}))
				var job ConverOutbox
				must(db.Where("BINARY JSON_UNQUOTE(JSON_EXTRACT(payload,'$.scope.sub_channel')) = BINARY ?", sub).Order("version DESC").Take(&job).Error)
				ids = append(ids, job.ScopeID)
			}
			if (ids[0] == ids[1]) != equal {
				t.Fatalf("SQL equivalence %q %q=%v but scope IDs %v", pair[0], pair[1], equal, ids)
			}
			t.Logf("SQL collation pair %q/%q equal=%v fixedScope=%v", pair[0], pair[1], equal, ids[0] == ids[1])
		}
	})
	t.Run("concurrent SQL aliases serialize versions and GET metadata", func(t *testing.T) {
		var wg sync.WaitGroup
		failures := make(chan error, 12)
		for i := 0; i < 12; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				v := item("msg_life_time", strconv.Itoa(500+i))
				v.SubChannel = []string{"résumé", "RESUME"}[i%2]
				failures <- store.Save(ctx, "alice", "bob", []models.ConverConf{v})
			}(i)
		}
		wg.Wait()
		close(failures)
		for e := range failures {
			must(e)
		}
		var rows []converScopeVersion
		must(db.Where("sub_channel=?", "resume").Find(&rows).Error)
		if len(rows) != 1 || rows[0].Version != 12 {
			t.Fatal(rows)
		}
		var jobs []ConverOutbox
		must(db.Where("scope_id=?", rows[0].ScopeID).Order("version").Find(&jobs).Error)
		for i, job := range jobs {
			if job.Version != int64(i+1) {
				t.Fatal(jobs)
			}
		}
		scope := ConverScope{AppKey: "a", ConverID: "alice_bob", ConverType: 1, SubChannel: "RESUME"}
		items, meta, e := store.ReadSnapshot(ctx, scope)
		must(e)
		if meta["version"] != int64(12) || meta["event_id"] != jobs[11].EventID || len(items) != 1 {
			t.Fatal(meta, items)
		}
		var event ConverConfigEvent
		must(json.Unmarshal([]byte(jobs[11].Payload), &event))
		if strconv.FormatInt(*event.Type, 10) != items[0].ItemValue {
			t.Fatal("GET version/config snapshot mismatch", meta, items, event)
		}
	})
	t.Run("all legacy lifetime raw values always have type", func(t *testing.T) {
		for _, value := range []string{"0", "1", "7", "30", "90", "180", "42", "86400", "9223372036854775807"} {
			v := item("msg_life_time", value)
			v.SubChannel = "legacy"
			// Seed a different value so even zero generates an actual change event.
			v.ItemValue = "99"
			must(store.Save(ctx, "alice", "bob", []models.ConverConf{v}))
			v.ItemValue = value
			must(store.Save(ctx, "alice", "bob", []models.ConverConf{v}))
			var job ConverOutbox
			must(db.Where("scope_id=?", (ConverScope{AppKey: "a", ConverID: "alice_bob", ConverType: 1, SubChannel: "legacy"}).ID()).Order("version DESC").Take(&job).Error)
			var oldDecoder struct {
				Type int64 `json:"type"`
			}
			must(json.Unmarshal([]byte(job.Payload), &oldDecoder))
			if strconv.FormatInt(oldDecoder.Type, 10) != value || !strings.Contains(job.Payload, "\"type\":") {
				t.Fatal("old decoder reads off or converted value", job.Payload)
			}
		}
	})
}

func TestConverRetryBoundedUnlimited(t *testing.T) {
	if retryDelay(1) != 1000 || retryDelay(10) != 512000 || retryDelay(1000000) != 512000 {
		t.Fatal("retry delay must cap without stopping retries")
	}
}

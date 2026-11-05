package deviceauth

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"gorm.io/gorm"
)

func TestMySQLAuthenticationSerializesRevocationAndRotation(t *testing.T) {
	for _, rotation := range []bool{false, true} {
		name := "revoke"
		if rotation {
			name = "rotate"
		}
		t.Run(name, func(t *testing.T) {
			s := mysqlStore(t)
			ctx := context.Background()
			token, err := s.Issue(ctx, "a", "u", params("d"), []byte("key"), 100)
			if err != nil {
				t.Fatal(err)
			}
			c, _ := Verify(token, []byte("key"), "a")
			// Pause exactly after authentication's read, before its activity write.
			// A row lock must still protect the validated session in this window.
			read, release := make(chan struct{}), make(chan struct{})
			var once sync.Once
			s.DB.Callback().Query().After("gorm:query").Register("test:pause_auth", func(db *gorm.DB) {
				if db.Statement.Context.Value(authReadKey{}) == true {
					once.Do(func() { close(read); <-release })
				}
			})
			defer s.DB.Callback().Query().Remove("test:pause_auth")
			authDone, mutationDone := make(chan error, 1), make(chan error, 1)
			go func() { authDone <- s.Authenticate(context.WithValue(ctx, authReadKey{}, true), c, 101) }()
			select {
			case <-read:
			case <-time.After(3 * time.Second):
				t.Fatal("authentication did not read")
			}
			go func() {
				if rotation {
					_, err := s.Issue(ctx, "a", "u", params("d"), []byte("key"), 102)
					mutationDone <- err
				} else {
					_, err := s.Delete(ctx, Claims{AppKey: "a", UserID: "u", DeviceID: "current"}, "d", 102)
					mutationDone <- err
				}
			}()
			var early bool
			select {
			case <-mutationDone:
				early = true
			case <-time.After(100 * time.Millisecond):
			}
			close(release)
			if err := <-authDone; err != nil {
				t.Fatal(err)
			}
			if !early {
				err := <-mutationDone
				if err != nil && !errors.Is(err, ErrIM) {
					t.Fatal(err)
				}
			}
			if early {
				t.Fatal("session mutation committed between authentication read and activity update")
			}
			if err := s.Authenticate(ctx, c, 103); !errors.Is(err, ErrAuth) {
				t.Fatal("old token accepted after mutation", err)
			}
		})
	}
}

type authReadKey struct{}

func TestMySQLAuthenticationThrottlingDoesNotRejectValidToken(t *testing.T) {
	s := mysqlStore(t)
	ctx := context.Background()
	token, err := s.Issue(ctx, "a", "u", params("d"), []byte("key"), 100)
	if err != nil {
		t.Fatal(err)
	}
	c, _ := Verify(token, []byte("key"), "a")
	for _, now := range []int64{101, 101, 102, 60101, 60102} {
		if err := s.Authenticate(ctx, c, now); err != nil {
			t.Fatalf("valid throttled token at %d: %v", now, err)
		}
		var d Device
		if err := s.DB.Take(&d).Error; err != nil {
			t.Fatal(err)
		}
		want := int64(101)
		if now == 60102 {
			want = 60102
		}
		if d.LastActiveTime == nil || *d.LastActiveTime != want {
			t.Fatalf("activity at %d: %+v", now, d.LastActiveTime)
		}
	}
}

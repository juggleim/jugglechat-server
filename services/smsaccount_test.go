package services

import (
	"errors"
	"testing"

	"github.com/juggleim/jugglechat-server/storages/models"
)

type smsAccountFake struct {
	attempts []models.User
	failures []error
	owner    *models.User
	findErr  error
}

func (s *smsAccountFake) CreateSMSUser(u models.User) error {
	s.attempts = append(s.attempts, u)
	if len(s.failures) == 0 {
		return nil
	}
	e := s.failures[0]
	s.failures = s.failures[1:]
	return e
}
func (s *smsAccountFake) FindByPhone(app, phone string) (*models.User, error) {
	return s.owner, s.findErr
}

func TestCreateSMSAccount(t *testing.T) {
	dbErr := errors.New("database unavailable")
	for _, tc := range []struct {
		name     string
		failures []error
		suffixes []string
		want     string
		attempts int
		code     int
	}{
		{"phone default", nil, nil, "13800138000", 1, 0},
		{"base occupied", []error{models.ErrAccountConflict}, []string{"a1b2"}, "13800138000a1b2", 2, 0},
		{"suffix also occupied", []error{models.ErrAccountConflict, models.ErrAccountConflict}, []string{"a1b2", "c3d4"}, "13800138000c3d4", 3, 0},
		{"other DB error", []error{dbErr}, nil, "", 1, 17000},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := &smsAccountFake{failures: tc.failures}
			i := 0
			suffix := func() (string, error) {
				if i >= len(tc.suffixes) {
					t.Fatal("unexpected suffix generation")
				}
				v := tc.suffixes[i]
				i++
				return v, nil
			}
			u, created, code := createVerifiedSMSUser(s, suffix, models.User{AppKey: "app", UserId: "new", Phone: "13800138000"})
			if int(code) != tc.code || len(s.attempts) != tc.attempts {
				t.Fatalf("code %d attempts %d", code, len(s.attempts))
			}
			if tc.code == 0 && (u == nil || !created || u.LoginAccount != tc.want || u.LoginPass != "") {
				t.Fatalf("%+v created=%v", u, created)
			}
			if tc.code != 0 && (u != nil || created) {
				t.Fatal("failed insert returned user")
			}
		})
	}
	t.Run("bounded exhaustion", func(t *testing.T) {
		s := &smsAccountFake{}
		for i := 0; i < 11; i++ {
			s.failures = append(s.failures, models.ErrAccountConflict)
		}
		n := 0
		u, created, code := createVerifiedSMSUser(s, func() (string, error) { n++; return "aaaa", nil }, models.User{Phone: "13800138000"})
		if u != nil || created || code != 17000 || n != 10 || len(s.attempts) != 11 {
			t.Fatalf("%+v %v %d suffixes=%d inserts=%d", u, created, code, n, len(s.attempts))
		}
	})
	t.Run("phone race returns actual owner", func(t *testing.T) {
		owner := &models.User{UserId: "winner", Phone: "13800138000", Nickname: "winner name"}
		s := &smsAccountFake{failures: []error{models.ErrSMSPhoneConflict}, owner: owner}
		u, created, code := createVerifiedSMSUser(s, func() (string, error) { t.Fatal("phone conflict must not randomize account"); return "", nil }, models.User{UserId: "loser", Phone: owner.Phone})
		if code != 0 || created || u != owner {
			t.Fatalf("%+v %v %d", u, created, code)
		}
	})
	for _, phone := range []string{"+8613800138000", "138_00138000", "12345", "123456789012345678901", "１２３４５６７８９"} {
		t.Run("invalid base "+phone, func(t *testing.T) {
			s := &smsAccountFake{}
			u, created, code := createVerifiedSMSUser(s, func() (string, error) { t.Fatal("invalid phone randomized"); return "", nil }, models.User{Phone: phone})
			if u != nil || created || code != 17003 || len(s.attempts) != 0 {
				t.Fatalf("%+v %v %d", u, created, code)
			}
		})
	}
	t.Run("randomness failure", func(t *testing.T) {
		s := &smsAccountFake{failures: []error{models.ErrAccountConflict}}
		u, created, code := createVerifiedSMSUser(s, func() (string, error) { return "", errors.New("entropy unavailable") }, models.User{Phone: "13800138000"})
		if u != nil || created || code != 17000 || len(s.attempts) != 1 {
			t.Fatalf("%+v %v %d", u, created, code)
		}
	})
	t.Run("phone conflict read failure", func(t *testing.T) {
		for _, e := range []error{nil, dbErr} {
			s := &smsAccountFake{failures: []error{models.ErrSMSPhoneConflict}, findErr: e}
			u, created, code := createVerifiedSMSUser(s, nil, models.User{Phone: "13800138000"})
			if u != nil || created || code != 17000 {
				t.Fatalf("%+v %v %d", u, created, code)
			}
		}
	})
}

func TestSMSAccountSuffix(t *testing.T) {
	for i := 0; i < 100; i++ {
		value, err := smsAccountSuffix()
		if err != nil || len(value) != 4 {
			t.Fatalf("suffix %q: %v", value, err)
		}
		for _, c := range value {
			if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9') {
				t.Fatalf("invalid suffix %q", value)
			}
		}
	}
}

func TestSMSAccountLengthBoundary(t *testing.T) {
	for _, conflict := range []bool{false, true} {
		s := &smsAccountFake{}
		if conflict {
			s.failures = []error{models.ErrAccountConflict}
		}
		u, created, code := createVerifiedSMSUser(s, func() (string, error) { t.Fatal("must not generate oversized candidate"); return "", nil }, models.User{Phone: "12345678901234567890"})
		if conflict {
			if u != nil || created || code != 17003 {
				t.Fatalf("%+v %v %d", u, created, code)
			}
		} else if u == nil || !created || code != 0 || u.LoginAccount != "12345678901234567890" {
			t.Fatalf("%+v %v %d", u, created, code)
		}
	}
}

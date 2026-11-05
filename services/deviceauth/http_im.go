package deviceauth

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha1" // IM Server's verified management API signature protocol.
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// IMConfig is resolved per app at runtime; never include it in logs/errors.
type IMConfig struct {
	BaseURL, AppKey, Secret string
	DeviceKick              bool // Explicit opt-in for kick/reconciliation; default false.
}

// HTTPIM kicks one device and checks current offline status. It cannot fence
// asynchronous upstream execution or invalidate existing IM credentials.
type HTTPIM struct {
	Resolve func(context.Context, string) (IMConfig, error)
	Timeout time.Duration
	Now     func() time.Time
}

func (a *HTTPIM) now() time.Time {
	if a.Now != nil {
		return a.Now()
	}
	return time.Now()
}

func (a *HTTPIM) bounded(ctx context.Context) (context.Context, context.CancelFunc) {
	d := a.Timeout
	if d <= 0 || d > 5*time.Second {
		d = 5 * time.Second
	}
	return context.WithTimeout(ctx, d)
}

func imError(reason string) error { return fmt.Errorf("%w: %s", ErrIM, reason) }

func (a *HTTPIM) config(ctx context.Context, app string) (IMConfig, error) {
	if a.Resolve == nil || app == "" {
		return IMConfig{}, imError("app configuration unavailable")
	}
	c, err := a.Resolve(ctx, app)
	if ctx.Err() != nil {
		return IMConfig{}, fmt.Errorf("%w: %w", ErrIM, ctx.Err())
	}
	if err != nil || c.AppKey != app || c.Secret == "" {
		return IMConfig{}, imError("app configuration unavailable")
	}
	u, err := url.Parse(c.BaseURL)
	if err != nil || u.Host == "" || (u.Scheme != "https" && u.Scheme != "http") || u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Path != "" && u.Path != "/") {
		return IMConfig{}, imError("invalid IM API origin")
	}
	c.BaseURL = strings.TrimRight(c.BaseURL, "/")
	return c, nil
}

// call performs bounded requests without automatic retries or redirects.
func (a *HTTPIM) call(ctx context.Context, c IMConfig, method, path string, payload interface{}, out interface{}) error {
	var body []byte
	var err error
	if payload != nil {
		body, err = json.Marshal(payload)
		if err != nil {
			return imError("invalid request")
		}
	}
	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, bytes.NewReader(body))
	if err != nil {
		return imError("invalid request")
	}
	var random [16]byte
	if _, err = rand.Read(random[:]); err != nil {
		return imError("nonce unavailable")
	}
	nonce := hex.EncodeToString(random[:])
	ts := strconv.FormatInt(a.now().UnixMilli(), 10)
	sig := sha1.Sum([]byte(c.Secret + nonce + ts))
	req.Header.Set("appkey", c.AppKey)
	req.Header.Set("nonce", nonce)
	req.Header.Set("timestamp", ts)
	req.Header.Set("signature", hex.EncodeToString(sig[:]))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	resp, err := client.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("%w: %w", ErrIM, ctx.Err())
		}
		return imError("transport failed; result unknown")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return imError("HTTP status rejected")
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, (1<<20)+1))
	if ctx.Err() != nil {
		return fmt.Errorf("%w: %w", ErrIM, ctx.Err())
	}
	if err != nil || len(b) > 1<<20 {
		return imError("invalid response body")
	}
	var envelope struct {
		Code *int            `json:"code"`
		Data json.RawMessage `json:"data"`
	}
	if json.Unmarshal(b, &envelope) != nil || envelope.Code == nil || *envelope.Code != 0 {
		return imError("invalid response or API error")
	}
	if out != nil {
		if len(envelope.Data) == 0 || bytes.Equal(bytes.TrimSpace(envelope.Data), []byte("null")) || json.Unmarshal(envelope.Data, out) != nil {
			return imError("incomplete response data")
		}
	}
	return nil
}

func (a *HTTPIM) Online(ctx context.Context, app, user string) (map[string]bool, error) {
	ctx, cancel := a.bounded(ctx)
	defer cancel()
	c, err := a.config(ctx, app)
	if err != nil {
		return nil, err
	}
	return a.online(ctx, c, user)
}

func (a *HTTPIM) online(ctx context.Context, c IMConfig, user string) (map[string]bool, error) {
	if user == "" {
		return nil, imError("missing user")
	}
	var data struct {
		Items []*struct {
			UserID   string `json:"user_id"`
			IsOnline *bool  `json:"is_online"`
			Devices  []*struct {
				DeviceID string `json:"device_id"`
				Platform string `json:"platform"`
			} `json:"devices"`
		} `json:"items"`
	}
	req := struct {
		UserIDs []string `json:"user_ids"`
	}{[]string{user}}
	if err := a.call(ctx, c, http.MethodPost, "/apigateway/users/onlinestatus/query", req, &data); err != nil {
		return nil, err
	}
	if len(data.Items) != 1 || data.Items[0] == nil {
		return nil, imError("missing or duplicate online user")
	}
	i := data.Items[0]
	if i.UserID != user || i.IsOnline == nil || i.Devices == nil || *i.IsOnline != (len(i.Devices) > 0) {
		return nil, imError("incomplete online status")
	}
	result := make(map[string]bool, len(i.Devices))
	for _, d := range i.Devices {
		if d == nil || !validDeviceID(d.DeviceID) || strings.TrimSpace(d.Platform) == "" || !validText(d.Platform, 32) || result[d.DeviceID] {
			return nil, imError("invalid online device")
		}
		result[d.DeviceID] = true
	}
	return result, nil
}

func (a *HTTPIM) revokeConfig(ctx context.Context, j RevokeJob) (IMConfig, error) {
	if j.OperationID == "" || j.TargetSessionID == "" || j.UserID == "" || !validDeviceID(j.DeviceID) {
		return IMConfig{}, imError("invalid revocation target")
	}
	c, e := a.config(ctx, j.AppKey)
	if e != nil {
		return c, e
	}
	if !c.DeviceKick {
		return c, imError("device IM kick integration disabled; business session revoked, IM disconnect pending")
	}
	return c, nil
}

func (a *HTTPIM) targetOffline(ctx context.Context, c IMConfig, j RevokeJob) (bool, error) {
	online, e := a.online(ctx, c, j.UserID)
	if e != nil {
		return false, e
	}
	if online[j.DeviceID] {
		return false, nil
	}
	return true, nil
}

func (a *HTTPIM) Reconcile(ctx context.Context, j RevokeJob) (bool, error) {
	ctx, cancel := a.bounded(ctx)
	defer cancel()
	c, e := a.revokeConfig(ctx, j)
	if e != nil {
		return false, e
	}
	return a.targetOffline(ctx, c, j)
}

func (a *HTTPIM) Disconnect(ctx context.Context, j RevokeJob) error {
	ctx, cancel := a.bounded(ctx)
	defer cancel()
	c, e := a.revokeConfig(ctx, j)
	if e != nil {
		return e
	}
	// Deliberately omit platforms/ext: never widen a device kick to a platform.
	kick := struct {
		UserID    string   `json:"user_id"`
		DeviceIDs []string `json:"device_ids"`
	}{j.UserID, []string{j.DeviceID}}
	if e = a.call(ctx, c, http.MethodPost, "/apigateway/users/kick", kick, nil); e != nil {
		return e
	}
	// Current offline observation is the contract, not proof of a particular
	// connection/session being kicked. An asynchronous acknowledgement is insufficient.
	offline, e := a.targetOffline(ctx, c, j)
	if e != nil {
		return e
	}
	if !offline {
		return imError("target device still online after kick")
	}
	return nil
}

var _ IMAdapter = (*HTTPIM)(nil)

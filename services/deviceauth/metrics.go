package deviceauth

import "sync/atomic"

// Credential-free migration counters for monitoring hooks. No user/device IDs,
// secrets or credentials are retained. Counters reset on process restart.
var LegacyLogins atomic.Uint64
var LegacyAuthentications atomic.Uint64
var LegacyQRRedemptions atomic.Uint64

package accounts

import "regexp"

var pattern = regexp.MustCompile(`^[a-zA-Z0-9]{6,20}$`)

// Valid uses the registration identity rule without trimming or case folding.
func Valid(account string) bool { return pattern.MatchString(account) }

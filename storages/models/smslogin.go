package models

import "errors"

var ErrSMSPhoneConflict = errors.New("SMS login phone binding conflict")
var ErrSMSLegacyMissing = errors.New("SMS legacy user disappeared")

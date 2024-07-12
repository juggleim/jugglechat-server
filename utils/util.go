package utils

import (
	"encoding/binary"
	"encoding/json"
	"math/rand"
	"strconv"
	"time"
)

func BytesToUInt64(buf []byte) uint64 {
	return binary.BigEndian.Uint64(buf)
}

func init() {
	rand.Seed(time.Now().UnixNano())

}
func RandIntn(i int) int {
	return rand.Intn(i)
}

func ParseInt64(value string) (int64, error) {
	ret, err := strconv.ParseInt(value, 10, 64)
	return ret, err
}

func ParseInt(value string) (int, error) {
	ret, err := strconv.Atoi(value)
	return ret, err
}

func ParseFloat(val string) (float64, error) {
	floatvar, err := strconv.ParseFloat(val, 64)
	return floatvar, err
}

func Int2String(val int64) string {
	return strconv.FormatInt(val, 10)
}

func ToJson(val interface{}) string {
	bs, err := json.Marshal(val)
	if err == nil {
		return string(bs)
	}
	return ""
}

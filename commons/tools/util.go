package tools

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"math/rand"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"google.golang.org/protobuf/proto"
)

var random *rand.Rand

func init() {
	random = rand.New(rand.NewSource(time.Now().UnixNano()))
}

func RandInt(n int) int {
	return random.Intn(n)
}

func BytesToUInt64(buf []byte) uint64 {
	return binary.BigEndian.Uint64(buf)
}

func RandIntn(i int) int {
	return random.Intn(i)
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

func BoolPtr(f bool) *bool {
	return &f
}

func IntPtr(i int) *int {
	return &i
}

func String2Int64(str string) (int64, error) {
	return strconv.ParseInt(str, 10, 64)
}

func ToInt(str string) int {
	intVal, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0
	}
	return int(intVal)
}

func PbMarshal(obj proto.Message) ([]byte, error) {
	bytes, err := proto.Marshal(obj)
	return bytes, err
}
func PbUnMarshal(bytes []byte, typeScope proto.Message) error {
	err := proto.Unmarshal(bytes, typeScope)
	return err
}

func JsonMarshal(obj interface{}) ([]byte, error) {
	return json.Marshal(obj)
}

func JsonUnMarshal(bytes []byte, obj interface{}) error {
	return json.Unmarshal(bytes, obj)
}

func HmacSha256(key []byte, data string) []byte {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(data))
	mac := h.Sum(nil)
	return mac
}

func HmacSha1(key []byte, data string) []byte {
	h := hmac.New(sha1.New, key)
	h.Write([]byte(data))
	mac := h.Sum(nil)
	return mac
}

func MapToStruct[T any](m map[string]interface{}) T {
	var t T
	data, _ := json.Marshal(m)
	_ = json.Unmarshal(data, &t)

	return t
}

func SHA1(s string) string {
	o := sha1.New()
	o.Write([]byte(s))
	return hex.EncodeToString(o.Sum(nil))
}

func GetConversationId(fromId, targetId string, channelType int32) string {
	if channelType == 1 {
		array := []string{fromId, targetId}
		sort.Strings(array)
		return strings.Join(array, ":")
	} else {
		return targetId
	}
}

func MaskEmail(email string) string {
	// 分割用户名和域名
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return email
	}
	username := parts[0]
	domain := parts[1]

	// 处理用户名脱敏
	var maskedUsername string
	switch len(username) {
	case 0:
		maskedUsername = "" // 用户名为空（异常情况）
	case 1:
		maskedUsername = "*" // 用户名仅1位：全隐藏
	case 2:
		maskedUsername = username // 用户名2位：不隐藏（太短，隐藏后失去辨识度）
	case 3:
		// 用户名3位：前2位+后1位（实际是前2位+最后1位，中间无字符，直接拼接）
		maskedUsername = username[:2] + "*"
	default:
		// 用户名≥4位：前2位 + 中间* + 最后1位
		maskedUsername = username[:2] + "***" + username[len(username)-1:]
	}

	// 拼接脱敏后的邮箱
	return maskedUsername + "@" + domain
}

func MaskPhone(phone string) string {
	start := 3
	length := 4
	if valid, _ := regexp.MatchString(`^1\d{10}$`, phone); !valid {
		return phone
	}
	// 校验起始位置和长度是否合法（不超出手机号范围）
	if start < 0 || start+length > len(phone) {
		return phone
	}
	// 替换指定位置为*
	runes := []rune(phone)
	for i := 0; i < length; i++ {
		runes[start+i] = '*'
	}
	return string(runes)
}

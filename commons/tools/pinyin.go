package tools

import (
	"strings"

	"github.com/mozillazg/go-pinyin"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
var letterMap map[rune]bool

func init() {
	letterMap = make(map[rune]bool)
	for _, r := range letters {
		letterMap[r] = true
	}
}

func GetFirstLetter(str string) string {
	str = strings.TrimSpace(str)
	if str == "" {
		return ""
	}
	array := []rune(str)
	if letterMap[array[0]] {
		return strings.ToLower(string(array[0]))
	} else {
		opts := pinyin.NewArgs()
		opts.Style = pinyin.Normal
		pyArr := pinyin.LazyPinyin(string(array[0]), opts)
		if len(pyArr) > 0 {
			str = pyArr[0]
			array = []rune(str)
			if letterMap[array[0]] {
				return strings.ToLower(string(array[0]))
			}
		}
	}
	return strings.ToLower(string(array[0]))
}

func GetPinyin(str string) string {
	str = strings.TrimSpace(str)
	if str == "" {
		return "#"
	}
	array := []rune(str)
	if letterMap[array[0]] {
		return strings.ToLower(string(array[0]))
	} else {
		opts := pinyin.NewArgs()
		opts.Style = pinyin.Normal
		pyArr := pinyin.LazyPinyin(str, opts)
		if len(pyArr) > 0 {
			ret := strings.Join(pyArr, "")
			ret = strings.TrimSpace(ret)
			if len(ret) > 0 {
				return string(ret[0])
			} else {
				return "#"
			}
		} else {
			return "#"
		}
	}
}

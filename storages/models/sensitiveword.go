package models

type SensitiveWordType int

const (
	SensitiveWordType_deny_word    SensitiveWordType = 1
	SensitiveWordType_replace_word SensitiveWordType = 2
)

type SensitiveWord struct {
	ID       int64
	Word     string
	WordType SensitiveWordType
	AppKey   string
}

type ISensitiveWordStorage interface {
	BatchUpsert(items []SensitiveWord) error
	UpdateWord(appkey string, wordStr string, wordType int) error
	DeleteWords(appkey string, words ...string) error
	QrySensitiveWords(appkey string, limit, startId int64) ([]*SensitiveWord, error)
	QrySensitiveWordsWithPage(appkey string, page, size int64, str string, wordType int32) (words []*SensitiveWord, total int, err error)
	Total(appkey string) int
}

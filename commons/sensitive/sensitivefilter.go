package sensitive

import (
	"fmt"
	"sync"
	"time"

	"github.com/juggleim/jugglechat-server/commons/caches"
	"github.com/juggleim/jugglechat-server/storages"
	"github.com/juggleim/jugglechat-server/storages/models"
)

var (
	filterCache *caches.LruCache
	filterLocks *sync.RWMutex
)

func init() {
	filterCache = caches.NewLruCacheWithAddReadTimeout("filter_cache", 10000, nil, 8*time.Minute, 10*time.Minute)
	filterLocks = &sync.RWMutex{}

	filterCache.SetValueCreator(func(key interface{}) interface{} {
		s := NewSensitiveService()
		start := time.Now()
		loadAppWords(s, key.(string))
		fmt.Println("load app words cost:", time.Since(start))
		return s
	})
}

func GetAppSensitiveFilter(appKey string) *SensitiveService {
	filterLocks.Lock()
	defer filterLocks.Unlock()

	v, ok := filterCache.GetByCreator(appKey, nil)
	if !ok {
		return nil
	}
	return v.(*SensitiveService)
}

func loadAppWords(service *SensitiveService, appKey string) (err error) {
	var (
		startId  int64 = 0
		pageSize int64 = 1000
	)
	storage := storages.NewSensitiveWordStorage()
	for {
		list, err := storage.QrySensitiveWords(appKey, pageSize, startId)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		for _, item := range list {
			if startId < item.ID {
				startId = item.ID
			}
		}
		service.AddWord(list...)
		if len(list) < int(pageSize) {
			break
		}
	}
	return nil
}

type SensitiveService struct {
	replaceFilter *Filter
	denyFilter    *Filter
	loadLock      *sync.RWMutex
}

func NewSensitiveService() *SensitiveService {
	return &SensitiveService{
		replaceFilter: NewFilter(),
		denyFilter:    NewFilter(),
		loadLock:      &sync.RWMutex{},
	}
}

func (s *SensitiveService) ReplaceSensitiveWords(text string) (isDeny bool, replacedText string) {
	s.loadLock.RLock()
	defer s.loadLock.RUnlock()

	if s.denyFilter != nil {
		var ok bool
		ok, _ = s.denyFilter.FindIn(text)
		if ok {
			isDeny = true
			return
		}
	}
	if s.replaceFilter != nil {
		replacedText = s.replaceFilter.Replace(text, '*')
	}

	return
}

func (s *SensitiveService) AddWord(words ...*models.SensitiveWord) {
	s.loadLock.Lock()
	defer s.loadLock.Unlock()
	if s.replaceFilter == nil {
		s.replaceFilter = NewFilter()
	}
	if s.denyFilter == nil {
		s.denyFilter = NewFilter()
	}
	for _, word := range words {
		if word.WordType == models.SensitiveWordType_deny_word {
			s.denyFilter.AddWord(word.Word)
		} else {
			s.replaceFilter.AddWord(word.Word)
		}
	}
}

func (s *SensitiveService) DelWord(words ...string) {
	s.loadLock.Lock()
	defer s.loadLock.Unlock()
	if s.denyFilter != nil {
		s.denyFilter.DelWord(words...)
	}
	if s.replaceFilter != nil {
		s.replaceFilter.DelWord(words...)
	}
}

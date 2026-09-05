package template

import (
	"log"
	"obsidian-config-sync/go_src/infa"

	"github.com/cockroachdb/errors"
)

// Cache Template v1.2

var (
	ExampleCache = newExampleCache()
)

func newExampleCache() *exampleCache {
	data := infa.NewStrCache[*ExampleCacheData]()
	return &exampleCache{data: data}
}

// 被缓存的数据类型
type ExampleCacheData struct {}

// exampleCache 用于缓存xxxxx
type exampleCache struct {
	data *infa.StrCache[*ExampleCacheData]
}


// Refresh
func (t *exampleCache) Refresh(keys []string) error {
	// 假如需要登陆/初始化
	// provide.Login()

	for _, key := range keys {
		if err := t.loadData(key); err != nil {
			return errors.WithStack(err)
		}
	}

	log.Printf("✅ Cache example, load keys %s", infa.PeekSlice(keys))
	return nil
}

// 加载数据
func (t *exampleCache) loadData(key string) error {
	data, err := t.todo_LoadData(key)
	if err != nil {
		return errors.WithStack(err)
	}
	t.data.Set(key, data)
	return nil
}

// Get
func (t *exampleCache) Get(key string) (*ExampleCacheData, error) {
	data, ok := t.data.GetKey(key)
	if ok {
		return data, nil
	}

	// 先加载 后再取
	if err := t.loadData(key); err != nil {
		return nil, errors.WithStack(err)
	}

	data, ok = t.data.GetKey(key)
	if ok {
		return data, nil
	}
	return data, errors.Newf("example cache. get and load failed. %+v", key)
}

// todo 
func (t *exampleCache) todo_LoadData(key string) (*ExampleCacheData, error) {
	_ = key
	return nil, nil
}



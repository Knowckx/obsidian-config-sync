package file_store

import (
	"reflect"
	"testing"

	"obsidian-config-sync/go_src/infa/ops"
)

type testConfigItem struct {
	Name  string `toml:"name"`
	Count int    `toml:"count"`
}

func Test_SaveGet_StructSliceAndMap(t *testing.T) {
	// storePath := filepath.Join(t.TempDir(), "config.toml")
	// err := os.WriteFile(storePath, nil, 0600)
	// ops.MustNoErrInTest(t, err)

	// store := &Store{path: storePath}

	store,err := newStore("go-trade", "config.toml")
	ops.MustNoErrInTest(t, err)

	wantList := []testConfigItem{
		{Name: "first", Count: 1},
		{Name: "second", Count: 2},
	}
	err = store.Save("list", wantList)
	ops.MustNoErrInTest(t, err)

	var gotList []testConfigItem
	err = store.Get("list", &gotList)
	ops.MustNoErrInTest(t, err)
	if !reflect.DeepEqual(gotList, wantList) {
		t.Fatalf("list mismatch\nwant: %+v\ngot:  %+v", wantList, gotList)
	}

	wantMap := map[string]testConfigItem{
		"a": {Name: "alpha", Count: 10},
		"b": {Name: "beta", Count: 20},
	}
	err = store.Save("item_map", wantMap)
	ops.MustNoErrInTest(t, err)

	var gotMap map[string]testConfigItem
	err = store.Get("item_map", &gotMap)
	ops.MustNoErrInTest(t, err)
	if !reflect.DeepEqual(gotMap, wantMap) {
		t.Fatalf("map mismatch\nwant: %+v\ngot:  %+v", wantMap, gotMap)
	}
}

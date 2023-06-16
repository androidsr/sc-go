package sleveldb

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/syndtr/goleveldb/leveldb"
)

var (
	db *leveldb.DB
)

func New(path string) {
	db, _ = leveldb.OpenFile(path, nil)
}

func WriteString(key string, v string) {
	if key == "" {
		return
	}
	db.Put([]byte(key), []byte(v), nil)
}

func WriteStruct(key string, v interface{}) {
	if key == "" {
		return
	}
	bs, err := json.Marshal(v)
	if err != nil {
		fmt.Println(err)
		return
	}
	db.Put([]byte(key), bs, nil)
}

func Read[T any]() []*T {
	iter := db.NewIterator(nil, nil)
	var datas []*T
	for iter.Next() {
		bs := iter.Value()
		data := new(T)
		json.Unmarshal(bs, data)
		datas = append(datas, data)
	}
	return datas
}

func ReadPrefix[T any](key string) []*T {
	iter := db.NewIterator(nil, nil)
	var datas []*T
	for iter.Next() {
		ky := iter.Key()
		if !strings.HasPrefix(string(ky), key) {
			continue
		}
		bs := iter.Value()
		data := new(T)
		json.Unmarshal(bs, data)
		datas = append(datas, data)
	}
	return datas
}

func Delete(key string) {
	db.Delete([]byte(key), nil)
}

func Get(key string) []byte {
	v, _ := db.Get([]byte(key), nil)
	return v
}

func GetStruct[T any](key string) *T {
	v, _ := db.Get([]byte(key), nil)
	data := new(T)
	json.Unmarshal(v, data)
	return data
}

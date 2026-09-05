package template

import (
	"fmt"
	"obsidian-config-sync/go_src/infa"
)

// struct Template v1.0

type Student struct {
	Name string // 名字
}

func (t Student) String() string {
	return fmt.Sprintf("Name: %s", t.Name)
}

// 切片类型
type Students []Student

func (t Students) String() string {
	msg := infa.PeekSlice(t)
	return msg
}

// Reverse 原地反转 K 线顺序。
func (t Students) Reverse() {
	for left, right := 0, len(t)-1; left < right; left, right = left+1, right-1 {
		t[left], t[right] = t[right], t[left]
	}
}

func (t Students) ToMap() StudentMap {
	out := make(map[string]Student, len(t))
	for _, item := range t {
		out[item.Name] = item
	}
	return out
}

// map检索类型
type StudentMap map[string]Student

func (t StudentMap) String() string {
	msg := infa.PeekMap(t)
	return msg
}



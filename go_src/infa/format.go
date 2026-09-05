package infa

import (
	"fmt"
	"math"
	"obsidian-config-sync/go_src/infa/ops"
)

// PeekStr 输出字符串长度和前缀预览。
func PeekStr(dataStr string) string {
	const limit = 30

	runes := []rune(dataStr)
	if len(runes) <= limit {
		return fmt.Sprintf("len=%d %s", len(dataStr), dataStr)
	}

	return fmt.Sprintf("len=%d %s...", len(dataStr), string(runes[:limit]))
}

// PeekSlice 默认预览3行数据
func PeekSlice[T any](ins []T) string {
	return PeekSliceN(ins, 3, "\n")
}

func PeekSliceFull[T any](ins []T) string {
	return PeekSliceN(ins, len(ins), "\n")
}

// PeekSliceN(lis, 5, " | ")   自定义数量和分隔符
func PeekSliceN[T any](ins []T, n int, sep string) string {
	n = ops.IfElse(n > len(ins), len(ins), n)

	out := NewStrJoiner()
	out.Addf("len=%d", len(ins))

	for i := range n {
		if sep == "\n"{
			out.Addf("%d %v", i, ins[i])
			continue
		}
		out.Addf("%v", ins[i])
	}

	return out.StringJoin(sep)
}

// PeekMap 预览切片 map 长度和样例键值。
func PeekMap[K comparable, V any](items map[K]V) string {
	out := NewStrJoiner()
	out.Addf("len=%d", len(items))
	cnt := 0
	for k, v := range items {
		out.Addf("%v=%v", k, v)
		cnt++
		if cnt >= 3 {
			break
		}
	}
	return out.String()
}

// 输入一个变动百分比，转换成阅读友好格式 10.12 → +10.12%
func FmtRate(numVal float64) string {
	fix := ops.IfElse(numVal > 0, "+", "")
	return fmt.Sprintf("%s%.3f%%", fix, numVal*100)
}

// FmtAmountW 金额以 万 来显示。
func FmtAmountW(value float64) string {
	if math.Abs(value) > 1000 {
		return fmt.Sprintf("%.1fw", value/10000)
	}
	if value == 0 {
		return "0"
	}
	return fmt.Sprintf("%.1f", value)
}

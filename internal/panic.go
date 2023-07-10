package internal

import "fmt"

func Panicf(format string, args ...any) {
	panic(fmt.Sprintf(format, args...))
}

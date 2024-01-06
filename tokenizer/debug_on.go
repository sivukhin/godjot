//go:build debug_on || test

package tokenizer

import (
	"fmt"
	"path/filepath"
	"runtime"
)

//go:noinline
func Assertf(condition bool, format string, args ...any) {
	if !condition {
		_, file, line, _ := runtime.Caller(1)
		panic(fmt.Errorf(fmt.Sprintf("%v:%v: assertion failed: ", filepath.Base(file), line)+format, args...))
	}
}

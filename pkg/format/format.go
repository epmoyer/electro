package format

import (
	"errors"
	"fmt"
	"path"
	"runtime"
)

func SprintfContext(format string, a ...interface{}) string {
	return fmt.Sprintf(getContextOfCaller()+format, a...)
}

func ErrorfContext(format string, a ...interface{}) error {
	return errors.New(SprintfContext(format, a...))
}

func getContextOfCaller() string {
	// The (2) gets the caller of our caller
	pc, file, line, ok := runtime.Caller(2)
	if !ok {
		return "unknown"
	}
	funcName := runtime.FuncForPC(pc).Name()
	fileName := path.Base(file) // Extract the file name from the full path
	return fmt.Sprintf("%s:%d:%s()::", fileName, line, funcName)
}

func CurrentFunctionName() string {
	// 1 to get the caller of this function
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		return "unknown"
	}
	return runtime.FuncForPC(pc).Name() + "()"
}

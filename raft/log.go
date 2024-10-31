package raft

import (
	"fmt"
	"runtime"
	"strings"
	"time"
)

var Name string

func Info(data ...interface{}) {
	var (
		buf strings.Builder
	)

	for _, v := range data {
		buf.WriteString(fmt.Sprintf("%+v", v))
	}
	_, file, line, _ := runtime.Caller(1)
	fmt.Printf("%s[info]%s:%d(%s)-->%s \n", Name, file, line, time.Now(), buf.String())
}
func Error(data ...interface{}) {
	var (
		buf strings.Builder
	)

	for _, v := range data {
		buf.WriteString(fmt.Sprintf("%+v", v))
	}
	_, file, line, _ := runtime.Caller(1)
	fmt.Printf("%s[error]%s:%d(%s)-->%s \n", Name, file, line, time.Now(), buf.String())
}

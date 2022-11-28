package log

import (
	"fmt"
	"time"
)

func Infof(format string, args ...interface{}) {
	fmt.Printf(time.Now().Format("2006-01-02 15:04:05")+" "+format+"\n", args...)
}

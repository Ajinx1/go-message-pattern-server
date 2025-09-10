package utility

import (
	"fmt"
	"runtime"

	log "github.com/jeanphorn/log4go"
)

func LogAndPrint(data interface{}, args ...interface{}) {
	if len(args) < 1 {
		fmt.Println(data)
		log.Log(log.INFO, getSource(), fmt.Sprintf("%v", data))
		return
	}

	fmt.Println(data, args)
	log.Log(log.INFO, getSource(), fmt.Sprintf("data: %v\n args: %v", data, args))
}

func getSource() (source string) {
	if pc, _, line, ok := runtime.Caller(2); ok {
		source = fmt.Sprintf("%s:%d", runtime.FuncForPC(pc).Name(), line)
	}
	return
}

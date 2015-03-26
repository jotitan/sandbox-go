package arguments

import (
	"strings"
	"os"
)

func ParseArgs() map[string]string {
	args := make(map[string] string)
	currentKey := ""
	for _, value := range os.Args[1:] {
		if strings.HasPrefix(value, "-") {
			if _,ok := args[currentKey] ; !ok {
				args[currentKey] = ""
			}
			currentKey = value[1:]
		}else {
			if currentKey != "" {
				args[currentKey] = value
			}
			currentKey = ""
		}
	}
	return args
}

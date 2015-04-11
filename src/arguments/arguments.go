package arguments

import (
	"strings"
	"os"
	"strconv"
)

type Arguments struct {
	data map[string] string
}

func NewArguments()Arguments{
	args := Arguments{make(map[string]string)}
	args.parse()
	return args
}

func (args Arguments)Check(names []string)bool{
	for _,name := range names {
		if _,ok := args.data[name] ; !ok {
			return false
		}
	}
	return true
}

func (args * Arguments) parse(){
	currentKey := ""
	for _, value := range os.Args[1:] {
		if strings.HasPrefix(value, "-") {
			if _,ok := args.data[currentKey] ; !ok {
				args.data[currentKey] = ""
			}
			currentKey = value[1:]
		}else {
			if currentKey != "" {
				args.data[currentKey] = value
			}
			currentKey = ""
		}
	}
}

func (args Arguments)GetUInt(argName string)uint{
	if value,ok := args.data[argName]; ok {
		if intValue, err := strconv.ParseUint(value,10,0) ; err == nil{
			return uint(intValue)
		}
		return 0
	}
	return 0
}

func (args Arguments)GetString(argName string)string{
	if value,ok := args.data[argName]; ok {
		return value
	}
	return ""
}


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

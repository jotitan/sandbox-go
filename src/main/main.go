package main

import (
	"arguments"
	"fmt"
	"crypt"
	"os"
	"path/filepath"
	"strings"
	"time"
)


func main(){

	args := arguments.ParseArgs()
	if _, ok := args["crypt"] ; ok {
		crypt.DataToImage(args["to"],args["from"],args["key"])
		return
	}
	if _, ok := args["decrypt"] ; ok {
		crypt.ImageToData(args["from"],args["to"],args["key"])
		return
	}
	if _, ok := args["zipcrypt"] ; ok {
		// Put image in temporary file
		zipFile := filepath.Join(os.TempDir(),fmt.Sprintf("tempfile_%d",time.Now().UnixNano()))
		excludes := make(map[string]struct{})
		if extensions,ok := args["excludes"] ; ok {
			for _,ext := range strings.Split(extensions,","){
				excludes[ext] = struct{}{}
			}
		}
		crypt.Archive(args["from"],zipFile,excludes)
		crypt.DataToImage(args["to"],zipFile,args["key"])
		return
	}
	fmt.Printf("Nothing to launch")
}

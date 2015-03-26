package main

import (
	"arguments"
	"fmt"
)


func main(){

	args := arguments.ParseArgs()

	if _, ok := args["crypt"] ; ok {
		fmt.Println("CRYPT")
		//crypt.DataToImage(args["to"],args["from"],args["key"])
	}
	if _, ok := args["decrypt"] ; ok {
		fmt.Println("DECRYPT")
		//crypt.ImageToData(args["from"],args["to"],args["key"])
	}

}

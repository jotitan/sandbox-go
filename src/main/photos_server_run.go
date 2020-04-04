package main

import (
	"os"
	"photos_server"
)

func main(){
	server := photos_server.NewPhotosServer(os.Args[1],os.Args[2])
	server.Launch()
}

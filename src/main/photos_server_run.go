package main

import (
	"arguments"
	"photos_server"
)

func main(){
	args := arguments.NewArguments()
	cacheFolder := args.GetMandatoryString("cache","Argument -cache is mandatory to specify where pictures are resized")
	webResources := args.GetMandatoryString("resources","Argument -resources is mandatory to specify where web resources are")

	server := photos_server.NewPhotosServer(cacheFolder,webResources)
	server.Launch()
}

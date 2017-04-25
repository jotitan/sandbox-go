package main

import (
	"net/http"
	"fmt"
	"path/filepath"
	"io"
	"os"
	"arguments"
	"encoding/json"
)

type UpManager struct {
	folder string
	resourcesFolder string
}
/* create a server to upload files */

func (um UpManager)uploadFile(w http.ResponseWriter,r * http.Request){
	if file,header,err := r.FormFile("myfile") ; err == nil {
		copyTo(file,filepath.Join(um.folder,header.Filename))
	}
}

func (um UpManager) cancel(w http.ResponseWriter,r * http.Request){
	dir,_ := os.Open(um.folder)
	defer dir.Close()
	files,_ := dir.Readdirnames(-1)
	for _,file := range files{
		os.Remove(filepath.Join(um.folder,file))
	}
}

// List files in folder
func (um UpManager)list(w http.ResponseWriter,r * http.Request){
	w.Header().Set("Content-type","application/json")
	dir,_ := os.Open(um.folder)
	defer dir.Close()
	files,_ := dir.Readdirnames(-1)
	data ,_ := json.Marshal(files)
	w.Write(data)
}

func (um UpManager) getNbFiles()int{
	dir,_ := os.Open(um.folder)
	defer dir.Close()
	files,_ := dir.Readdirnames(-1)
	return len(files)
}

// Launch import (call a .sh)
func (um UpManager)launch(w http.ResponseWriter,r * http.Request){
	fmt.Println("Do something")
}

func copyTo(input io.Reader,outputPath string){
	if outputFile,err := os.OpenFile(outputPath,os.O_CREATE|os.O_RDWR|os.O_TRUNC,os.ModePerm) ; err == nil {
		defer outputFile.Close()
		if _,err := io.Copy(outputFile,input) ; err == nil {
			fmt.Println("File well copied in",outputPath)
		}
	}else{
		fmt.Println("Error",err)
	}
}

func (um UpManager)root(w http.ResponseWriter,r * http.Request){
	if r.RequestURI == "/" {
		r.RequestURI = "html/upload.html"
	}
	fmt.Println("Load",r.RequestURI,"=>",filepath.Join(um.resourcesFolder,"resources",r.RequestURI))
	http.ServeFile(w,r,filepath.Join(um.resourcesFolder,"resources",r.RequestURI))
}

func main(){
	args := arguments.NewArguments()
	folder := args.GetString("folder")
	resources := args.GetString("resources")
	if folder == ""{
		fmt.Println("Impossible, add folder")
		return
	}
	manager := UpManager{folder:folder,resourcesFolder:resources}

	s := http.NewServeMux()
	s.HandleFunc("/upload",manager.uploadFile)
	s.HandleFunc("/launch",manager.launch)
	s.HandleFunc("/list",manager.list)
	s.HandleFunc("/cancel",manager.cancel)
	s.HandleFunc("/",manager.root)

	fmt.Println("Server well launch on 8808")
	http.ListenAndServe(":8808",s)

}
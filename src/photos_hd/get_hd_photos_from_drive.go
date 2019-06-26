package main

import "os"
import "io"
import "fmt"
import "path/filepath"
import "net/http"
import "strings"
import "regexp"
import "encoding/json"

// For each path, store the list of photos name
// Use a short path to simplify search
var foldersMap map[string][]string
var prefix string

func main(){
	fmt.Println("Started service, parse folders")
	folder := os.Args[1]
	// Prefix to remove in map
	prefix = os.Args[2]
	port := os.Args[3]
	removeFirstLevel := len(os.Args) == 5 && os.Args[4] == "true"
	foldersMap = getPaths(folder,removeFirstLevel)
	
	server := http.NewServeMux()
	server.HandleFunc("/photo",getPhoto)
	server.HandleFunc("/folders",func(w http.ResponseWriter, r * http.Request){w.Write(getFoldersAsString())})
	server.HandleFunc("/folder",getFolder)	

	fmt.Println("Started server with",len(foldersMap),"folders on port",port)
	http.ListenAndServe(":" + port,server)
}

func getFoldersAsString()[]byte{
	list := make([]string,0,len(foldersMap))
	for v,_ := range foldersMap {
		list = append(list,v)
	}
	data,_ := json.Marshal(list)
	return data
}

// Send photos of a folder
func getFolder(w http.ResponseWriter, r * http.Request){
	
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-type","application/json")
	
	if photos,exist := foldersMap[r.FormValue("folder")] ; exist {
		fmt.Println("Get info on folder",r.FormValue("folder"),"and found",len(photos),"photos")
		data,_ := json.Marshal(photos)
		w.Write(data)
	}else{
			fmt.Println("Error, impossible to found folder with name",r.FormValue("folder"))
	}
	
}

// For returning a photo, check :
//1) folder belongs to set
//2) no hacking characters exists in name like .., / or \
//3) add itself .jpg to avoid bad extensions search
func getPhoto(w http.ResponseWriter,r * http.Request){
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-type","application/json")
	
	folder := r.FormValue("folder")
	name := r.FormValue("name")
	if suffix := strings.Index(name,".jpg") ; suffix != -1 {
		name = name[0:suffix]
	}
	if folder == "" || name == "" {
		http.Error(w,"Bad request",400)
		return
	}

	if _,exist := foldersMap[folder] ; !exist {
		http.Error(w,"Access denied",403)
		return
	} 
	reg,_ := regexp.Compile("([/\\\\])|(\\.\\.)")
	if len(reg.FindAllString(name,-1)) != 0 {
		http.Error(w,"Access denied strictly",403)
		return
	}
	// Check if photo exist in hd folder
	filename := filepath.Join(prefix,folder,"hd",name+".jpg")
	fmt.Println("Show picture",filename)
	if img,err := os.Open(filename) ; err== nil {
		defer img.Close()
		w.Header().Set("Content-type","image/jpeg")
		io.Copy(w,img)
	}else{
		http.Error(w,"Image not found",404)
	}
}

// Return all valids paths with hd folder exist
func getPaths(folder string, removeFirstLevel bool)map[string][]string{
	// search if folder contains an hd folder
	if _,err := os.Stat(filepath.Join(folder,"hd")) ; err == nil || !os.IsNotExist(err) {
		// All pictures of folder
		if dirPhotos, err2 := os.Open(filepath.Join(folder,"hd")) ; err2 == nil {
			if files,err3 := dirPhotos.Readdirnames(-1) ; err3 == nil {
				photos := make([]string,0,len(files))
				for _,file := range files {
					if strings.HasSuffix(file,"jpg") {
						photos = append(photos,file)
					}
				}
				if len(photos) > 0 {
					// Remove prefix and replace \ by / to uniformize
					shortFolder := strings.Replace(strings.Replace(folder,prefix,"",-1),"\\","/",-1)
					if removeFirstLevel {
						shortFolder = shortFolder[strings.Index(shortFolder,"/")+1:]
					}
					fmt.Println(folder,"=>",shortFolder)					
					return map[string][]string{shortFolder:photos}
				}
			}
		}
		return make(map[string][]string)
	}
	// Browse all folders and relaunch on sons
	results := make(map[string][]string,0)
	if dir,err := os.Open(folder) ; err == nil {
		if files,er := dir.Readdir(-1) ; er == nil {
			for _,file := range files {
				if file.IsDir() {
					for path,photos := range getPaths(filepath.Join(folder,file.Name()),removeFirstLevel) {
						results[path] = photos
					}
				}
			}
		}
	}
	return results
}

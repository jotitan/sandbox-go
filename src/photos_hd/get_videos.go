package main

import (
	"log"
	"os"
	"io"
	"path/filepath"
	"net/http"
	"strings"
	"regexp"
	"encoding/json"
)

// Stocke pour chaque chemins la liste des noms de photo
// On utilise une version courte du path pour simplifier recherch
var foldersMap map[string][]string
var prefix string

func main(){
	log.Println("Started service, parse folders")
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

	log.Println("Started server with",len(foldersMap),"folders on port",port)
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

// Renvoie les photos du repertoire
func getFolder(w http.ResponseWriter, r * http.Request){

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-type","application/json")

	if photos,exist := foldersMap[r.FormValue("folder")] ; exist {
		log.Println("Get info on folder",r.FormValue("folder"),"and found",len(photos),"photos")
		data,_ := json.Marshal(photos)
		w.Write(data)
	}else{
		log.Println("Error, impossible to found folder with name",r.FormValue("folder"))
	}

}

// Renvoie une photo, verifie :
//1) que le repertoire est dans le set
//2) qu'aucun caractere permettant de hacker n'est present : .., /, \
//3) on ajoute .jpg a la fin
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
	// Test si la photo existe dans le repertoire hd
	filename := filepath.Join(prefix,folder,"hd",name+".jpg")
	log.Println("Show picture",filename)
	if img,err := os.Open(filename) ; err== nil {
		defer img.Close()
		w.Header().Set("Content-type","image/jpeg")
		io.Copy(w,img)
	}else{
		http.Error(w,"Image not found",404)
	}
}

func getFolders(w http.ResponseWriter, r * http.Request){
	//w.Write([]byte(strings.Join(folders,",")))
}

func listToSet(list []string)map[string]struct{}{
	set := make(map[string]struct{},len(list))
	for _,value := range list {
		set[value] = struct{}{}
	}
	return set
}

// Renvoie tous les chemins ou un repertoire hd est present
func getPaths(folder string, removeFirstLevel bool)map[string][]string{
	// Cherche si un repertoire hd est present
	if _,err := os.Stat(filepath.Join(folder,"hd")) ; err == nil || !os.IsNotExist(err) {
		// Liste les photos du repertoire
		if dirPhotos, err2 := os.Open(filepath.Join(folder,"hd")) ; err2 == nil {
			if files,err3 := dirPhotos.Readdirnames(-1) ; err3 == nil {
				photos := make([]string,0,len(files))
				for _,file := range files {
					if strings.HasSuffix(file,"jpg") {
						photos = append(photos,file)
					}
				}
				if len(photos) > 0 {
					// Supprime le prefixe et remplace \ par /
					shortFolder := strings.Replace(strings.Replace(folder,prefix,"",-1),"\\","/",-1)
					if removeFirstLevel {
						shortFolder = shortFolder[strings.Index(shortFolder,"/")+1:]
					}
					log.Println(folder,"=>",shortFolder)
					return map[string][]string{shortFolder:photos}
				}
			}
		}
		return make(map[string][]string)
	}
	// Parcourt les repertoires et relance sur les enfants
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

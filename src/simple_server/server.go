package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

var localFolder string = ""

func main(){

	if len(os.Args) > 1 {
		localFolder = os.Args[1]
	}else{
		log.Println("You can define a resources folder as first argument")
	}

	server := http.NewServeMux()

	server.HandleFunc("/help",help)
	server.HandleFunc("/hello",hello)
	server.HandleFunc("/mirror",mirror)
	server.HandleFunc("/format",format)
	server.HandleFunc("/",serveFiles)

	log.Println("Start server on port 9000 with folder",localFolder)
	http.ListenAndServe(":9000",server)

}

func help(w http.ResponseWriter,r * http.Request){
	log.Println("Launch help")
	w.Header().Set("Content-type","text/html")
	w.Write([]byte("<h2>List of services</h2><ul>" +
		"<li><b>/help</b> : display help</li>" +
		"<li><b>/hello</b> : simple hello</li>" +
		"<li><b>/mirror?value=myvalue</b> : return a message with the given value</li>" +
		"<li><b>/format</b> : return a json version of all given parameters (ex : format?f1=v1&f2=v2 => {\"f1\":\"v1\",\"f2\":\"v2\"}</li>" +
		"</ul>"))
}

func hello(w http.ResponseWriter,r * http.Request){
	log.Println("Hello")
	w.Write([]byte("Hellow world"))
}

func mirror(w http.ResponseWriter,r * http.Request){
	value := r.FormValue("value")
	log.Println("Mirror",value)
	w.Write([]byte("Receive value '" + value + "'"))
}

func format(w http.ResponseWriter,r * http.Request){
	r.ParseForm()
	log.Println("Format",r.Form)
	values := make(map[string]string,len(r.Form))
	for field,value := range r.Form {
		values[field] = value[0]
	}
	data,_ := json.Marshal(values)
	w.Header().Set("Content-type","application/json")
	w.Write(data)
}

func serveFiles(w http.ResponseWriter,r * http.Request){
	if r.URL.Path == "/"{
		help(w,r)
		return
	}
	if localFolder == "" {
		w.Write([]byte("Must define a folder as argument as startup !"))
		return
	}

	path := r.URL.Path[1:]
	log.Println("Serve file",path)
	http.ServeFile(w,r,filepath.Join(localFolder,path))
	//w.Write([]byte("Receive value '" + value + "'"))
}

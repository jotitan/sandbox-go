package photos_server

import (
	"encoding/json"
	"io"
	"logger"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)


type Server struct {
	foldersManager *foldersManager
	resources string
}

func NewPhotosServer(cache,resources string)Server{
	return Server{
		foldersManager:NewFoldersManager(cache),
		resources:resources,
	}
}

func (s Server)listFolders(w http.ResponseWriter,r * http.Request){
	names := make([]string,0,len(s.foldersManager.Folders))
	for name := range s.foldersManager.Folders {
		names = append(names,name)
	}
	if data,err := json.Marshal(names) ; err == nil{
		w.Header().Set("Content-type","application/json")
		w.Write(data)
	}
}

func (s Server)addFolder(w http.ResponseWriter,r * http.Request){
	folder := r.FormValue("folder")
	s.foldersManager.AddFolder(folder)
}

func (s Server)analyse(w http.ResponseWriter,r * http.Request){
	folder := r.FormValue("folder")
	logger.GetLogger2().Info("Analyse",folder)
	nodes := foldersManager{}.Analyse("",folder)
	if data,err := json.Marshal(nodes) ; err == nil {
		w.Header().Set("Content-type","application/json")
		w.Write(data)
	}
}

func (s Server)defaultHandle(w http.ResponseWriter,r * http.Request){
	switch {
	case strings.Index(r.URL.Path,"/browserf") == 0:
		s.browseRestful(w,r)
		break
	case strings.Index(r.URL.Path,"/browse") == 0:
		s.browse(w,r)
		break
	case strings.Index(r.URL.Path,"/image") == 0:
		s.image(w,r)
		break
	default:
		logger.GetLogger2().Info("Receive request", r.URL, r.URL.Path)
		http.ServeFile(w,r,filepath.Join(s.resources,r.RequestURI[1:]))

	}
}

func (s Server)image(w http.ResponseWriter,r * http.Request){
	path := r.URL.Path[7:]
	logger.GetLogger2().Info("Image",path)
	w.Header().Set("Content-type","image/jpeg")
	if file,err := os.Open(filepath.Join(s.foldersManager.reducer.cache,path)) ; err == nil {
		if _,e := io.Copy(w,file) ; e != nil {
			http.Error(w,"Error during image rendering",404)
		}
	}else{
		http.Error(w,"Image not found",404)
	}
}

func (s Server)browse(w http.ResponseWriter,r * http.Request){
	// Extract folder
	path := r.URL.Path[7:]
	logger.GetLogger2().Info("Browse receive request",path)
	if files,err := s.foldersManager.Browse(path) ; err == nil {
		if data,err := json.Marshal(files) ; err == nil {
			w.Header().Set("Content-type","application/json")
			w.Write(data)
		}
	}else{
		http.Error(w,err.Error(),400)
	}
}

func (s Server)update(w http.ResponseWriter,r * http.Request){
	logger.GetLogger2().Info("Launch update")
	if err := s.foldersManager.Update() ; err != nil {
		logger.GetLogger2().Error(err.Error())
	}
}

func (s Server)getRootFolders(w http.ResponseWriter,r * http.Request){
	logger.GetLogger2().Info("Get root folders")
	w.Header().Set("Access-Control-Allow-Origin","*")
	nodes := make([]*Node,0,len(s.foldersManager.Folders))
	for _,node := range s.foldersManager.Folders {
		nodes = append(nodes,node)
	}
	root := folderRestFul{Name:"Racine",Link:"",Children:s.convertPaths(nodes,true)}
	if data,err := json.Marshal(root) ; err == nil {
		w.Header().Set("Content-type","application/json")
		w.Write(data)
	}

}

func (s Server)browseRestful(w http.ResponseWriter,r * http.Request){
	// Extract folder
	path := r.URL.Path[9:]
	logger.GetLogger2().Info("Browse restfull receive request",path)
	w.Header().Set("Access-Control-Allow-Origin","*")
	if files,err := s.foldersManager.Browse(path) ; err == nil {
		formatedFiles := s.convertPaths(files,false)
		if data,err := json.Marshal(formatedFiles) ; err == nil {
			w.Header().Set("Content-type","application/json")
			w.Write(data)
		}
	}else{
		http.Error(w,err.Error(),400)
	}
}

// Restful representation : real link instead real path
type imageRestFul struct{
	Name string
	ThumbnailLink string
	ImageLink string
	Width int
	Height int
}

type folderRestFul struct{
	Name string
	Link string
	Children []interface{}
}

// Convert node to restful response
func (s Server)convertPaths(nodes []*Node,onlyFolders bool)[]interface{}{
	files := make([]interface{},0,len(nodes))
	for _,node := range nodes {
		if !node.IsFolder {
			if !onlyFolders {
				files = append(files, imageRestFul{Name: node.Name, Width: node.Width, Height: node.Height,
					ThumbnailLink: filepath.ToSlash(filepath.Join("/image", s.foldersManager.GetSmallImageName(*node))),
					ImageLink:     filepath.ToSlash(filepath.Join("/image", s.foldersManager.GetMiddleImageName(*node)))})
			}
		}else{
			folder := folderRestFul{Name:node.Name,Link:filepath.ToSlash(filepath.Join("/browserf",node.RelativePath))}
			if onlyFolders {
				// RElaunch on subfolders
				subNodes := make([]*Node,0,len(node.Files))
				for _,n := range node.Files {
					subNodes = append(subNodes,n)
				}
				childrens :=s.convertPaths(subNodes,onlyFolders)
				folder.Children = childrens
			}
			files = append(files,folder)
		}
	}
	return files
}

func (s Server)Launch(){

	server := http.ServeMux{}
	server.HandleFunc("/analyse",s.analyse)
	server.HandleFunc("/addFolder",s.addFolder)
	server.HandleFunc("/rootFolders",s.getRootFolders)
	server.HandleFunc("/update",s.update)
	server.HandleFunc("/listFolders",s.listFolders)
	server.HandleFunc("/",s.defaultHandle)

	logger.GetLogger2().Info("Start server on port 9006")
	http.ListenAndServe(":9006",&server)
}
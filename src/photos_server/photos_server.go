package photos_server

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"logger"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)


type Server struct {
	foldersManager *foldersManager
	resources string
}

func NewPhotosServer(cache,resources,garbage,maskDelete string)Server{
	return Server{
		foldersManager:NewFoldersManager(cache,garbage,maskDelete),
		resources:resources,
	}
}

func (s Server)canDelete(w http.ResponseWriter,r * http.Request){
	w.Header().Set("Access-Control-Allow-Origin","*")
	w.Header().Set("Content-type","application/json")
	canDelete := s.foldersManager.garbageManager != nil && s.foldersManager.garbageManager.CanDelete(r.Referer())
	w.Write([]byte(fmt.Sprintf("{\"can\":%t}",canDelete)))
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
	forceRotate := r.FormValue("forceRotate") == "true"
	logger.GetLogger2().Info("Add folder",folder,"and forceRotate :",forceRotate)
	s.foldersManager.AddFolder(folder,forceRotate)
}

func (s Server)delete(w http.ResponseWriter,r * http.Request){
	w.Header().Set("Access-Control-Allow-Origin","*")
	if s.foldersManager.garbageManager == nil || !s.foldersManager.garbageManager.CanDelete(r.Referer()){
		logger.GetLogger2().Info("Try to delete by",r.Referer())
		http.Error(w,"You can't execute this action",403)
		return
	}
	data,_ := ioutil.ReadAll(r.Body)
	deletions := make([]string,0)
	json.Unmarshal(data,&deletions)
	imagesPath := make([]string,len(deletions))
	for i,deletion := range deletions {
		imagesPath[i] = strings.ReplaceAll(deletion,"/imagehd/","")
	}
	successDeletions := s.foldersManager.garbageManager.Remove(imagesPath)
	w.Write([]byte(fmt.Sprintf("{\"success\":%d,\"errors\":%d}",successDeletions,len(imagesPath)-successDeletions)))
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
	case strings.Index(r.URL.Path,"/imagehd") == 0:
		s.imageHD(w,r)
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
	s.writeImage(w,filepath.Join(s.foldersManager.reducer.cache,path))
}

func (s Server)writeImage(w http.ResponseWriter,path string){
	w.Header().Set("Content-type","image/jpeg")
	if file,err := os.Open(path) ; err == nil {
		defer file.Close()
		if _,e := io.Copy(w,file) ; e != nil {
			http.Error(w,"Error during image rendering",404)
		}
	}else{
		http.Error(w,"Image not found",404)
	}
}

// Return original image
func (s Server)imageHD(w http.ResponseWriter,r * http.Request){
	path := r.URL.Path[9:]
	// Find absolute path based on first folder
	baseDir := strings.Split(path,"/")[0]
	if folder,exist := s.foldersManager.Folders[baseDir] ; !exist {
		http.Error(w,"Image folder hd not found",404)
	}else{
		imgPath :=filepath.Join(folder.AbsolutePath,filepath.Join(strings.Split(path,"/")[1:]...))
		s.writeImage(w,imgPath)
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

// Update a specific folder, faster than all folders
func (s Server)updateFolder(w http.ResponseWriter,r * http.Request){
	w.Header().Set("Access-Control-Allow-Origin","*")
	folder := r.FormValue("folder")
	logger.GetLogger2().Info("Launch update folder :",folder)
	if err := s.foldersManager.UpdateFolder(folder) ; err != nil {
		logger.GetLogger2().Error(err.Error())
	}else{
		logger.GetLogger2().Info("End update folder",folder)
		w.Write([]byte("success"))
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
	w.Header().Set("Access-Control-Allow-Origin","*")
	path := r.URL.Path[9:]
	logger.GetLogger2().Info("Browse restfull receive request",path)
	if files,err := s.foldersManager.Browse(path) ; err == nil {
		formatedFiles := s.convertPaths(files,false)
		imgResponse := imagesResponse{Files:formatedFiles,UpdateUrl:"/updateFolder?folder=" + path[1:]}
		if data,err := json.Marshal(imgResponse) ; err == nil {
			w.Header().Set("Content-type","application/json")
			w.Write(data)
		}
	}else{
		logger.GetLogger2().Info("Impossible to browse",path,err.Error())
		http.Error(w,err.Error(),400)
	}
}

type imagesResponse struct {
	Files []interface{}
	UpdateUrl string
}

// Restful representation : real link instead real path
type imageRestFul struct{
	Name string
	ThumbnailLink string
	ImageLink string
	HdLink string
	Width int
	Height int
	Date time.Time
	Orientation int
}

type folderRestFul struct{
	Name string
	Link string
	// Means that folder also have images to display
	HasImages bool
	Children []interface{}
}

// Convert node to restful response
func (s Server)convertPaths(nodes []*Node,onlyFolders bool)[]interface{}{
	files := make([]interface{},0,len(nodes))
	for _,node := range nodes {
		if !node.IsFolder {
			if !onlyFolders {
				files = append(files, imageRestFul{
					Name: node.Name, Width: node.Width, Height: node.Height,Date:node.Date,
					HdLink:filepath.ToSlash(filepath.Join("/imagehd",node.RelativePath)),
					ThumbnailLink: filepath.ToSlash(filepath.Join("/image", s.foldersManager.GetSmallImageName(*node))),
					ImageLink:     filepath.ToSlash(filepath.Join("/image", s.foldersManager.GetMiddleImageName(*node)))})
			}
		}else{
			folder := folderRestFul{Name:node.Name,Link:filepath.ToSlash(filepath.Join("/browserf",node.RelativePath))}
			if onlyFolders {
				// Relaunch on subfolders
				subNodes := make([]*Node,0,len(node.Files))
				hasImages := false
				for _,n := range node.Files {
					subNodes = append(subNodes,n)
					if !n.IsFolder {
						hasImages = true
					}
				}
				childrens :=s.convertPaths(subNodes,onlyFolders)
				folder.Children = childrens
				folder.HasImages = hasImages
			}
			files = append(files,folder)
		}
	}
	return files
}

func (s Server)Launch(port string){
	//test()
	server := http.ServeMux{}
	server.HandleFunc("/analyse",s.analyse)
	server.HandleFunc("/delete",s.delete)
	server.HandleFunc("/addFolder",s.addFolder)
	server.HandleFunc("/rootFolders",s.getRootFolders)
	server.HandleFunc("/update",s.update)
	server.HandleFunc("/updateFolder",s.updateFolder)
	server.HandleFunc("/listFolders",s.listFolders)
	server.HandleFunc("/canDelete",s.canDelete)
	server.HandleFunc("/",s.defaultHandle)

	logger.GetLogger2().Info("Start server on port " + port)
	err := http.ListenAndServe(":" + port,&server)
	logger.GetLogger2().Error("Server stopped cause",err)
}
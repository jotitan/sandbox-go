package photos_server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"logger"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type INode interface {
	IsDir()bool
}

type Files map[string]*Node

type Node struct {
	AbsolutePath string
	// Path of node relative to head
	RelativePath string
	Width int
	Height int
	Name string
	IsFolder bool
	// Store files in a map with name
	Files Files
	ImagesResized bool
}

func (n Node)applyOnEach(rootFolder string, fct func(path,relativePath string, node * Node)){
	for _,file := range n.Files{
		if file.IsFolder{
			file.applyOnEach(rootFolder,fct)
		}else{
			fct(file.AbsolutePath,file.RelativePath,file)
		}
	}
}

func (n Node)String()string{
	return fmt.Sprintf("%s : %s : %t : %s : %s",n.AbsolutePath,n.RelativePath,n.ImagesResized,n.Name,n.Files)
}

func NewImage(rootFolder,path,name string)*Node{
	relativePath := strings.Replace(path,rootFolder,"",-1)
	return &Node{AbsolutePath:path,RelativePath:relativePath,Name:name,IsFolder:false,Files:nil,ImagesResized:false}
}

func NewFolder(rootFolder,path,name string,files Files, imageResized bool)*Node{
	relativePath := strings.Replace(path,rootFolder,"",-1)
	return &Node{AbsolutePath:path,RelativePath:relativePath,Name:name,IsFolder:true,Files:files,ImagesResized:imageResized}
}

// Store many folders
type foldersManager struct{
	Folders map[string]*Node
	reducer Reducer
}

func NewFoldersManager(cache string)*foldersManager{
	fm := &foldersManager{Folders:make(map[string]*Node,0),reducer:NewReducer(cache,[]uint{1080,250})}
	fm.load()
	return fm
}

func (fm foldersManager) GetSmallImageName(node Node)string{
	return fm.reducer.createJpegFile(filepath.Dir(node.RelativePath),node.RelativePath,fm.reducer.sizes[1])
}

func (fm foldersManager) GetMiddleImageName(node Node)string{
	return fm.reducer.createJpegFile(filepath.Dir(node.RelativePath),node.RelativePath,fm.reducer.sizes[0])
}

var extensions = []string{"jpg","jpeg","png"}

// Compare old and new version of folder
// For each files in new version : search if old version exist, if true, keep information, otherwise, store new node in separate list
// To detect deletion, create a copy at beginning and remove element at each turn
func (files Files)Compare(previousFiles Files)([]*Node,map[string]*Node){
	//fmt.Println("Compare",files,previousFiles)
	newNodes := make([]*Node,0)
	nodesToDelete := make(map[string]*Node,0)
	// First recopy old version
	deletionMap := make(map[string]*Node,len(previousFiles))
	for name,node := range previousFiles {
		deletionMap[name] = node
	}
	for name,file := range files {
		if oldValue, exist := previousFiles[name]; exist {
			// Remove element from deletionMap
			delete(deletionMap,name)
			if !file.IsFolder {
				file.Height = oldValue.Height
				file.Width = oldValue.Width
				file.ImagesResized = oldValue.ImagesResized
			}else{
				// Relaunch on folder
				delta,deletions := file.Files.Compare(oldValue.Files)
				for _,node := range delta {
					newNodes = append(newNodes,node)
				}
				for name,node := range deletions {
					nodesToDelete[name] = node
				}

			}
		}else{
			// Treat folder
			if !file.IsFolder {
				newNodes = append(newNodes,file)
			}else{
				delta,deletions := file.Files.Compare(Files{})
				for _,node := range delta {
					newNodes = append(newNodes,node)
				}
				for name,node := range deletions {
					nodesToDelete[name] = node
				}
			}
		}
	}
	// Add local nodes to delete with other
	for name,node := range deletionMap {
		nodesToDelete[name] = node
	}
	return newNodes,nodesToDelete
}

// Add a locker to check if an update is running
var updateLocker = sync.Mutex{}
//Update : compare structure in memory and folder to detect changes
func (fm * foldersManager)Update()error{
	// Have to block to avoid second update in same time
	// A lock is blocking, so use a chanel tiomeout to check if locker is still block
	updateWaiter := sync.WaitGroup{}
	updateWaiter.Add(1)
	begin := time.Now()
	chanelRunning := make(chan struct{},1)
	go func() {
		// Is still block after one second, method exit and go routine is never executed
		updateLocker.Lock()
		chanelRunning <- struct{}{}
		// Stop the thread when get the lock after stop
		if time.Now().Sub(begin) > time.Duration(100)*time.Millisecond {
			return
		}
		time.Sleep(time.Second)
		// For each folder, launch an analyse and detect differences
		newFolders := make(map[string]*Node, len(fm.Folders))
		for _, folder := range fm.Folders {
			rootFolder := filepath.Dir(folder.AbsolutePath)
			files := fm.Analyse(rootFolder, folder.AbsolutePath)
			// Include dry run and full (compare length a nodes or compare always everything)
			delta, deletions := files.Compare(fm.Folders)
			logger.GetLogger2().Info("After update", len(delta), "new pictures and", len(deletions), "to remove")
			// Launch indexation of new images,
			if len(delta) > 0 {
				waiter := &sync.WaitGroup{}
				for _, node := range delta {
					logger.GetLogger2().Info("Launch update image resize", node.AbsolutePath)
					waiter.Add(1)
					fm.reducer.AddImage(node.AbsolutePath, node.RelativePath, node, waiter)
				}
				waiter.Wait()
				logger.GetLogger2().Info("All pictures have been resized")
			}

			// remove deletions in cache
			fm.removeFiles(deletions)
			for name, f := range files {
				newFolders[name] = f
			}
		}
		fm.Folders = newFolders
		fm.save()
		updateWaiter.Done()
		updateLocker.Unlock()
	}()

	// Dectect if an update is already running. Is true (after one secon), return error, otherwise, wait for end of update
	select {
	case <- chanelRunning :
		updateWaiter.Wait()
		return nil
	case <- time.NewTimer(time.Millisecond*100).C:
		return errors.New("An update is already running")
	}
}

func (fm foldersManager)removeFiles(files map[string]*Node){
	for _,node := range files {
		fm.removeFile(filepath.Join(fm.reducer.cache,fm.GetSmallImageName(*node)))
		fm.removeFile(filepath.Join(fm.reducer.cache,fm.GetMiddleImageName(*node)))
	}
}

func (fm foldersManager)removeFile(path string){
	logger.GetLogger2().Info("Remove file",path)
	os.Remove(path)
}

func (fm * foldersManager)AddFolder(folderPath string){
	rootFolder := filepath.Dir(folderPath)
	node := fm.Analyse(rootFolder,folderPath)
	logger.GetLogger2().Info("Add folder",folderPath,"with",len(node))
	for name,folder := range node{
		fm.Folders[name] = folder
		fm.launchImageResize(folder,strings.Replace(folderPath,name,"",-1))
	}
	//
	fm.save()
}

func (fm * foldersManager)load(){
	if f,err := os.Open("save-images.json") ; err == nil {
		defer f.Close()
		data,_ := ioutil.ReadAll(f)
		folders := make(map[string]*Node,0)
		json.Unmarshal(data,&folders)
		fm.Folders = folders
	}
}

func (fm foldersManager)save(){
	data,_ := json.Marshal(fm.Folders)
	if f,err := os.OpenFile("save-images.json",os.O_TRUNC|os.O_CREATE|os.O_RDWR,os.ModePerm) ; err == nil {
		defer f.Close()
		f.Write(data)
		logger.GetLogger2().Info("Save tree in file")
	}
}

func (fm * foldersManager)launchImageResize(folder *Node, rootFolder string){
	waiter := &sync.WaitGroup{}
	folder.applyOnEach(rootFolder,func(path,relativePath string,node * Node){
		waiter.Add(1)
		fm.reducer.AddImage(path,relativePath,node,waiter)})
	go func(w *sync.WaitGroup,node *Node){
		w.Wait()
		logger.GetLogger2().Info("End of resize folder")
		node.ImagesResized=true
	}(waiter,folder)
}

//Analyse a cache and detect all files of types images
func (fm foldersManager)Analyse(rootFolder,path string)Files{
	if file,err := os.Open(path) ; err == nil{
		defer file.Close()
		// If cache, create cache and go deep
		if stat,errStat := file.Stat() ; errStat == nil {
			if stat.IsDir() {
				files,_ := file.Readdirnames(-1)
				nodes := make(map[string]*Node,0)
				for _,file := range files {
					for name,node := range fm.Analyse(rootFolder,filepath.Join(path,file)){
						nodes[name] = node
					}
				}
				if len(nodes) > 0 {
					r := make(map[string]*Node,0)
					r[stat.Name()] = NewFolder(rootFolder,path, stat.Name(), nodes,false)
					return r
				}
			}else{
				// Test if is image
				if isImage(stat.Name()){
					r := make(map[string]*Node,0)
					r[stat.Name()] = NewImage(rootFolder,path, stat.Name())
					return r
				}
			}
		}
	}else{
		logger.GetLogger2().Error(err.Error() + " : " + rootFolder + " ; " + path)
	}
	return Files{}
}

func (fm foldersManager)List()[]*Node{
	nodes := make([]*Node,0,len(fm.Folders))
	for name,folder := range fm.Folders{
		nodes = append(nodes,NewFolder("",name,name,nil,folder.ImagesResized))
	}
	return nodes
}

func (fm *foldersManager) Browse(path string) ([]*Node,error){
	if len(path) < 2 {
		// Return list
		return fm.List(),nil

	}else{
		var node *Node
		var exist bool
		// Browse path
		for i,folder := range strings.Split(path[1:],"/") {
			if i == 0 {
				if node,exist = fm.Folders[folder] ; !exist {
					return nil,errors.New("Invalid path " + folder)
				}
			}else{
				if !strings.EqualFold("",strings.Trim(folder," ")) {
					if !node.IsFolder {
						return nil, errors.New("Not a valid cache " + folder)
					}
					node = node.Files[folder]
				}
			}
		}
		// Parse file of nodes
		nodes := make([]*Node,0,len(node.Files))
		for _,file := range node.Files {
			nodes = append(nodes,file)
		}
		return nodes,nil
	}
}

func isImage(name string)bool{
	for _,suffix := range extensions {
		if  strings.HasSuffix(strings.ToLower(name),suffix){
			return true
		}
	}
	return false
}
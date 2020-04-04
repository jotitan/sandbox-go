package photos_server

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"resize"
	"sync"
)

//Manage reducing pictures

type Reducer struct {
	// Where reduced images are created
	cache string
	// Differents sizes to produce
	sizes []uint
	// Receive an absolute path of image and a relative path to cache
	imagesToResize chan ImageToResize
	resize resize.GoResizer
	totalCount int
}

func NewReducer(folder string, sizes []uint)Reducer{
	r := Reducer{cache: folder,sizes:sizes,resize:resize.GoResizer{},imagesToResize:make(chan ImageToResize,100)}
	go r.listenAndResize()
	return r
}

type ImageToResize struct{
	path string
	relativePath string
	node * Node
	waiter * sync.WaitGroup
}

func (r Reducer)AddImage(path,relativePath string,node * Node,waiter * sync.WaitGroup){
	r.imagesToResize <- ImageToResize{path,relativePath,node,waiter}
}

// Return number of images wating to reduce and number of images reduced
func (r * Reducer)Stat()(int,int){
	return len(r.imagesToResize),r.totalCount
}

func (r * Reducer)listenAndResize(){
	go func(){
		for {
			imageToResize := <-r.imagesToResize
			r.totalCount++
			targetFolder := filepath.Dir(imageToResize.relativePath)
			folder := filepath.Join(r.cache, targetFolder)
			if r.createPathInCache(folder) == nil {
				r.resizeMultiformat(imageToResize,folder)
			}
		}
	}()
}

func (r Reducer) resizeMultiformat(imageToResize ImageToResize,folder string){
	// Reuse computed image to accelerate
	from := imageToResize.path
	for i, size := range r.sizes {
		to := r.createJpegFile(folder,imageToResize.path,size)
		_,width,height := r.resize.Resize(from, to, 0, size)
		if i == len(r.sizes) -1 {
			// Set ratio on node
			imageToResize.node.Height = int(height)
			imageToResize.node.Width = int(width)
		}
		from = to
	}
	imageToResize.node.ImagesResized = true
	imageToResize.waiter.Done()
}

func (r Reducer)createPathInCache(path string)error{
	if f,err := os.Open(path) ; err != nil {
		// Create folder
		return os.MkdirAll(path,os.ModePerm)
	}else{
		if stat,err := f.Stat() ; err != nil || !stat.IsDir(){
			return errors.New("Impossible to use this folder : "  + path)
		}
	}
	return nil
}

func (r Reducer)createJpegFile(folder, basePath string, size uint)string{
	return filepath.Join(folder, r.CreateJpegName(filepath.Base(basePath), size))
}

// Generate a jpeg name from size
func (r Reducer)CreateJpegName(name string, size uint)string{
	extension := filepath.Ext(name)
	baseName := name[:len(name) - len(extension)]
	return fmt.Sprintf("%s-%d%s",baseName,size,".jpg")
}
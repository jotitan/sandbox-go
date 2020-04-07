package photos_server

import (
	"errors"
	"fmt"
	"logger"
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
	resize resize.AsyncGoResizer
	totalCount int
}

func NewReducer(folder string, sizes []uint)Reducer{
	r := Reducer{cache: folder,sizes:sizes,resize:resize.NewAsyncGoResize(),imagesToResize:make(chan ImageToResize,100)}
	go r.listenAndResize()
	return r
}

type ImageToResize struct{
	path string
	relativePath string
	node * Node
	waiter * sync.WaitGroup
	existings map[string]struct{}
}

func (itr ImageToResize)update(h,w uint){
	itr.node.Height = int(h)
	itr.node.Width = int(w)
	itr.node.ImagesResized = true
	itr.waiter.Done()
}

func (r Reducer)AddImage(path,relativePath string,node * Node,waiter * sync.WaitGroup, existings map[string]struct{}){
	r.imagesToResize <- ImageToResize{path,relativePath,node,waiter,existings}
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
	conversions := make([]resize.ImageToResize,len(r.sizes))
	// Check if both exist, if true, return, otherwise, resize
	nbExist := 0
	for i, size := range r.sizes {
		conversions[i] = resize.ImageToResize{To:r.createJpegFile(folder,imageToResize.path,size),Width:0,Height:size}
		if _,exist := imageToResize.existings[conversions[i].To]; exist {
			nbExist++
		}
	}
	if nbExist == len(r.sizes){
		// All exist, get Size of little one and return
		w,h := resize.GetSize(conversions[len(conversions)-1].To)
		imageToResize.update(h,w)
		return
	}
	callback := func(err error,width,height uint){
		if err != nil {
			logger.GetLogger2().Info("Got error on resize",err)
		}else{
			if width != 0 && height != 0 {
				imageToResize.update(height,width)
			}
		}
	}
	r.resize.ResizeAsync(from,conversions,callback)
}

func (r Reducer)createPathInCache(path string)error{
	if f,err := os.Open(path) ; err != nil {
		// Create folder
		return os.MkdirAll(path,os.ModePerm)
	}else{
		defer f.Close()
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
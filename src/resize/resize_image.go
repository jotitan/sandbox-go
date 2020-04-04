package resize

import (
	"errors"
	"fmt"
	resizer "github.com/nfnt/resize"
	"image"
	"image/jpeg"
	"image/png"
	"logger"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Request struct {
	from string
	to string
}

func GetPhotosToTreat(from,to string)[]Request{
	var requests []Request
	if file,err := os.Open(from) ; err == nil {
		if info,_ := file.Stat() ; info.IsDir() {
			// List jpg file
			names,_ := file.Readdirnames(0)
			for _,name := range names {
				if strings.HasSuffix(name,".jpg") {
					from := filepath.Join(from,name)
					to := fmt.Sprintf("%s_%s", to, name)
					requests = append(requests,Request{from,to})
				}
			}
		}else{
			requests = []Request{Request{from,to}}
		}
	}
	return requests
}

func ResizeMany(from, to string, width,height uint){
	counter := sync.WaitGroup{}
	begin := time.Now()
	gor := GetResizer()

	requests := GetPhotosToTreat(from,to)
	for _,r := range requests {
		counter.Add(1)
		go func(req Request) {
			if err,_,_ := gor.Resize(req.from, req.to, width,height); err == nil {
				fmt.Println("Img resized", req.to)
			}else {
				fmt.Println("Impossible",err)
			}
			counter.Done()
		}(r)
	}
	counter.Wait()
	fmt.Println("Done",time.Now().Sub(begin))
}

type Resizer interface{
	Resize(from,to string,width,height uint)(error,uint,uint)
	ToString()string
}

// GetResizer return a resizer acoording to context
func GetResizer()Resizer{
	// CHeck if convert (imagemagick) is present
	if result,err := exec.Command("convert","-version") .Output() ; err == nil {
		if strings.Contains(string(result),"ImageMagick") {
			return ImageMagickResizer{}
		}
	}
	return GoResizer{}
}

// ImageMagickResizer use image magick (convert command) to compress
type ImageMagickResizer struct{}

func (gor ImageMagickResizer)ToString()string{
	return "ImageMagick"
}

func (gor ImageMagickResizer)Resize(from,to string,width,height uint)(error,uint,uint){
	cmd := exec.Command("convert",from,"-resize",fmt.Sprintf("x%d",height),"-auto-orient","-interpolate","bicubic","-quality","80",to)
	_,err := cmd.Output()

	return err,0,0
}

type GoResizer struct{}

func (gor GoResizer)ToString()string{
	return "Go Resizer"
}

func (gor GoResizer)Resize(from,to string,width,height uint)(error,uint,uint){
	// Check if image already exist
	if f,err := os.Open(to) ; err == nil {
		// Already exist, close and return, open the light one to get size ? Return 0,0 for now
		f.Close()
		return nil,0,0
	}

	if img,err := openImage(from) ; err == nil {
		//fmt.Println("Time read",time.Now().Sub(begin))
		imgResize,w,h := resizeImage(img, width, height)
		//fmt.Println("Time resize",time.Now().Sub(begin))
		return saveImage(imgResize, to),w,h
		//fmt.Println("Time save",time.Now().Sub(begin))
	}else{
		logger.GetLogger2().Info("Impossible to resize",err)
		return err,0,0
	}
}

func saveImage(img image.Image, path string)error{
	if f,err := os.OpenFile(path,os.O_CREATE|os.O_TRUNC|os.O_RDWR,os.ModePerm) ; err == nil{
		defer f.Close()
		return jpeg.Encode(f,img,&(jpeg.Options{75}))
	}else{
		return err
	}
}

func openImage(path string)(image.Image,error) {
	if f,err := os.Open(path) ; err == nil{
		defer f.Close()
		var img image.Image
		var err2 error
		ext := strings.ToLower(filepath.Ext(path))
		switch {
		case strings.EqualFold(ext,".jpg") || strings.EqualFold(ext,".jpeg") :
			img,err2 = jpeg.Decode(f)
			break
		case strings.EqualFold(ext,".png") :
			img,err2 = png.Decode(f)
			break
		default:err2 = errors.New("unknown format")
		}
		if err2 == nil {
			return img,nil
		}else{
			return nil,err2
		}
	}else {
		return nil, err
	}
}

func resizeImage(img image.Image,width,height uint)(image.Image,uint,uint){
	x,y := float32(img.Bounds().Size().X),float32(img.Bounds().Size().Y)
	if float32(height) > y || float32(width) > x{
		return img,uint(x),uint(y)
	}
	switch {
	case width == 0 && height == 0 : return img,uint(x),uint(y)
	case width == 0 : width = uint((float32(height) / y) * x)
	case height == 0 : height = uint((float32(width) / x) * y)
	}
	logger.GetLogger2().Info("=>",width,height)
	return resizer.Resize(width,height,img,resizer.Bicubic),width,height
}

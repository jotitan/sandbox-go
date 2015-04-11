package resize
import (
    "image"
    resizer "github.com/nfnt/resize"
    "image/jpeg"
    "os"
)


func Resize(from,to string,width,height uint)error{
    //begin := time.Now()
    if img,err := openImage(from) ; err == nil {
        //fmt.Println("Time read",time.Now().Sub(begin))
        img = resizeImage(img, width, height)
        //fmt.Println("Time resize",time.Now().Sub(begin))
        saveImage(img, to)
        //fmt.Println("Time save",time.Now().Sub(begin))
        return nil
    }else{
        return err
    }
}

func saveImage(img image.Image, path string)error{
    if f,err := os.OpenFile(path,os.O_CREATE|os.O_TRUNC,os.ModePerm) ; err == nil{
        jpeg.Encode(f,img,&(jpeg.Options{75}))
        return nil
    }else{
        return err
    }
}

func openImage(path string)(image.Image,error) {
    if f,err := os.Open(path) ; err == nil{
        if img,err2 := jpeg.Decode(f) ; err2 == nil {
            return img,nil
        }else{
            return nil,err2
        }
    }else {
        return nil, err
    }
}

func resizeImage(img image.Image,width,height uint)image.Image{
    switch {
        case width == 0 && height == 0 : return img
        case width == 0 : width = (height / uint(img.Bounds().Size().Y)) * uint(img.Bounds().Size().X)
        case height == 0 : height = (width / uint(img.Bounds().Size().X)) * uint(img.Bounds().Size().Y)
    }
    return resizer.Resize(width,height,img,resizer.Bicubic)
}
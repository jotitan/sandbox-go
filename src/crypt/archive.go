package crypt

import (
	"os"
	"fmt"
	"archive/zip"
	"path/filepath"
	"io/ioutil"
	"strings"
)

// Create an archive with all informations



// Archive folder in a file. Use zip or lz4
func Archive(folder,path string,excludesExt map[string]struct{}){

	out,_ := os.OpenFile(path,os.O_CREATE|os.O_TRUNC,os.ModePerm)
	zipWriter := zip.NewWriter(out)

	if excludesExt == nil {
		excludesExt = make(map[string]struct{})
	}

	addFileToZip(zipWriter,folder,"","",excludesExt)

	zipWriter.Close()

}

func addFileToZip(zipWriter *zip.Writer,absoluteRoot,zipRoot,filename string,excludesExt map[string]struct{}){
	absolutePath := filepath.Join(absoluteRoot,filename)
	if strings.HasPrefix(filename,"."){
		return
	}
	file,_ := os.Open(absolutePath)
	if info,_ := file.Stat() ; info.IsDir() {
		zipRoot = filepath.Join(zipRoot,filename)
		files,_ := file.Readdirnames(0)
		for _,file := range files {
			addFileToZip(zipWriter,absolutePath,zipRoot,file,excludesExt)
		}
	}else{
		ext := filename[strings.LastIndex(filename,".")+1:]
		if _,ok := excludesExt[ext] ; !ok {
			fmt.Println("Add file", filepath.Join(zipRoot, filename))
			w, _ := zipWriter.Create(filepath.Join(zipRoot, filename))
			data, _ := ioutil.ReadAll(file)
			w.Write(data)
		}
	}
}

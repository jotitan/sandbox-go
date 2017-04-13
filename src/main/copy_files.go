package main

import (
	"os"
	"fmt"
	"path/filepath"
	"regexp"
	"io"
)

/* Use to copy files from a folder to an other */


func main(){
	folderIn := os.Args[1]
	folderOut := os.Args[2]
	pattern := os.Args[3]
	oldFolder := os.Args[4]
	copySubFolder := os.Args[5]

	reg,_ := regexp.Compile(pattern)

	if file,err := os.Open(folderIn) ; err == nil {
		dirs, _ := file.Readdir(-1)
		for _,dir := range dirs {
			if dir.IsDir() {
				if folder,err := os.Open(filepath.Join(folderIn,dir.Name())) ; err == nil {
					files, _ := folder.Readdir(-1)
					nb, total := 0, len(files)
					toCopy := make([]string, 0)
					for _, fileInfo := range files {
						if reg.Match([]byte(fileInfo.Name())) {
							toCopy = append(toCopy, fileInfo.Name())
							nb++
						}
						total++
					}
					if nb > 0 {
						if _, err := os.Open(filepath.Join(folderOut, dir.Name(), copySubFolder)); err != nil {
							// move content if oldFolder and create copySubFolder
							if outDir, err := os.Open(filepath.Join(folderOut, dir.Name())); err == nil {
								outFiles, _ := outDir.Readdir(-1)
								os.MkdirAll(filepath.Join(folderOut,dir.Name(),oldFolder),os.ModePerm)
								fmt.Println("Create folder", filepath.Join(folderOut, dir.Name(), oldFolder))
								for _, outFile := range outFiles {
									oldOutName := filepath.Join(folderOut, dir.Name(), oldFolder, outFile.Name())
									oldOutFile,_ :=os.OpenFile(oldOutName,os.O_CREATE|os.O_RDWR,os.ModePerm)
									outNameFile := filepath.Join(folderOut, dir.Name(), outFile.Name())
									outFile,_ := os.Open(outNameFile)
									fmt.Println("Deplace ", outNameFile, "vers", oldOutName)
									io.Copy(oldOutFile,outFile)
									oldOutFile.Close()
									outFile.Close()
									os.Remove(outNameFile)
								}
							}else {
								os.MkdirAll(filepath.Join(folderOut,dir.Name(),copySubFolder),os.ModePerm)
								fmt.Println("Create folder", filepath.Join(folderOut, dir.Name()))
							}
							// check if distant folder has hd folder
							for _, name := range toCopy {
								// Copy
								newOutFile, _ := os.OpenFile(filepath.Join(folderOut, dir.Name(), copySubFolder, name), os.O_CREATE | os.O_RDWR, os.ModePerm)
								inputFile, _ := os.Open(filepath.Join(folderIn, dir.Name(), name))
								io.Copy(newOutFile, inputFile)
								newOutFile.Close()
								inputFile.Close()
							}
							fmt.Println("Copy", folder.Name(), nb, "on", total)
						}

					}

				}
			}
		}
	}
}


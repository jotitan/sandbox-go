package music
import (
    "os"
    "logger"
    "path/filepath"
    "github.com/mjibson/id3"
    "encoding/json"
    "io"
	"strings"
	"regexp"
	"strconv"
	"fmt"
	"errors"
	"os/exec"
)

/* Give methods to browse musics in a specific directory */


//var folder = "D:\\Documents\\Projets\\data"
var folder = "C:\\tmp\\zik"

const (
	limitMusicFile = 10
)

// Browse a folder to get all data
func Browse(folderName string){
    dictionnary := LoadDictionnary()
	dictionnary.browseFolder(folderName)
	dictionnary.Save()
}

func (md * MusicDictionnary)browseFolder(folderName string){
	if folder,err := os.Open(folderName) ; err == nil {
		defer folder.Close()
		// List all files
		files,_ := folder.Readdir(-1)
		for _,file := range files {
			path := filepath.Join(folderName,file.Name())
			if file.IsDir() {
				logger.GetLogger().Info("Parse",path)
				md.browseFolder(path)
			} else{
				// TODO : check in bloomfilter if exist. If true, final check in dictionnary
				if strings.HasSuffix(file.Name(),".mp3") {
					md.Add(*extractInfo(path))
				}
			}
		}
	}
}


type MusicDictionnary struct{
    previousSize int
    // Data in json format
    musics []Music
    // Used to store header to write. List of position
    header []int64
    // Nb of element when opening file
    currentRead int
	// If file is full, change directory
	changeFolder bool
	// Next id for file
	nextId int64
}

func (md MusicDictionnary)currentSize()int{
    return md.previousSize + len(md.musics)
}

type Music struct{
	file id3.File
	id int64
}

func (m Music)toJSON()[]byte{
	data := make(map[string]string,7)
	data["title"] = m.file.Name
	data["artist"] = m.file.Artist
	data["album"] = m.file.Album
	data["length"] = m.file.Length
	data["year"] = m.file.Year
	data["genre"] = m.file.Genre
	data["track"] = m.file.Track

	jsonData,_ := json.Marshal(data)
	return jsonData
}


// return err
func (md MusicDictionnary)findLastFile()(int64,error){
	pattern := "music_([0-9]+).dico"
	r,_ := regexp.Compile(pattern)
	max := int64(-1)
	filesFolder,_ := os.Open(folder)
	names,_ := filesFolder.Readdirnames(-1)
	for _,name := range names {
		if result := r.FindStringSubmatch(name) ; len(result) >1 {
			if id,err := strconv.ParseInt(result[1],10,32) ; err == nil && id > max {
				max = id
			}
		}
	}
	if max == -1 {
		return 0,errors.New("No dictionnary yet")
	}
	return max,nil
}

// Save the file. Header with fix size (fixe number of element, 10000 for example). Header link position of element.
// nb element (8) | pos 1 (8) | pos 2 (8) | ... | dataLength 1 (4) | data 1 (n) | ...
func (md * MusicDictionnary)Save(){
    // Save in the file. Create data in buffer instead of write everytime
    fileId,_ := md.findLastFile()
	// Get next file
	if md.changeFolder {
		fileId++
		md.changeFolder = false
	}
	path := filepath.Join(folder,fmt.Sprintf("music_%d.dico",fileId))
	logger.GetLogger().Info("Save in file",path)
    f,err := os.OpenFile(path,os.O_CREATE|os.O_RDWR|os.O_EXCL,os.ModePerm)
    // If exist, just append result at the end
    md.header = make([]int64,0,len(md.musics))
    headerPos := int64(8)
	totalElements := int64(0)
	if err != nil {
        f,_ = os.OpenFile(path,os.O_RDWR,os.ModePerm)
		defer f.Close()
        info,_ := f.Stat()
        md.header = append(md.header,info.Size())
		// Get total elements
		totalElements = getInt64FromFile(f,0)

		if totalElements == limitMusicFile {
			md.changeFolder = true
			md.Save()
			return
		}

		f.Seek(0,2)	// Back to the end
		// Position in header depend on number element
		headerPos += totalElements*8
    }else{
		defer f.Close()
        // Create header at begin
        f.Write(getInt64AsByte(int64(len(md.musics))))
        f.Write(make([]byte,limitMusicFile*8))
        md.header = append(md.header,8*(1+limitMusicFile))
    }
    md.currentRead = 0
    // Use a reader over md. Write header at the end
    io.Copy(f,md)
	// Write total elements
	f.WriteAt(getInt64AsByte(totalElements + int64(len(md.musics))),0)
    // Write header
    f.WriteAt(getInts64AsByte(md.header[:len(md.header)-1]),headerPos)
}

// Read used in copy to save data in file
func (md * MusicDictionnary)Read(tab []byte)(int,error){
    // Read md musics, evaluate if enough place in tab (int32 for length + len data
    nbWrite := 0
    for{
        // Check if all data at been read
        if md.currentRead >= len(md.musics){
            return nbWrite,io.EOF
        }
        data := md.musics[md.currentRead].toJSON()
        if size := 8 + len(data) ; nbWrite + size < len(tab) {
            writeBytes(tab,getInt64AsByte(int64(len(data))),nbWrite)
            writeBytes(tab,data,nbWrite+8)
            nbWrite+=size
            md.currentRead++
            // Save position in temp header
            // First case, init
            md.header = append(md.header,md.header[len(md.header)-1]+int64(size))
        }else{
            break
        }
    }

    return nbWrite,nil
}

// Add music in dictionnary. If file limit is reach, save the file
func (md * MusicDictionnary)Add(music id3.File){
    if md.currentSize() >= limitMusicFile {
        md.Save()
		md.changeFolder = true
        // Save, use new file
        md.musics = make([]Music,0,limitMusicFile)
		md.previousSize = 0
    }
    md.musics = append(md.musics,Music{music,md.nextId})
	md.nextId++
}

func (md MusicDictionnary)GetMusicFromId(id int)map[string]string{
	fileId := id / limitMusicFile

	path := filepath.Join(folder,fmt.Sprintf("music_%d.dico",fileId))
	if f,err := os.Open(path) ; err == nil {
		defer f.Close()
		pos := int64(id - fileId*limitMusicFile)*8
		posInFile := getInt64FromFile(f,pos)
		lengthData := getInt64FromFile(f,posInFile)

		data := make([]byte,lengthData)
		f.ReadAt(data,posInFile+8)

		var results map[string]string
		json.Unmarshal(data,&results)
		return results
	}
	return nil
}

// LoadDictionnary load the dictionnary which store music info by id
func LoadDictionnary()MusicDictionnary{
    md := MusicDictionnary{changeFolder:false}

    fileId,notExist := md.findLastFile()
	if notExist == nil{
		// Load the last file and get current element inside
		path := filepath.Join(folder,fmt.Sprintf("music_%d.dico",fileId))
		f,_ := os.Open(path)
		defer f.Close()
		tabNb := make([]byte,8)
		f.ReadAt(tabNb,0)
		md.previousSize = int(getInt64FromFile(f,0))
		md.nextId = fileId*limitMusicFile + int64(md.previousSize+1)
	}else{
        md.previousSize = 0
        md.musics = make([]Music,0,limitMusicFile)
		md.nextId = 1
    }
    //
    return md
}

// TODO get a real path
const (
	mp3InfoPath = "C:\\tmp\\zik\\mp3info.exe"
)

// extractInfo get id3tag info
func extractInfo(filename string)*id3.File{
    r,_ := os.Open(filename)
	music := id3.Read(r)

	cmd := exec.Command(mp3InfoPath,"-p","%S",filename)
	if result,error := cmd.Output() ; error == nil {
		music.Length = string(result)
	}

    return music
}

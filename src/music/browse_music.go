package music
import (
    "os"
    "logger"
    "path/filepath"
    "github.com/mjibson/id3"
    "encoding/json"
    "io"
)

/* Give methods to browse musics in a specific directory */


var folder = "D:\\Documents\\Projets\\data"

// Browse a folder to get all data
func Browse(folderName string){
    dictionnary := LoadDictionnary()

    if folder,err := os.Open(folderName) ; err == nil {
        defer folder.Close()
        // List all files
        files,_ := folder.Readdir(-1)
        logger.GetLogger().Info("=>",len(files))
        for _,file := range files {
            path := filepath.Join(folderName,file.Name())
            if file.IsDir() {
                logger.GetLogger().Info("Parse",path)
                Browse(path)
            } else{
                // TODO : check in bloomfilter if exist. If true, final check in dictionnary
                dictionnary.Add(*extractInfo(path))
            }
        }
        logger.GetLogger().Info(folderName,"INDEX SIZE => ",len(dictionnary.musics))
    }
}


const (
    limitMusicFile = 10000
)

type MusicDictionnary struct{
    previousSize int
    // Data in json format
    musics [][]byte
    // Used to store header to write. List of position
    header []int64
    //
    currentRead int
}

func (md MusicDictionnary)currentSize()int{
    return md.previousSize + len(md.musics)
}

// GetInt64AsByte return a byte array representation of int64
func GetInt64AsByte(n int64) []byte {
    return []byte{byte(n), byte(n >> 8), byte(n >> 16), byte(n >> 24),
        byte(n >> 32), byte(n >> 40), byte(n >> 48), byte(n >> 56),
    }
}

func writeBytes(to,from []byte,pos int){
    for i := 0 ; i < len(from) ; i++ {
        to[i+pos] = from[i]
    }
}

// Save the file. Header with fix size (fixe number of element, 10000 for example). Header link position of element.
// nb element (8) | pos 1 (8) | pos 2 (8) | ... | dataLength 1 (4) | data 1 (n) | ...
func (md MusicDictionnary)Save(){
    // Save in the file. Create data in buffer instead of write everytime
    path := filepath.Join(folder,"music_1.dico")
    f,err := os.OpenFile(path,os.O_CREATE|os.O_RDWR|os.O_EXCL,os.ModePerm)
    // If exist, just append result at the end
    md.header = make([]int64,0,len(md.musics))
    if err != nil {
        f,_ = os.OpenFile(path,os.O_APPEND|os.O_RDWR,os.ModePerm)
        info,_ := f.Stat()
        md.header = append(md.header,info.Size())
    }else{
        // Create header at begin
        f.Write(GetInt64AsByte(len(md.musics)))
        f.Write(make([]byte,limitMusicFile*8))
        md.header = append(md.header,8*(1+limitMusicFile))
    }
    md.currentRead = 0
    // Use a reader over md. Write header at the end
    io.Copy(f,md)
    //get header
    f.WriteAt()
    md.header[:len(md.header)-1]

    //no need to write the last
}

func (md * MusicDictionnary)Read(tab []byte)(int,error){
    // Read md musics, evaluate if enough place in tab (int32 for length + len data
    nbWrite := 0
    for{
        // Check if all data at been read
        if md.currentRead >= len(md.musics){
            return nbWrite,io.EOF
        }
        data := md.musics[md.currentRead]
        if size := 8 + len(data) ; nbWrite + size < len(tab) {
            writeBytes(tab,GetInt64AsByte(len(data)),nbWrite)
            writeBytes(tab,data,nbWrite+8)
            nbWrite+=size
            md.currentRead++
            // Save position in temp header
            // First case, init
            md.header = append(md.header,md.header[len(md.header)-1]+size)
        }else{
            break
        }
    }

    return nbWrite
}

// Add music in dictionnary. If file limit is reach, save the file
func (md * MusicDictionnary)Add(music id3.File){
    if md.currentSize() > limitMusicFile {
        md.Save()
        // Save, use new file
        md.musics = make([][]byte,0,limitMusicFile)
    }
    data,_ := json.Marshal(music)
    md.musics = append(md.musics,data)
}

// LoadDictionnary load the dictionnary which store music info by id
func LoadDictionnary()MusicDictionnary{
    md := MusicDictionnary{}
    // Load the last file
    if false == true{

    }else{
        md.previousSize = 0
        md.musics = make([][]byte,0,limitMusicFile)
    }
    //
    return md
}

// extractInfo get id3tag info
func extractInfo(filename string)*id3.File{
    r,_ := os.Open(filename)
    // TODO must get length from external solution
    return id3.Read(r)
}
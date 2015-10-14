package music
import (
    "encoding/gob"
    "os"
    "path/filepath"
    "io"
)

// Give methods to manage album


type AlbumsList struct{
    albums []string
}

// One file with link between a father element id and son element id (map)
type ElementsIndex struct{

}

// AlbumByArtist store all album for artist
type ElementsByFather map[int][]int

func (ebf * ElementsByFather)Add(fatherId,sonId int){
    if albums,ok := (*ebf)[fatherId] ; ok {
        (*ebf)[fatherId] = append(albums,sonId)
    }else{
        (*ebf)[fatherId] = []int{sonId}
    }
}

func (ebf * ElementsByFather)Adds(fatherId int,sonsId []int){
    if list,present := (*ebf)[fatherId] ; present {
        (*ebf)[fatherId] = append(list,sonsId...)
    }else{
        (*ebf)[fatherId] = sonsId
    }
}

func (ebf ElementsByFather)Save(folder string){
    path := filepath.Join(folder,"artist_music.index")
    f,_ := os.OpenFile(path,os.O_TRUNC|os.O_CREATE|os.O_RDWR,os.ModePerm)
    defer f.Close()
    enc := gob.NewEncoder(f)
    enc.Encode(ebf)
}

func LoadElementsByFather(folder, filename string)ElementsByFather{
    path := filepath.Join(folder,filename + ".index")
    ebf := ElementsByFather{}
    if f,err := os.Open(path);err == nil {
        dec := gob.NewDecoder(f)
        dec.Decode(&ebf)
        f.Close()
    }else{
        ebf = ElementsByFather(make(map[int][]int))
    }
    return ebf
}

// Album have a list of music. Id album => ids music
type MusicByAlbum struct {
    albums [][]int
    // used to save data
    currentWriteId int
    // Store all positions
    header []int64
}

// return id of album. Id start at one
func (mba * MusicByAlbum)Adds(musicsId []int)int{
    id := len(mba.albums)+1
    mba.albums = append(mba.albums,musicsId)
    return id
}


// The file is trunced at each time (full save)
func (mba MusicByAlbum)Save(folder string){
    path := filepath.Join(folder,"album_music.index")
    f,_ := os.OpenFile(path,os.O_CREATE|os.O_TRUNC,os.ModePerm)
    // Reserve header size (nb elements * 8 + 4)
    f.Write(make([]byte,len(mba.albums)+4))
    mba.header = make([]int64,len(mba.albums))
    io.Copy(f,&mba)
    // Rewrite header at the beginning
}

// Struct : nb (4) pos 1 (8) | pos 2 (8) | ... | nb music 1 (2) | musics list (list of 4) | ...
// Just copy data and save position
func (mba * MusicByAlbum)Read(p []byte)(int,error){
    lengthData := 0
    for {
        if mba.currentWriteId >= len(mba.albums){
            return lengthData,io.EOF
        }
        // Check if enough place to length
        // Check enougth place to write data nb element (2o)
        if len(p) < lengthData + 2 {
            return lengthData,nil
        }
        album := mba.albums[mba.currentWriteId]
        writeBytes(p,getInt16AsByte(int16(len(album))),lengthData)
        lengthData+=2
        if mba.header[mba.currentWriteId] == 0{
            if mba.currentWriteId == 0 {
                // first position is just after the header
                mba.header[mba.currentWriteId] = int64(4 + 8*len(mba.albums))
            }else{
                // Take last position and add last length data
                mba.header[mba.currentWriteId] = mba.header[mba.currentWriteId-1] + int64(len(mba.albums[mba.currentWriteId-1])*4 + 2)
            }
        }

        // Write in header only if header is empty (cause partial write could append)
        // Check enough place to write musics. If not, check number of music which can be written
        nbWritable := (len(p) - lengthData)/4
        if len(album)>nbWritable {
            // Partial write, just some musics
            data := getInts32AsByte(album[:nbWritable])
            writeBytes(p,data,lengthData)
            mba.albums[mba.currentWriteId] = album[nbWritable:]
            lengthData+=len(data)
        }else{
            // write all music
            data := getInts32AsByte(album)
            writeBytes(p,data,lengthData)
            mba.currentWriteId++
            lengthData+=len(data)
        }
    }
    return lengthData,nil

}
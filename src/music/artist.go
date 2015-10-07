package music

import (
	"io"
	"os"
	"path/filepath"
	"encoding/binary"
)


type ArtistIndex struct{
	// Used to define if an artist exist (id of artist)
	artists map[string]int
	// Used in write
	tempBuffer []byte
	currentId int
	// new artists
	artistsToSave []string
	currentSave int
}

// Get artist index to search...
func LoadArtistIndex(folder string)ArtistIndex{
	path := filepath.Join(folder,"artist.dico")
	f,err := os.Open(path)
	ai := ArtistIndex{artists:make(map[string]int),currentId:1,artistsToSave:make([]string,0)}
	if err == nil {
		io.Copy(&ai,f)
	}
	return ai
}

// Add the artist in index. Return id
func (ai * ArtistIndex)Add(artist string)int{
	// Check if exist
	if id,exist := ai.artists[artist] ; exist {
		return id
	}
	id := ai.currentId
	ai.artists[artist] = id
	ai.artistsToSave = append(ai.artistsToSave,artist)
	ai.currentId++
	return id
}

// FindAll return all artists with id
func (ai ArtistIndex)FindAll()map[string]int {
	return ai.artists
}

// Save only new artists
func (ai * ArtistIndex)Save(folder string){
	path := filepath.Join(folder,"artist.dico")
	f,err := os.OpenFile(path,os.O_CREATE|os.O_EXCL|os.O_RDWR,os.ModePerm)
	if err == nil {
		// New, write size
		f.Write(getInt32AsByte(int32(len(ai.artists))))
	}else{
		//
		f.WriteAt(getInt32AsByte(int32(len(ai.artists))),0)
		f.Seek(0,2)
	}
	ai.currentSave = 0
	io.Copy(f,ai)
	f.Close()
}

func (ai * ArtistIndex)Read(p []byte)(int,error){
	pos := 0
	for {
		if ai.currentSave >= len(ai.artistsToSave){
			return pos,io.EOF
		}
		artist := ai.artistsToSave[ai.currentSave]
		if pos + 2 + len(artist) > len(p){
			return pos,nil
		}
		writeBytes(p,getInt16AsByte(int16(len(artist))),pos)
		writeBytes(p,[]byte(artist),pos+2)
		pos+=2+len(artist)
		ai.currentSave++
	}
}

// Write get data in p and write in object
// nb artist (4) | lenght name (2) | data name...
func (ai * ArtistIndex)Write(p []byte)(int,error){
	pos := 0
	if ai.artists == nil || len(ai.artists) == 0{
		// Load number, first 4 bytes
		ai.artists = make(map[string]int,int(binary.LittleEndian.Uint32(p[0:4])))
		ai.currentId = 1
		pos=4
	}
	pSize := len(p)
	if ai.tempBuffer != nil && len(ai.tempBuffer) > 0{
		p = append(ai.tempBuffer,p...)
		ai.tempBuffer = nil
	}
	for {
		if pos + 2 > len(p) {
			// Save rest in buffer
			ai.tempBuffer = p[pos:]
			return pSize,nil
		}
		artistSize := int(binary.LittleEndian.Uint16(p[pos:pos+2]))
		if pos + 2 + artistSize > len(p)   {
			ai.tempBuffer = p[pos:]
			return pSize,nil
		}
		ai.artists[string(p[pos+2:pos+2+artistSize])] = ai.currentId
		ai.currentId++
		pos+=2+artistSize
	}
	return pSize,nil
}
